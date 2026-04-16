# Genetica Resolutio

**Privacy-first desktop DNA analysis.** Upload your raw 23andMe, AncestryDNA, or MyHeritage file and get a science-backed health report — cross-referenced against curated variants and optional full public databases. Everything runs locally. Nothing ever leaves your machine.

---

## Features

- **Fully local analysis.** Your genome never touches a server. No telemetry, no account, no upload.
- **Curated SNP library** spanning nutrition, metabolism, fitness, sleep, disease risk, pharmacogenomics, and more — every variant backed by a PubMed citation.
- **Optional full reference databases** you can download on demand: GWAS Catalog, ClinVar, PharmGKB, and the clinical subset of dbSNP.
- **Per-finding notes** and **saved analysis sessions** so you can revisit a report later without re-parsing the file.
- **rsID lookup** for ad-hoc variant investigation across installed databases.
- **Two-file comparison** — diff two raw DNA files SNP-by-SNP (useful for family members or kit reprocessing).
- **Database update checker** flags when a reference source has a newer version upstream.
- **Report export** to plain text, CSV, or JSON.
- **Gender and ancestry-aware filtering** — sex-specific findings are hidden for the opposite sex; non-European ancestries get a calibration caveat banner.

## Install

Download the latest pre-built binary from the [Releases page](https://github.com/waveheadreport/genetica-resolutio/releases/latest):

| Platform | File |
|---|---|
| Linux (x86-64) | `genetica-resolutio-linux-amd64.tar.gz` |
| macOS (Universal — Intel + Apple Silicon) | `genetica-resolutio-macos-universal.zip` |
| Windows | `genetica-resolutio-windows-amd64.zip` |

Binaries are unsigned (this is a hobby project, not a $100/yr signed release). First-launch notes:

- **macOS:** Right-click the app → Open → Open. Or: System Settings → Privacy & Security → Open Anyway.
- **Windows:** SmartScreen will warn → More info → Run anyway.
- **Linux:** `chmod +x genetica-resolutio-linux-amd64 && ./genetica-resolutio-linux-amd64`.

## Privacy

This app's entire purpose is to keep your genetic data on your machine.

- No network calls are made during analysis.
- The only outbound traffic is when *you* explicitly click "Download" on a reference database — and those requests go directly to the official source (EBI, NCBI, PharmGKB), not through any third party.
- Saved sessions, notes, and settings live in `~/.local/share/genetica-resolutio/` (Linux) or the platform equivalent.

See [`docs/privacy.md`](docs/privacy.md) for the full breakdown.

## Medical disclaimer

This is **not a medical device** and does not provide a diagnosis. Findings are statistical tendencies drawn from published research, not certainties. Do not start, stop, or change medications based on this report. Share findings with a qualified physician, pharmacist, or genetic counselor before acting on them.

## Reference data sources

All citations and DOIs are visible inside the app's About dialog.

- **GWAS Catalog** — Sollis E, et al. *Nucleic Acids Res.* 2023;51(D1):D977–D985. [doi:10.1093/nar/gkac1010](https://doi.org/10.1093/nar/gkac1010)
- **ClinVar** — Landrum MJ, et al. *Nucleic Acids Res.* 2020;48(D1):D835–D844. [doi:10.1093/nar/gkz972](https://doi.org/10.1093/nar/gkz972)
- **PharmGKB** — Whirl-Carrillo M, et al. *Clin Pharmacol Ther.* 2021;110(3):563–572. [doi:10.1002/cpt.2350](https://doi.org/10.1002/cpt.2350)
- **dbSNP** — Sherry ST, et al. *Nucleic Acids Res.* 2001;29(1):308–311.

## Build from source

See [`docs/building-from-source.md`](docs/building-from-source.md). In short:

```bash
# Requires Go 1.21+, Node 20+, and the Wails v2 CLI.
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build                    # or: wails build -tags webkit2_41  (Linux)
```

The [`.github/workflows/release.yml`](.github/workflows/release.yml) workflow is the canonical build recipe — it's what produces the official releases.

## Documentation

- [Getting started](docs/getting-started.md)
- [Using the app](docs/using-the-app.md)
- [File formats](docs/file-formats.md)
- [Database manager](docs/database-manager.md)
- [Understanding your results](docs/understanding-results.md)
- [Privacy](docs/privacy.md)
- [FAQ](docs/faq.md)

## License

[MIT](LICENSE).
