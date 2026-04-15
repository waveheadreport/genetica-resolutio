package backend

import (
	"context"
	"net/http"
	"time"
)

// SourceUpdateInfo reports whether a remote source appears newer than the local copy.
type SourceUpdateInfo struct {
	SourceID       string `json:"sourceID"`
	RemoteModified string `json:"remoteModified"`
	LocalSaved     string `json:"localSaved"`
	UpdateAvail    bool   `json:"updateAvail"`
	Checked        bool   `json:"checked"` // false if HEAD failed
	Message        string `json:"message,omitempty"`
}

// CheckSourceUpdate issues a HEAD request for a source URL and compares Last-Modified
// against the locally-stored DownloadedAt timestamp.
func CheckSourceUpdate(ctx context.Context, sourceID, url, localSaved string) SourceUpdateInfo {
	info := SourceUpdateInfo{SourceID: sourceID, LocalSaved: localSaved}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		info.Message = err.Error()
		return info
	}
	resp, err := client.Do(req)
	if err != nil {
		info.Message = err.Error()
		return info
	}
	defer resp.Body.Close()
	info.Checked = true
	lm := resp.Header.Get("Last-Modified")
	info.RemoteModified = lm
	if lm == "" {
		info.Message = "remote has no Last-Modified header"
		return info
	}
	remote, err := http.ParseTime(lm)
	if err != nil {
		info.Message = "unparseable Last-Modified"
		return info
	}
	if localSaved == "" {
		info.UpdateAvail = true // never downloaded but somehow listed
		return info
	}
	// localSaved format: "2006-01-02 15:04:05"
	local, err := time.Parse("2006-01-02 15:04:05", localSaved)
	if err != nil {
		return info
	}
	if remote.After(local.Add(24 * time.Hour)) {
		info.UpdateAvail = true
	}
	return info
}
