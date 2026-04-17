package backend

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// KV is the package-level bbolt handle, opened once at startup.
//
// Layout:
//   Bucket "meta"   : key = source name (e.g. "gwas")          → JSON(SourceMeta)
//   Bucket "<src>"  : key = rsid + 0x00 + 8-byte big-endian seq → gob(SNPRecord)
//
// The NUL separator + monotonic sequence lets us store multiple SNPRecord values
// per rsID while keeping lookups to a single prefix-scan with bbolt's cursor.
// bbolt's lexicographic key ordering ensures all records for one rsID are
// contiguous, and the NUL byte guarantees no false-positive prefix matches
// (e.g., rs123 never matches rs1234 since the byte after "rs123" must be 0x00).
var KV *bolt.DB

const metaBucket = "meta"
const posBucket = "pos"

// DBPath returns the filesystem path for the local reference database.
func DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "genetica-resolutio")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "snpdb.bolt"), nil
}

// OpenSQLiteDB opens the local reference database. The name is historical —
// the storage engine is bbolt, not SQLite; the public function name is kept
// so existing callers don't need to change.
func OpenSQLiteDB() error {
	path, err := DBPath()
	if err != nil {
		return fmt.Errorf("could not determine DB path: %w", err)
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return fmt.Errorf("could not open bbolt DB: %w", err)
	}
	KV = db
	return initSchema()
}

// CloseKV closes the database handle. Safe to call at shutdown.
func CloseKV() error {
	if KV == nil {
		return nil
	}
	err := KV.Close()
	KV = nil
	return err
}

func initSchema() error {
	return KV.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(metaBucket)); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists([]byte(posBucket))
		return err
	})
}

// makeRecordKey builds a lexicographically-sorted key: rsid + NUL + big-endian seq.
func makeRecordKey(rsid string, seq uint64) []byte {
	key := make([]byte, 0, len(rsid)+1+8)
	key = append(key, rsid...)
	key = append(key, 0)
	var s [8]byte
	binary.BigEndian.PutUint64(s[:], seq)
	return append(key, s[:]...)
}

// rsidPrefix returns the lookup prefix for a given rsid: rsid + NUL.
func rsidPrefix(rsid string) []byte {
	p := make([]byte, 0, len(rsid)+1)
	p = append(p, rsid...)
	p = append(p, 0)
	return p
}

// QuerySNPsByRSID returns all records across every installed source for an rsid.
func QuerySNPsByRSID(rsid string) ([]SNPRecord, error) {
	if KV == nil {
		return nil, nil
	}
	prefix := rsidPrefix(strings.ToLower(rsid))
	var out []SNPRecord

	err := KV.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if bytes.Equal(name, []byte(metaBucket)) {
				return nil
			}
			c := b.Cursor()
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				var rec SNPRecord
				if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&rec); err == nil {
					out = append(out, rec)
				}
			}
			return nil
		})
	})
	return out, err
}

// PosEntry maps a genomic position to an rsID for one reference build.
type PosEntry struct {
	Build string
	Chrom string
	Pos   string
	RSID  string
}

func makePosKey(build, chrom, pos string) []byte {
	return []byte(build + ":" + chrom + ":" + pos)
}

// BulkInsertPositionalIndex writes chr:pos → rsID mappings into the positional
// index bucket. These are additive — a given genomic position always maps to the
// same rsID regardless of which database reported it.
func BulkInsertPositionalIndex(entries []PosEntry) error {
	if KV == nil {
		return fmt.Errorf("database not open")
	}
	if len(entries) == 0 {
		return nil
	}
	return KV.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(posBucket))
		if b == nil {
			return fmt.Errorf("pos bucket missing")
		}
		for _, e := range entries {
			key := makePosKey(e.Build, e.Chrom, e.Pos)
			if err := b.Put(key, []byte(e.RSID)); err != nil {
				return err
			}
		}
		return nil
	})
}

// QueryRSIDByPosition looks up an rsID from the positional index.
// If build is empty, tries GRCh38 then GRCh37.
func QueryRSIDByPosition(chrom, pos, build string) (string, error) {
	if KV == nil {
		return "", nil
	}
	chrom = strings.TrimPrefix(strings.TrimPrefix(chrom, "chr"), "Chr")
	builds := []string{build}
	if build == "" {
		builds = []string{"GRCh38", "GRCh37"}
	}
	var rsid string
	err := KV.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(posBucket))
		if b == nil {
			return nil
		}
		for _, bld := range builds {
			key := makePosKey(bld, chrom, pos)
			if v := b.Get(key); v != nil {
				rsid = string(v)
				return nil
			}
		}
		return nil
	})
	return rsid, err
}

// BulkInsertSNPs writes a batch of records into the given source's bucket.
// Multiple records for the same rsid are preserved (each gets a unique sequence).
func BulkInsertSNPs(records []SNPRecord, source string) error {
	if KV == nil {
		return fmt.Errorf("database not open")
	}
	if len(records) == 0 {
		return nil
	}
	return KV.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(source))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		for _, r := range records {
			seq, err := b.NextSequence()
			if err != nil {
				return err
			}
			buf.Reset()
			if err := gob.NewEncoder(&buf).Encode(r); err != nil {
				return err
			}
			key := makeRecordKey(strings.ToLower(r.RSID), seq)
			// Copy the encoded value — bbolt requires the slice to remain
			// valid until the transaction commits, and buf is reused in
			// the next iteration.
			val := make([]byte, buf.Len())
			copy(val, buf.Bytes())
			if err := b.Put(key, val); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpsertSourceMeta records that a source was imported (or re-imported).
func UpsertSourceMeta(name, displayName, version string, rowCount int) error {
	if KV == nil {
		return fmt.Errorf("database not open")
	}
	m := SourceMeta{
		Name:         name,
		DisplayName:  displayName,
		RowCount:     rowCount,
		DownloadedAt: time.Now().UTC().Format("2006-01-02 15:04:05"),
		Version:      version,
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return KV.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		if b == nil {
			return fmt.Errorf("meta bucket missing")
		}
		return b.Put([]byte(name), data)
	})
}

// DeleteSource removes all records for a source and its metadata.
func DeleteSource(name string) error {
	if KV == nil {
		return fmt.Errorf("database not open")
	}
	return KV.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(name)) != nil {
			if err := tx.DeleteBucket([]byte(name)); err != nil {
				return err
			}
		}
		if meta := tx.Bucket([]byte(metaBucket)); meta != nil {
			return meta.Delete([]byte(name))
		}
		return nil
	})
}

// SourceMeta holds metadata for one installed source.
type SourceMeta struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	RowCount     int    `json:"rowCount"`
	DownloadedAt string `json:"downloadedAt"`
	Version      string `json:"version"`
}

// InstalledSources returns metadata for all sources that have been imported.
func InstalledSources() ([]SourceMeta, error) {
	if KV == nil {
		return nil, nil
	}
	var out []SourceMeta
	err := KV.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metaBucket))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var m SourceMeta
			if err := json.Unmarshal(v, &m); err == nil {
				out = append(out, m)
			}
			return nil
		})
	})
	return out, err
}

// SQLiteDBStats returns per-source record counts. The name is historical; data
// is stored in bbolt, not SQLite.
func SQLiteDBStats() (map[string]int, error) {
	counts := make(map[string]int)
	if KV == nil {
		return counts, nil
	}
	err := KV.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if bytes.Equal(name, []byte(metaBucket)) {
				return nil
			}
			counts[string(name)] = b.Stats().KeyN
			return nil
		})
	})
	return counts, err
}
