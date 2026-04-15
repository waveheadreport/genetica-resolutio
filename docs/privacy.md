# Privacy & Security

Your genetic data is among the most sensitive personal information that exists. This page explains exactly what happens to your data when you use Genetica Resolutio — in technical terms, not marketing language.

---

## The short version

- Your DNA file never leaves your computer
- The app has no server to send data to
- No analytics, telemetry, or crash reporting
- No account, login, or registration
- The analysis result exists only in the application window and is discarded when you close the app or start a new analysis

---

## What the app does with your file

When you select a DNA file and click **Analyze →**, the following happens entirely within the running application process on your machine:

1. The file is opened from its location on your filesystem using a standard OS file read. The app does not copy, move, or modify the file.

2. The file is read line by line into memory. Each line is parsed to extract the rsID and allele values. Only these key-value pairs (rsID → genotype) are retained. The raw file content is not stored.

3. The in-memory rsID map is cross-referenced against the curated variant database (compiled into the application binary) and, if installed, the local SQLite database on your machine.

4. Results are assembled into the report data structure in memory.

5. The JavaScript frontend receives the report object and renders it.

At no point is any data written to a file other than the SQLite database on your own machine. At no point is any data transmitted over a network.

---

## Network access

The application binary contains no code that connects to a remote server during analysis. There is no telemetry endpoint, no license validation server, no health check URL.

The only network activity the app performs is **database downloads**, which occur exclusively when you explicitly click the **Download** button in the Databases tab. These downloads:
- Fetch files from the official public distribution URLs of NCBI (ClinVar, dbSNP), EMBL-EBI (GWAS Catalog), and Stanford/PharmGKB
- Transfer no data about you, your file, or your analysis to those servers — a standard HTTP GET request for a public file
- Can be cancelled at any time
- Occur over whatever network connection your OS uses

If you do not use the Databases tab, the application makes zero outbound network connections during its entire lifetime.

---

## Local data storage

The only file the app writes to is the optional SQLite database:

| Platform | Path |
|---|---|
| Linux | `~/.local/share/genetica-resolutio/snpdb.sqlite` |
| macOS | `~/Library/Application Support/genetica-resolutio/snpdb.sqlite` |
| Windows | `%APPDATA%\genetica-resolutio\snpdb.sqlite` |

This file contains only the contents of the reference databases you downloaded (variant rsIDs, trait descriptions, risk alleles, and citations from public sources). It contains no information from your DNA file.

If this file does not exist (no databases have been downloaded), the app leaves no traces on disk beyond its own binary.

---

## Analysis results

The report generated from your analysis exists only in memory while the app window is open. It is never:
- Written to disk automatically
- Transmitted anywhere
- Cached between sessions

When you close the application or click **↩ New Analysis**, the result is discarded. The next time you open the app, there is no record of any previous analysis.

The only way a report is saved is if you explicitly use the **💾 Save Report** button, which opens a native OS save dialog and writes a plain-text doctor summary to whatever location you choose. This file contains a subset of your findings but no raw genotype data (no rsIDs, no allele values — only interpreted findings like "MTHFR C677T: impaired methylfolate conversion").

---

## The application binary

Genetica Resolutio is compiled from open-source Go code and ships as a single executable with the curated SNP database compiled directly into the binary as Go data structures. There is no dynamic loading of remote code, no update mechanism that replaces code, and no plugin system.

You can verify the contents of the binary by building from source. See [Building from Source](building-from-source.md).

---

## What Wails (the framework) does

Genetica Resolutio is built with Wails v2, which uses the OS's native WebView to render the interface. The WebView renders only local HTML/CSS/JS compiled into the binary — it does not browse the internet. Wails itself has no telemetry.

The WebView component is:
- **Linux**: WebKitGTK — the same engine used by GNOME Web
- **macOS**: WebKit — the same engine used by Safari
- **Windows**: WebView2 — the Chromium-based component pre-installed on Windows 10/11

---

## Threat model

What this design protects against:

- **Server-side data exposure**: There is no server. A breach of a remote infrastructure cannot expose your data because your data is never there.
- **Man-in-the-middle during analysis**: Analysis uses no network. No interception is possible.
- **Account credential theft**: There are no accounts. No credentials exist to steal.

What this design does not protect against:

- **Local machine access**: If someone has access to your computer and your DNA file is on the filesystem, they can read it. This is true of any locally stored file.
- **Screen capture / shoulder surfing**: Your report is displayed on screen. Physical security of your screen is your responsibility.
- **The saved report file**: If you use **💾 Save Report**, the resulting `.txt` file is an ordinary file on your filesystem subject to normal file access controls.
- **The DNA file itself**: Genetica Resolutio does not encrypt or modify your raw DNA file. Its privacy depends on your filesystem permissions and storage security.

---

## Questions

If you have specific questions about the app's behavior that this page does not answer, the source code is the authoritative reference. See [Building from Source](building-from-source.md) for instructions on obtaining and reviewing the code.
