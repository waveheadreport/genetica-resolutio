# Database Manager

The built-in curated database contains 357 hand-selected, evidence-reviewed variants. The Database Manager lets you expand that coverage significantly by downloading from four major public reference databases.

---

## Opening the Database Manager

From the home screen, click the **🗄 Databases** tab in the top navigation bar. This switches the right panel from the file upload interface to the database manager.

The Databases tab shows a table with one row per available source. Each row displays the source name, download size, variant count, data coverage description, current status, and an action button.

---

## Available databases

| Database | Download Size | Variants | Coverage |
|---|---|---|---|
| **GWAS Catalog** | ~200 MB | ~600,000 | Genome-wide association study findings: disease predispositions, traits, complex conditions |
| **ClinVar** | ~650 MB | ~200,000 | Clinically submitted variant-disease associations with pathogenicity classifications |
| **PharmGKB** | ~15 MB | ~20,000 | Drug-gene interactions, dosing guidelines, adverse effect predictions |
| **dbSNP (clinical subset)** | ~2 GB | ~500,000 | NCBI variant reference with clinical significance annotations |

These databases are maintained by NCBI, EBI, and Stanford/PharmGKB and are updated periodically. Genetica Resolutio downloads them directly from their official distribution URLs.

### GWAS Catalog

The GWAS Catalog, maintained by EMBL-EBI and NHGRI, aggregates published genome-wide association studies. Each entry maps a SNP to a reported trait or disease, along with the study's p-value and risk allele.

**Best for:** Broad coverage of complex traits and disease associations reported in published research. Adds thousands of associations across every health category.

**Confidence note:** GWAS associations range from robustly replicated to preliminary. The app derives confidence from the association's p-value: p < 10⁻⁸ = high, p < 10⁻⁵ = moderate, otherwise = low.

### ClinVar

ClinVar, maintained by NCBI, is a public archive of variants submitted by clinical laboratories, researchers, and participants in hereditary disease testing. Variants are classified by clinical significance: pathogenic, likely pathogenic, benign, and so on.

**Best for:** Hereditary disease risk variants. ClinVar is the gold standard for Mendelian conditions — single-gene disorders with strong, direct genotype-phenotype relationships.

**What gets imported:** Only pathogenic, likely pathogenic, and risk-factor classifications are imported. Benign and uncertain significance variants are filtered out.

### PharmGKB

PharmGKB, maintained by Stanford University, curates the literature on pharmacogenomics — how genetic variants affect drug response, metabolism, dosing, and toxicity. It forms the basis for many clinical pharmacogenomics guidelines.

**Best for:** Drug response and medication safety. If you take or are prescribed common medications, PharmGKB coverage significantly increases the Drug Response section of your report.

**Smallest download** (~15 MB) with the highest information density per variant. Recommended as a first install.

### dbSNP (clinical subset)

dbSNP is NCBI's comprehensive SNP reference database — it contains virtually every characterized human polymorphism. The full database is ~50 GB and not practical to import. Genetica Resolutio imports only the subset of variants that carry a clinical significance annotation (`CLNSIG` field in the VCF), producing ~500K clinically relevant entries from the ~2 GB compressed file.

**Best for:** Maximizing variant coverage for rarer hereditary conditions and filling gaps left by the other three databases.

**Note:** The download and import process for dbSNP is noticeably slower than the other sources due to file size. Expect several minutes on a typical connection.

---

## Downloading a database

1. Click **Download** in the action column of the database you want to install
2. A progress bar appears below the row showing download percentage and a status message
3. After the download completes, the app automatically parses and imports the data into your local SQLite database
4. When import finishes, the status column updates to show **Installed** and the action button changes to **Delete**

Downloads can be **cancelled** at any time by clicking **Cancel** while the progress bar is active. A partial download leaves no data in your database — the import only commits on full success.

You can download multiple databases. They are stored independently and any combination can be installed simultaneously.

---

## Cancelling a download

Click **Cancel** in the action column while a download is in progress. The download stops, the temporary file is deleted, and the row resets to its original state. Your existing database is unaffected.

---

## Deleting a database

If you no longer want a database installed:

1. Find the row for that source in the Databases tab
2. Click **Delete**
3. All variants from that source are removed from your local SQLite database

Deletion is immediate and cannot be undone. You can re-download the database at any time.

Deleting a database does not affect:
- The built-in 357-variant curated database (this cannot be removed)
- Any other installed databases

---

## Where databases are stored

Downloaded databases are stored in a single SQLite file:

| Platform | Path |
|---|---|
| Linux | `~/.local/share/genetica-resolutio/snpdb.sqlite` |
| macOS | `~/Library/Application Support/genetica-resolutio/snpdb.sqlite` |
| Windows | `%APPDATA%\genetica-resolutio\snpdb.sqlite` |

The file grows as you install databases. Approximate sizes on disk:

| Configuration | Approximate size |
|---|---|
| No optional databases | ~0 MB (file not created) |
| PharmGKB only | ~20 MB |
| GWAS Catalog | ~800 MB |
| ClinVar | ~500 MB |
| dbSNP clinical | ~1.5–2 GB |
| All four | ~3 GB |

---

## How databases affect analysis

When you run an analysis, the app performs two passes:

1. **Curated pass** — every rsID in your DNA file is looked up in the built-in 357-variant database. This always runs regardless of what optional databases are installed.

2. **SQLite pass** — if any optional databases are installed, every rsID is also looked up in the SQLite database. Results are merged with the curated findings.

**Deduplication:** If both the curated database and an optional database have an entry for the same rsID and the same trait, only one finding appears in your report. The curated entry takes precedence because it has been manually reviewed.

**Matched variant count:** Installing additional databases will increase the "variants matched" count in your report because more rsIDs in your DNA file will have database entries. A count of 150–300 is typical with only the curated database; with all four databases installed, counts in the thousands are common.

---

## Internet access

Database downloads require an internet connection. The download happens in a controlled, visible process when you explicitly click Download — no background network activity occurs otherwise.

During analysis (file parsing and report generation), no network access occurs. Analysis is entirely local.

See [Privacy & Security](privacy.md) for the full technical explanation.
