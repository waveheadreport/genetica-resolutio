package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionMeta is the summary shown in the "recent analyses" list.
type SessionMeta struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Provider    string `json:"provider"`
	Matched     int    `json:"matched"`
	HighCount   int    `json:"highCount"`
	ModCount    int    `json:"modCount"`
	SavedAt     string `json:"savedAt"`
	GeneratedAt string `json:"generatedAt"`
}

// SavedSession is what actually lands on disk — meta + the full result.
type SavedSession struct {
	Meta   SessionMeta     `json:"meta"`
	Result *AnalysisResult `json:"result"`
}

func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "genetica-resolutio", "sessions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// SaveAnalysisSession writes the result to a JSON file and returns the session id.
func SaveAnalysisSession(r *AnalysisResult, filename string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil result")
	}
	dir, err := sessionsDir()
	if err != nil {
		return "", err
	}
	id := time.Now().UTC().Format("20060102-150405")
	meta := SessionMeta{
		ID:          id,
		Filename:    filepath.Base(filename),
		Provider:    r.Parsed.Provider,
		Matched:     r.Matched,
		HighCount:   len(r.Summary.High),
		ModCount:    len(r.Summary.Moderate),
		SavedAt:     time.Now().UTC().Format("2006-01-02 15:04"),
		GeneratedAt: r.GeneratedAt,
	}
	ss := SavedSession{Meta: meta, Result: r}
	data, err := json.Marshal(ss)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, id+".json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return id, nil
}

// ListSessionsOnDisk returns all saved sessions, newest first.
func ListSessionsOnDisk() ([]SessionMeta, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []SessionMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var ss SavedSession
		if err := json.Unmarshal(data, &ss); err != nil {
			continue
		}
		if ss.Meta.ID == "" {
			ss.Meta.ID = strings.TrimSuffix(e.Name(), ".json")
		}
		out = append(out, ss.Meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

// LoadSessionByID reads one session from disk.
func LoadSessionByID(id string) (*AnalysisResult, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ss SavedSession
	if err := json.Unmarshal(data, &ss); err != nil {
		return nil, err
	}
	if ss.Result == nil {
		return nil, fmt.Errorf("session has no result")
	}
	return ss.Result, nil
}

// DeleteSessionByID removes a session file.
func DeleteSessionByID(id string) error {
	dir, err := sessionsDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, id+".json"))
}
