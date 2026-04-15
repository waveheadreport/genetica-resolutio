package backend

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadProgress carries real-time download state to the caller.
type DownloadProgress struct {
	BytesDownloaded int64
	TotalBytes      int64   // -1 if unknown
	Percent         float64 // 0–100, or -1 if indeterminate
	Phase           string  // "downloading", "decompressing", "importing", "done"
	Message         string
}

// ProgressFn is called periodically during a download/import operation.
type ProgressFn func(DownloadProgress)

// DownloadToTemp downloads a URL to a temporary file, calling fn with progress.
// Returns the path to the temp file; caller is responsible for deleting it.
func DownloadToTemp(ctx context.Context, url string, fn ProgressFn) (string, error) {
	client := &http.Client{Timeout: 0} // no timeout — large files

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("User-Agent", "GeneticaResolutio/1.0 (desktop; +https://github.com)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	total := resp.ContentLength // may be -1

	tmp, err := os.CreateTemp("", "gr-download-*")
	if err != nil {
		return "", fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	var downloaded int64
	buf := make([]byte, 256*1024) // 256 KB chunks
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	reportProgress := func() {
		pct := float64(-1)
		if total > 0 {
			pct = float64(downloaded) / float64(total) * 100
		}
		fn(DownloadProgress{
			BytesDownloaded: downloaded,
			TotalBytes:      total,
			Percent:         pct,
			Phase:           "downloading",
			Message:         fmt.Sprintf("Downloaded %s", formatBytes(downloaded)),
		})
	}

	for {
		select {
		case <-ctx.Done():
			tmp.Close()
			os.Remove(tmpPath)
			return "", ctx.Err()
		case <-ticker.C:
			reportProgress()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := tmp.Write(buf[:n]); werr != nil {
				tmp.Close()
				os.Remove(tmpPath)
				return "", werr
			}
			downloaded += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return "", fmt.Errorf("read error: %w", err)
		}
	}

	tmp.Close()
	reportProgress()
	return tmpPath, nil
}

// OpenMaybeGzip opens a file, transparently decompressing if it is gzip.
func OpenMaybeGzip(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Peek at first two bytes for gzip magic number
	magic := make([]byte, 2)
	if _, err := f.Read(magic); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, err
	}
	if magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		return &gzipCloser{gz: gz, f: f}, nil
	}
	return f, nil
}

type gzipCloser struct {
	gz *gzip.Reader
	f  *os.File
}

func (g *gzipCloser) Read(p []byte) (int, error) { return g.gz.Read(p) }
func (g *gzipCloser) Close() error {
	g.gz.Close()
	return g.f.Close()
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
