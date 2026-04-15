# FAQ

---

## General

### What DNA providers are supported?

23andMe, AncestryDNA, MyHeritage, FamilyTreeDNA, LivingDNA, and any standard VCF file. See [Supported File Formats](file-formats.md) for format details and column structure.

### Does the app require an internet connection?

Only to download optional reference databases from the Databases tab. Once downloaded, they are stored locally and all analysis runs entirely offline. If you do not use the Databases tab, the app never makes a network connection.

### Is my DNA data sent anywhere?

No. See [Privacy & Security](privacy.md) for the full technical explanation. The short answer: analysis is 100% local, there is no server, and no data leaves your machine.

### What does "357 curated variants" mean?

The app ships with 357 hand-selected SNPs drawn from replicated research literature. These cover the most actionable and well-studied variants across 15 health categories. This is the minimum you get without downloading any additional databases. With all four optional databases installed, coverage expands to over 1 million variants.

### Can I analyze someone else's DNA file?

The app processes whatever file you give it. There are no identity checks. However, genetic data is uniquely personal — ensure you have appropriate consent before analyzing someone else's file.

---

## File issues

### My file was rejected / nothing happened after I selected it

Check that:
- The file is a raw data export, not a health report PDF or ancestry summary
- If it came in a `.zip`, extract it first — the app reads `.gz` compressed files but not `.zip` archives
- The file is from a supported provider or follows standard VCF format

### My matched variant count is very low (under 20)

This can happen when:
- The file was not parsed correctly — the format was not recognized
- Your provider uses a genotyping chip that covers different positions than the curated database
- The file is corrupted or truncated

Try opening the raw file in a text editor and checking that it looks like a normal tab-separated or comma-separated table with rsIDs in the first column.

If the format looks correct but matching is low, install one or more optional databases from the Databases tab. The curated 357-variant set was built around variants commonly covered by 23andMe and AncestryDNA chips — other providers may have less overlap.

### The app says it detected the wrong provider

Provider detection is based on comment lines at the top of your file. If your file was processed through a third-party tool before uploading, those comment lines may have been stripped. The app falls back to column-count detection, which should still work for most formats. Provider identification affects labeling in the report but not the analysis itself.

### I have a `.vcf.gz` file from a sequencing service. Will it work?

Yes. VCF files, including gzip-compressed ones, are supported. The app reads the standard `ID`, `REF`, `ALT`, and `FORMAT`/`SAMPLE` columns. Only single-nucleotide variants with valid rsIDs are analyzed — structural variants and complex alleles are skipped.

---

## Results

### Why are some findings hidden by default?

Findings where you carry no copies of the risk allele ("Normal" status) are hidden by default because they require no action and add noise to the report. Use the **All findings** filter button on any category tab to show them.

### What does "Homozygous Risk" actually mean?

It means you carry two copies of the risk allele for that variant — one from each parent. This is the highest-risk genotype for that specific variant. It does not mean you will develop the associated condition. See [Understanding Your Results](understanding-results.md) for a fuller explanation, including statistics on how many people with high-risk genotypes develop various conditions.

### A finding says "High Risk" but I feel fine / have no symptoms

Genetic variants are statistical predispositions, not diagnoses. Many people with high-risk genotypes never develop the associated condition. Lifestyle, other genes, environment, and chance all play roles. The purpose of the report is to inform preventive choices, not to alarm.

### I have a finding I want to discuss with my doctor. How do I share it?

Use the **Doctor Summary** tab, which presents your key findings in clinical plain-text format. Click **Copy to clipboard** to paste into an email, patient portal message, or printed summary. Or use **💾 Save Report** in the toolbar to export a `.txt` file.

### The Action Plan recommends something I'm already doing. Is that right?

Yes — the Action Plan reflects what the evidence suggests is optimal for your genotype, regardless of current behavior. It's a target state, not a judgment on your current habits.

### Why does my Action Plan have "General" recommendations?

Some recommendations apply universally regardless of specific variants — for example, 7–9 hours of sleep, regular resistance training, or eating diverse vegetables. These are included because the evidence base is strong enough to apply to everyone, not because a specific gene variant drove them.

---

## Databases

### Which database should I install first?

**PharmGKB** if you take any regular medications — it's the smallest download (~15 MB) and has the highest practical value per variant for most people.

**GWAS Catalog** for the broadest increase in variant coverage across all health categories.

**ClinVar** for hereditary disease and clinical pathogenicity data.

**dbSNP clinical subset** last — it's the largest download and mostly overlaps with ClinVar for clinically significant variants.

### How long does a database download take?

Approximate download times on a 100 Mbps connection:

| Database | Time |
|---|---|
| PharmGKB | < 30 seconds |
| GWAS Catalog | 2–5 minutes |
| ClinVar | 5–10 minutes |
| dbSNP clinical | 15–30 minutes |

Import (parsing and inserting into SQLite) adds additional time, particularly for dbSNP. The progress bar shows both download and import phases.

### Can I use the app while a database is downloading?

No — avoid running an analysis while a download is in progress. The download is a background operation, but running analysis simultaneously could produce inconsistent results if the import is mid-write.

### Does downloading a database update automatically?

No. Databases are downloaded on demand and stored statically. They do not auto-update. To get the latest version of a database, delete it and re-download it.

### Where are the databases stored? How much disk space do they need?

See the [Database Manager](database-manager.md#where-databases-are-stored) page for paths and size estimates. All four databases together use approximately 3 GB.

---

## Installation and startup

### Linux: the app won't launch / shows a blank window

The app requires WebKitGTK 4.1. Install it with your package manager:

```bash
# Fedora / RHEL
sudo dnf install webkit2gtk4.1

# Ubuntu / Debian
sudo apt install libwebkit2gtk-4.1-dev

# Arch
sudo pacman -S webkit2gtk-4.1
```

If WebKitGTK is installed and the window is still blank, try launching from a terminal to see any error output:
```bash
./genetica-resolutio
```

### macOS: "cannot be opened because the developer cannot be verified"

This is Gatekeeper. The app is unsigned. To allow it:

1. Open **System Settings** → **Privacy & Security**
2. Scroll to the security section and click **Open Anyway** next to the app name
3. Confirm in the dialog that appears

Alternatively, from Terminal:
```bash
xattr -d com.apple.quarantine /path/to/genetica-resolutio.app
```

### Windows: SmartScreen warning on first launch

Click **More info** then **Run anyway**. SmartScreen shows this warning for any unsigned executable that hasn't been run by many users.

### The app is slow to start / the window takes a few seconds to appear

Normal on first launch — the OS WebView may take a moment to initialize. Subsequent launches are faster.

---

## Building and development

See [Building from Source](building-from-source.md) for complete build instructions.

### Can I modify the curated SNP database?

Yes. The curated database is defined in `backend/snpdata.go` as Go source code. Add or modify `add(...)` calls in the `initSNPDB()` function following the existing pattern, then rebuild.

### Is the source code available?

Yes. See [Building from Source](building-from-source.md) for repository access and build instructions.
