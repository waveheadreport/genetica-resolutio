# Genetica Resolutio — Documentation

**Your genome, decoded privately.**

Genetica Resolutio is a desktop application that analyzes your raw DNA file and generates a comprehensive, science-backed health report — entirely on your own device. No internet connection is required during analysis. Your genetic data never leaves your computer.

---

## Quick navigation

| | |
|---|---|
| [Getting Started](getting-started.md) | Download, install, and run your first analysis in under five minutes |
| [Using the App](using-the-app.md) | Step-by-step walkthrough of every screen and feature |
| [Understanding Your Results](understanding-results.md) | How to read your report, risk levels, and action plan |
| [Database Manager](database-manager.md) | Expand coverage with GWAS Catalog, ClinVar, PharmGKB, and dbSNP |
| [Supported File Formats](file-formats.md) | Every DNA provider format the app accepts |
| [Privacy & Security](privacy.md) | Exactly what happens to your data — and what doesn't |
| [FAQ](faq.md) | Common questions answered |
| [Building from Source](building-from-source.md) | For developers who want to compile or modify the app |

---

## What it does

When you upload a raw DNA file from any major consumer genetics provider (23andMe, AncestryDNA, MyHeritage, and others), Genetica Resolutio:

1. **Parses** your file locally — reading the rsID and allele pairs that make up your genotype
2. **Cross-references** your variants against a curated database of evidence-based SNPs, plus any optional reference databases you have installed
3. **Interprets** each matched variant — determining whether you carry zero, one, or two copies of the risk allele
4. **Generates** a full tabbed health report organized by category (nutrition, cardiovascular, sleep, mental health, and more)
5. **Builds** a personalized action plan — specific diet, supplement, exercise, sleep, and health monitoring recommendations derived from your variants

Everything happens on your machine. The analysis completes in seconds.

---

## Key features

- **357+ curated variants** built in — no setup required, works immediately after install
- **Optional database expansion** — download GWAS Catalog (~600K associations), ClinVar (~200K clinical variants), PharmGKB (~20K drug interactions), or dbSNP clinical subset (~500K variants) to dramatically increase coverage
- **15 health categories** — nutrition, supplements, fitness, sleep, cardiovascular, disease risk, hormones, mental health, longevity, immune, neurological, gut, bone, cancer risk, and pharmacogenomics
- **Personalized action plan** — every recommendation is tied to the specific variant that drives it
- **Doctor-friendly summary** — a plain-text export formatted for sharing with your physician or genetic counselor
- **Privacy by design** — the app has no network access during analysis; your DNA is never transmitted anywhere

---

## System requirements

| Platform | Requirement |
|---|---|
| **Linux** | Any modern distro with WebKitGTK 4.1 installed (most desktops include this) |
| **macOS** | macOS 10.13 (High Sierra) or later |
| **Windows** | Windows 10 or later (WebView2 is pre-installed on Windows 10/11) |

The application is a single executable — no installer, no dependencies to manage beyond what is already on your system.
