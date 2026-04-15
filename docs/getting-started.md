# Getting Started

This guide takes you from download to your first completed analysis in under five minutes.

---

## Step 1 — Get your raw DNA file

Genetica Resolutio works with the raw genotype export from your DNA testing provider, **not** the health or ancestry reports those services produce. You need the underlying data file — a plain text file containing your rsIDs and alleles.

### How to export your raw data

**23andMe**
1. Go to [you.23andme.com/tools/data](https://you.23andme.com/tools/data)
2. Click **Download Raw Data**
3. Complete the security verification
4. Download the `.txt` file (it may be delivered as a `.zip`)

**AncestryDNA**
1. Go to [www.ancestry.com/dna](https://www.ancestry.com/dna) → your test → **Settings**
2. Scroll to **Download DNA Raw Data**
3. Confirm your identity and download the `.zip`
4. Extract the `.txt` file from the archive

**MyHeritage**
1. Go to **DNA** → **Manage DNA kits** → **Actions** → **Download**
2. Select **Download raw DNA data** and confirm
3. Download the `.zip` and extract the `.csv` file inside

**FamilyTreeDNA**
1. Go to **myDNA** → **Download Raw Data**
2. Select **Autosomal Raw Data** and download
3. Extract the file from the `.zip`

**LivingDNA**
1. Go to **Account** → **Access your raw data**
2. Click **Download** and extract the file

The file you're looking for is typically 10–30 MB, tab-separated or comma-separated, and contains columns for rsID, chromosome, position, and allele(s). You do not need to unzip it — the app reads `.gz` compressed files directly.

---

## Step 2 — Download and install the app

Download the binary for your platform from the releases page. There is no installer — the file is the application.

| Platform | File |
|---|---|
| Linux (x86-64) | `genetica-resolutio` |
| macOS (Intel) | `genetica-resolutio.app` |
| macOS (Apple Silicon) | `genetica-resolutio-arm64.app` |
| Windows | `genetica-resolutio.exe` |

### Linux
```bash
chmod +x genetica-resolutio
./genetica-resolutio
```

Or double-click the file in your file manager. If your system asks about execution permissions, click **Run** or **Execute**.

### macOS
Double-click `genetica-resolutio.app`. On first launch, macOS Gatekeeper may show a warning because the app is not from the App Store. To allow it:

1. Open **System Settings** → **Privacy & Security**
2. Scroll to the security section and click **Open Anyway** next to the app name
3. Confirm in the dialog that appears

### Windows
Double-click `genetica-resolutio.exe`. Windows SmartScreen may show a warning on first run. Click **More info** then **Run anyway** to proceed.

---

## Step 3 — Run your first analysis

1. Launch the application
2. On the home screen, either drag your DNA file onto the drop zone or click **Browse Files** to select it
3. Your file name and size will appear below the drop zone
4. Click **Analyze →**
5. A progress screen shows real-time steps as your file is parsed and cross-referenced
6. When analysis completes, a summary modal shows your result counts — High Risk, Moderate, Protective, and Drug Interaction findings
7. Choose how you want to enter the report:
   - **Show My Action Plan** — jumps directly to your personalized recommendations
   - **View Full Report** — opens the complete tabbed report at the Overview
   - **Hide Disease Risk Findings** — opens the report with disease predisposition findings hidden (useful if you find that information stressful without clinical context)

That's it. Your full report is ready.

---

## What happens to your file?

Nothing is stored, transmitted, or logged. The app reads your file into memory, performs the analysis, and discards the raw data. The report exists only in the application window. Your DNA file remains exactly where you left it on your filesystem — the app never moves, copies, or modifies it.

See [Privacy & Security](privacy.md) for the full technical explanation.
