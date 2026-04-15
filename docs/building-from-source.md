# Building from Source

This guide is for developers who want to compile Genetica Resolutio themselves, review the code, or modify it.

---

## Prerequisites

You need the following installed on your build machine:

| Tool | Version | Notes |
|---|---|---|
| Go | 1.21 or later | [golang.org/dl](https://golang.org/dl/) |
| Node.js | 18 or later | For building the frontend |
| npm | 9 or later | Bundled with Node.js |
| Wails CLI | v2.9 or later | Install via `go install` (see below) |
| C compiler | Any | `gcc` on Linux/Windows; Xcode CLT on macOS |

### Platform-specific dependencies

**Linux**

WebKitGTK 4.1 development headers are required:

```bash
# Fedora / RHEL
sudo dnf install webkit2gtk4.1-devel gcc

# Ubuntu / Debian
sudo apt install libwebkit2gtk-4.1-dev gcc build-essential

# Arch
sudo pacman -S webkit2gtk-4.1 base-devel
```

**macOS**

Install Xcode Command Line Tools if not already present:
```bash
xcode-select --install
```

**Windows**

Install [MinGW-w64](https://www.mingw-w64.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/). WebView2 is pre-installed on Windows 10/11.

---

## Install the Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify:
```bash
wails version
```

Make sure `$GOPATH/bin` (typically `~/go/bin`) is in your `PATH`. If `wails: command not found`, add it:

```bash
export PATH=$PATH:$HOME/go/bin
```

Add this line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

---

## Clone / obtain the source

```bash
git clone <repository-url> genetica-resolutio-desktop
cd genetica-resolutio-desktop
```

---

## Build

### Linux

```bash
wails build -tags webkit2_41
```

The `-tags webkit2_41` flag is required on Linux to use the WebKitGTK 4.1 API. Without it, the build will fail with a `pkg-config: webkit2gtk-4.0 not found` error on systems that have 4.1 installed.

### macOS

```bash
wails build
```

For Apple Silicon (M-series):
```bash
wails build -platform darwin/arm64
```

For a universal binary (runs on both Intel and Apple Silicon):
```bash
wails build -platform darwin/universal
```

### Windows

```bash
wails build
```

To build a Windows binary from Linux (cross-compilation):
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc wails build -platform windows/amd64
```

Cross-compiling for Windows requires `mingw-w64` installed on the Linux host.

---

## Output

The compiled binary is placed in `build/bin/`:

| Platform | Output file |
|---|---|
| Linux | `build/bin/genetica-resolutio` |
| macOS | `build/bin/genetica-resolutio.app` |
| Windows | `build/bin/genetica-resolutio.exe` |

The binary is self-contained. Copy it anywhere and run it.

---

## Development mode

To run a live-reloading development server:

```bash
wails dev -tags webkit2_41   # Linux
wails dev                     # macOS / Windows
```

This opens the app in a window and watches both Go and frontend source files. Changes to Go files trigger a backend recompile; changes to frontend files hot-reload in the window without restarting.

---

## Project structure

```
genetica-resolutio-desktop/
├── main.go              # Wails entry point; window configuration
├── app.go               # All Wails-bound methods exposed to the frontend
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── wails.json           # Wails project configuration
│
├── backend/
│   ├── types.go         # Shared data structures (Finding, ActionPlan, etc.)
│   ├── snpdata.go       # 357 curated SNP records compiled into the binary
│   ├── parser.go        # DNA file parser (all provider formats + VCF)
│   ├── analyzer.go      # Variant matching and genotype interpretation
│   ├── actionplan.go    # Action plan builder
│   ├── sqlitedb.go      # Local SQLite database (optional reference DBs)
│   ├── downloader.go    # HTTP download utility with progress reporting
│   └── importers.go     # GWAS Catalog, ClinVar, PharmGKB, dbSNP parsers
│
├── frontend/
│   ├── index.html       # HTML shell
│   ├── package.json     # Node dependencies (Vite)
│   ├── vite.config.js   # Vite configuration
│   └── src/
│       ├── main.js      # All UI logic (~750 lines)
│       └── style.css    # Complete design system
│
└── docs/
    └── *.md             # Documentation source files
```

---

## Key source files

### `backend/snpdata.go`

Contains the `initSNPDB()` function that seeds the in-memory curated variant database. Each variant is defined with:

```go
add(SNPRecord{
    RSID:        "rs1801133",
    Gene:        "MTHFR",
    Trait:       "Impaired Methylfolate Conversion",
    Category:    "nutrition",
    RiskAllele:  "A",
    Effect:      "Homozygous for A risk allele (A/A). MTHFR C677T reduces folate→5-MTHF conversion ~70% (hom). Raises homocysteine. Affects DNA methylation, cardiovascular risk, and mood.",
    Recommendation: "Use methylfolate (5-MTHF) not folic acid. Eat dark leafy greens. Take methylcobalamin B12. Test homocysteine annually — target <8 µmol/L.",
    Citation:    "10630380",
    Confidence:  "high",
})
```

To add a new curated variant, add a new `add(SNPRecord{...})` call in this function and rebuild.

### `backend/parser.go`

`ParseDNAFile(path string) (*ParseResult, error)` is the entry point. It handles gzip detection, provider identification, and format dispatching. The result is a `ParseResult` containing:
- `Provider` — detected provider name (string)
- `TotalSNPs` — count of valid SNPs parsed
- `SNPs` — `map[string][2]string` of rsID → [allele1, allele2]

### `backend/analyzer.go`

`RunAnalysis(parsed *ParseResult) *AnalysisResult` performs two-pass matching: curated map then SQLite. `interpretGenotype()` takes a SNPRecord and a [2]string allele pair and returns the status (high_risk, moderate, protective, normal), the badge label, and the CSS class.

### `app.go`

All methods on the `App` struct are automatically exposed to the frontend JavaScript by Wails. The JavaScript bindings are auto-generated into `frontend/wailsjs/go/` at build time.

---

## Running tests

```bash
go test ./backend/...
```

---

## Modifying the curated SNP database

The `snpdata.go` file can be regenerated from any source. If you have a CSV file of variants you want to add in bulk, a simple Go or Python script can generate the `add(SNPRecord{...})` calls.

The original 357 variants were ported from a web application's JavaScript source using:

```bash
grep "^addSNP(" source.html | python3 -c "
import sys, re
for line in sys.stdin:
    m = re.match(r'addSNP\((.+)\)$', line.strip())
    if m:
        args = [a.strip().strip('\"') for a in m.group(1).split(',', 8)]
        print(f'    add(SNPRecord{{RSID: \"{args[0]}\", Gene: \"{args[1]}\", ...')
"
```

---

## Wails documentation

The Wails v2 documentation at [wails.io/docs](https://wails.io/docs/introduction) covers the full framework API including additional window options, system tray integration, file associations, and code signing for distribution.
