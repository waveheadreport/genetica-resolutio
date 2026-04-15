package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"genetica-resolutio-desktop/backend"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are bound to the frontend.
type App struct {
	ctx context.Context

	// download cancellation: one cancel func per active source download
	dlMu      sync.Mutex
	dlCancels map[string]context.CancelFunc
}

func NewApp() *App {
	return &App{
		dlCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	backend.InitDB()
	if err := backend.OpenSQLiteDB(); err != nil {
		// Non-fatal: log and continue with curated-only mode
		fmt.Fprintf(os.Stderr, "SQLite init warning: %v\n", err)
	}
}

// ── FILE HANDLING ─────────────────────────────────────────────────────────────

func (a *App) OpenFileDialog() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select your raw DNA file",
		Filters: []runtime.FileFilter{
			{DisplayName: "DNA Files (*.txt, *.csv, *.tsv, *.gz)", Pattern: "*.txt;*.csv;*.tsv;*.gz"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) GetFileInfo(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	return map[string]interface{}{
		"name":    filepath.Base(path),
		"sizeMB":  fmt.Sprintf("%.2f", sizeMB),
		"path":    path,
		"modTime": info.ModTime().Format(time.RFC3339),
	}, nil
}

// ── ANALYSIS ──────────────────────────────────────────────────────────────────

func (a *App) AnalyzeFile(path string) (*backend.AnalysisResult, error) {
	emit := func(pct int, label, text string) {
		runtime.EventsEmit(a.ctx, "analysis:progress", map[string]interface{}{
			"pct": pct, "label": label, "text": text,
		})
		time.Sleep(30 * time.Millisecond)
	}

	emit(5, "Opening", "Reading DNA file from disk…")
	parsed, err := backend.ParseDNAFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	emit(35, "Indexing", fmt.Sprintf("Parsed %s — %d SNPs loaded", parsed.Provider, parsed.TotalSNPs))
	time.Sleep(60 * time.Millisecond)

	emit(55, "Matching", "Cross-referencing variant databases…")
	time.Sleep(50 * time.Millisecond)

	emit(70, "Interpreting", "Analyzing genotype effects…")
	result := backend.RunAnalysis(parsed)

	emit(88, "Building", "Constructing action plan…")
	time.Sleep(50 * time.Millisecond)

	emit(96, "Rendering", "Finalizing report…")
	time.Sleep(40 * time.Millisecond)

	emit(100, "Complete", "Analysis complete")
	return result, nil
}

// ── DATABASE STATS ────────────────────────────────────────────────────────────

func (a *App) GetDBStats() backend.DBStats {
	return backend.GetDBStats()
}

// ── DATABASE SOURCES CATALOG ─────────────────────────────────────────────────

// DatabaseSource describes one available reference database.
type DatabaseSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Coverage    string `json:"coverage"`
	FileSize    string `json:"fileSize"`
	Variants    string `json:"variants"`
	URL         string `json:"url"`
	FileExt     string `json:"fileExt"` // for import routing
	Warning     string `json:"warning"` // optional caution note
	// populated dynamically
	Installed    bool   `json:"installed"`
	RowCount     int    `json:"rowCount"`
	DownloadedAt string `json:"downloadedAt"`
	Downloading  bool   `json:"downloading"`
}

var dbSources = []DatabaseSource{
	{
		ID:          "gwas",
		Name:        "GWAS Catalog",
		Description: "The NHGRI-EBI Genome-Wide Association Studies catalog — the most comprehensive collection of published SNP-trait associations from peer-reviewed GWAS studies worldwide.",
		Coverage:    "Population-level trait associations: disease risk, body measurements, biomarkers, cognitive traits, behavioral phenotypes. Every entry is backed by a published study.",
		FileSize:    "~90 MB (zipped)",
		Variants:    "~700,000 associations",
		URL:         "https://ftp.ebi.ac.uk/pub/databases/gwas/releases/latest/gwas-catalog-associations_ontology-annotated-full.zip",
		FileExt:     ".zip",
	},
	{
		ID:          "clinvar",
		Name:        "ClinVar",
		Description: "NCBI's database of clinically-interpreted human genetic variants — submitted by laboratories, clinics, and researchers worldwide. The gold standard for clinical variant significance.",
		Coverage:    "Clinical significance of variants: pathogenic, likely pathogenic, risk factor, protective. Covers Mendelian diseases, cancer predisposition, pharmacogenomics, and rare conditions.",
		FileSize:    "~100 MB (gzipped)",
		Variants:    "~2,800,000 variants",
		URL:         "https://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh38/clinvar.vcf.gz",
		FileExt:     ".vcf.gz",
	},
	{
		ID:          "pharmgkb",
		Name:        "PharmGKB",
		Description: "The Pharmacogenomics Knowledge Base — curated drug-gene interactions, clinical annotations, and dosing guidelines backed by evidence from peer-reviewed pharmacogenomics literature.",
		Coverage:    "Drug response and metabolism: which variants affect how your body processes specific medications. Essential before starting new prescriptions.",
		FileSize:    "~5–20 MB",
		Variants:    "~5,000 drug-variant annotations",
		URL:         "https://api.pharmgkb.org/v1/download/file/data/clinicalAnnotations.zip",
		FileExt:     ".zip",
	},
	{
		ID:          "dbsnp",
		Name:        "dbSNP (Clinical Subset)",
		Description: "NCBI's reference SNP database — the world's largest repository of human genetic variation. We import only the clinically-annotated subset (variants with ClinVar significance), not the full ~1 billion variant catalog.",
		Coverage:    "Reference variant identifiers with clinical significance annotations. Broadens coverage of known clinically-relevant SNPs beyond ClinVar alone.",
		FileSize:    "~50 GB (full) — ~2–5 GB (clinical subset only imported)",
		Variants:    "~500,000 clinically-annotated variants extracted",
		URL:         "https://ftp.ncbi.nlm.nih.gov/snp/latest_release/VCF/GCF_000001405.40.gz",
		FileExt:     ".vcf.gz",
		Warning:     "This file is very large (~50 GB download). Only variants with clinical annotations are imported; the rest are discarded. Requires significant disk space and time.",
	},
}

// GetDatabaseSources returns the catalog of available databases with live install status.
func (a *App) GetDatabaseSources() ([]DatabaseSource, error) {
	installed, err := backend.InstalledSources()
	if err != nil {
		return dbSources, nil // return catalog even if DB query fails
	}

	instMap := make(map[string]backend.SourceMeta)
	for _, m := range installed {
		instMap[m.Name] = m
	}

	a.dlMu.Lock()
	defer a.dlMu.Unlock()

	result := make([]DatabaseSource, len(dbSources))
	copy(result, dbSources)
	for i, src := range result {
		if m, ok := instMap[src.ID]; ok {
			result[i].Installed = true
			result[i].RowCount = m.RowCount
			result[i].DownloadedAt = m.DownloadedAt
		}
		_, result[i].Downloading = a.dlCancels[src.ID]
	}
	return result, nil
}

// ── DOWNLOAD / IMPORT ─────────────────────────────────────────────────────────

// DownloadDatabase downloads and imports one of the reference databases.
// It runs in a goroutine and emits "db:progress" and "db:done" events.
func (a *App) DownloadDatabase(sourceID string) error {
	var src *DatabaseSource
	for i := range dbSources {
		if dbSources[i].ID == sourceID {
			src = &dbSources[i]
			break
		}
	}
	if src == nil {
		return fmt.Errorf("unknown source: %s", sourceID)
	}

	a.dlMu.Lock()
	if _, already := a.dlCancels[sourceID]; already {
		a.dlMu.Unlock()
		return fmt.Errorf("download already in progress for %s", sourceID)
	}
	dlCtx, cancel := context.WithCancel(a.ctx)
	a.dlCancels[sourceID] = cancel
	a.dlMu.Unlock()

	emit := func(phase, msg string, pct float64) {
		runtime.EventsEmit(a.ctx, "db:progress", map[string]interface{}{
			"sourceID": sourceID,
			"phase":    phase,
			"message":  msg,
			"pct":      pct,
		})
	}

	go func() {
		defer func() {
			a.dlMu.Lock()
			delete(a.dlCancels, sourceID)
			a.dlMu.Unlock()
		}()

		// Remove any existing data for this source first
		_ = backend.DeleteSource(sourceID)

		emit("downloading", fmt.Sprintf("Starting download of %s…", src.Name), 0)

		fn := func(p backend.DownloadProgress) {
			pct := p.Percent
			if pct < 0 {
				pct = -1
			}
			emit("downloading", p.Message, pct*0.6) // downloading = 0–60%
		}

		tmpPath, err := backend.DownloadToTemp(dlCtx, src.URL, fn)
		if err != nil {
			runtime.EventsEmit(a.ctx, "db:done", map[string]interface{}{
				"sourceID": sourceID, "error": err.Error(),
			})
			return
		}
		defer os.Remove(tmpPath)

		emit("importing", fmt.Sprintf("Importing %s into local database…", src.Name), 60)

		importFn := func(p backend.DownloadProgress) {
			emit("importing", p.Message, 60+p.Percent*0.38) // importing = 60–98%
		}

		var count int
		switch sourceID {
		case "gwas":
			count, err = backend.ImportGWASCatalog(dlCtx, tmpPath, importFn)
		case "clinvar":
			count, err = backend.ImportClinVar(dlCtx, tmpPath, importFn)
		case "pharmgkb":
			count, err = backend.ImportPharmGKB(dlCtx, tmpPath, importFn)
		case "dbsnp":
			count, err = backend.ImportdbSNP(dlCtx, tmpPath, importFn)
		default:
			err = fmt.Errorf("no importer for source: %s", sourceID)
		}

		if err != nil {
			runtime.EventsEmit(a.ctx, "db:done", map[string]interface{}{
				"sourceID": sourceID, "error": err.Error(),
			})
			return
		}

		_ = backend.UpsertSourceMeta(sourceID, src.Name, time.Now().Format("2006-01-02"), count)

		emit("done", fmt.Sprintf("Imported %s (%d records)", src.Name, count), 100)
		runtime.EventsEmit(a.ctx, "db:done", map[string]interface{}{
			"sourceID": sourceID,
			"count":    count,
			"error":    "",
		})
	}()

	return nil
}

// CancelDownload cancels an in-progress download for the given source.
func (a *App) CancelDownload(sourceID string) {
	a.dlMu.Lock()
	defer a.dlMu.Unlock()
	if cancel, ok := a.dlCancels[sourceID]; ok {
		cancel()
	}
}

// DeleteDatabase removes all records for a source from the local database.
func (a *App) DeleteDatabase(sourceID string) error {
	return backend.DeleteSource(sourceID)
}

// ── RSID LOOKUP ──────────────────────────────────────────────────────────────

// LookupRSID returns all annotations across installed databases for a given rsID.
func (a *App) LookupRSID(rsid string) ([]backend.SNPRecord, error) {
	return backend.QuerySNPsByRSID(rsid)
}

// ── SESSIONS ──────────────────────────────────────────────────────────────────

func (a *App) SaveSession(result backend.AnalysisResult, filename string) (string, error) {
	return backend.SaveAnalysisSession(&result, filename)
}

func (a *App) ListSessions() ([]backend.SessionMeta, error) {
	return backend.ListSessionsOnDisk()
}

func (a *App) LoadSession(id string) (*backend.AnalysisResult, error) {
	return backend.LoadSessionByID(id)
}

func (a *App) DeleteSession(id string) error {
	return backend.DeleteSessionByID(id)
}

// ── DATABASE UPDATE CHECKER ──────────────────────────────────────────────────

// CheckDatabaseUpdates issues concurrent HEAD requests against every installed
// source's URL and returns whether each is older than the remote copy.
func (a *App) CheckDatabaseUpdates() (map[string]backend.SourceUpdateInfo, error) {
	installed, err := backend.InstalledSources()
	if err != nil {
		return nil, err
	}

	urlByID := make(map[string]string)
	for _, s := range dbSources {
		urlByID[s.ID] = s.URL
	}

	out := make(map[string]backend.SourceUpdateInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, m := range installed {
		url, ok := urlByID[m.Name]
		if !ok {
			continue
		}
		wg.Add(1)
		go func(id, u, saved string) {
			defer wg.Done()
			info := backend.CheckSourceUpdate(a.ctx, id, u, saved)
			mu.Lock()
			out[id] = info
			mu.Unlock()
		}(m.Name, url, m.DownloadedAt)
	}
	wg.Wait()
	return out, nil
}

// ── FILE COMPARISON ──────────────────────────────────────────────────────────

func (a *App) CompareFiles(pathA, pathB string) (*backend.ComparisonResult, error) {
	return backend.CompareTwoFiles(pathA, pathB)
}

// ── EXPORT ────────────────────────────────────────────────────────────────────

func (a *App) SaveReport(content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Genetic Report",
		DefaultFilename: fmt.Sprintf("genetic-report-%s.txt", time.Now().Format("2006-01-02")),
		Filters: []runtime.FileFilter{
			{DisplayName: "Text File (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}
	return path, nil
}
