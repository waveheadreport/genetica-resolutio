# Using the App

A full walkthrough of every screen, panel, and button.

---

## The home screen

When you launch the app you land on the home screen. The topbar contains two tabs:

- **🧬 Analyze** — the file upload interface (default)
- **🗄 Databases** — the reference database manager

### Analyze tab

The left column contains an overview of the app's capabilities and a live count of variants in the active database. The right column contains the upload panel.

**Uploading a file**

- **Drag and drop** your DNA file anywhere on the drop zone
- Or click **Browse Files** to open a native file picker

The app accepts files from AncestryDNA, 23andMe, MyHeritage, FamilyTreeDNA, LivingDNA, and generic VCF files. Gzip-compressed files (`.gz`) are read directly without needing to decompress first.

Once a file is selected, a row appears showing the filename and size. Click **Analyze →** to begin.

### Databases tab

See the [Database Manager](database-manager.md) guide for full details. In brief: this tab shows the four major public reference databases you can optionally download to expand your analysis coverage beyond the 357 built-in curated variants.

---

## The progress screen

After clicking Analyze, the app switches to a progress screen showing real-time steps:

| Step | What is happening |
|---|---|
| Opening | Reading the file from disk |
| Indexing | Parsing all rsID / allele pairs from the file |
| Matching | Looking up each rsID in the curated and installed databases |
| Interpreting | Classifying each match as high risk, moderate, protective, or normal |
| Building | Constructing your personalized action plan |
| Rendering | Finalizing the report structure |

This typically takes 1–5 seconds for standard consumer genotype files (~700K SNPs). Files with optional databases installed (particularly ClinVar or GWAS Catalog) may take slightly longer.

---

## The welcome modal

When analysis completes, a modal appears showing four counts:

| Tile | Meaning |
|---|---|
| **High Risk** (red) | Variants where you carry two copies of the risk allele (homozygous) |
| **Moderate** (amber) | Variants where you carry one copy (heterozygous) |
| **Protective** (green) | Variants associated with a reduced risk or favorable effect |
| **Drug Interactions** (blue) | Pharmacogenomic variants affecting medication processing |

**Three entry points into the report:**

1. **⚡ Show My Action Plan** — skips directly to your personalized diet, supplement, exercise, sleep, and monitoring recommendations
2. **View Full Report** — opens the Overview tab with all findings
3. **Hide Disease Risk Findings** — opens the report with the Disease Risk category tab and its findings hidden. This is useful if you prefer to review those results with a healthcare provider rather than on your own.

---

## The report

The report is organized into a tab bar across the top. Tabs appear only for categories where you have matched findings.

### Overview tab

A dashboard of your most important findings grouped into three sections:

- **High Risk** — up to 10 of your highest-priority findings, each clickable to jump to the relevant category tab
- **Moderate** — up to 10 of your moderate findings
- **Protective** — all protective variants identified

Five statistics appear at the top: your DNA provider, how many SNPs were in your file, how many variants were matched, the total database size, and the date of analysis.

A banner at the top of the Overview invites you to jump to your Action Plan.

### ⚡ Action Plan tab

The Action Plan is the most actionable part of the report. It is built entirely from your specific variants — not generic advice.

Five expandable panels:

| Panel | Contents |
|---|---|
| **🥗 Diet & Nutrition** | Food choices, dietary patterns, nutrients to prioritize or avoid |
| **💊 Supplement Protocol** | Specific supplements with doses, tied to the gene variant driving each recommendation |
| **🏋️ Exercise & Training** | Training style, intensity, frequency based on your muscle fiber and metabolic variants |
| **🌙 Sleep Optimization** | Circadian timing, environment, and habits relevant to your chronotype and sleep variants |
| **🩺 Health Monitoring** | Blood tests, screenings, and check-up intervals recommended based on your risk variants |

Each bullet in the panels shows the gene name that drove the recommendation (e.g., `MTHFR`, `APOE`, `ACTN3`). Generic recommendations that apply to everyone show `General`.

Click any panel header to expand or collapse it.

### Category tabs

Each health category has its own tab. Categories present in your results:

| Tab | What it covers |
|---|---|
| 🥗 Nutrition & Eating | Folate metabolism, vitamin absorption, macronutrient response, food intolerances |
| 💊 Supplement Protocol | B vitamins, vitamin D, omega-3s, minerals, antioxidants |
| 🏋️ Exercise & Training | Muscle fiber type, VO2 max, endurance vs. power, injury risk |
| 🌙 Sleep & Recovery | Circadian rhythm, sleep quality, melatonin, restless legs |
| 🫀 Cardiovascular Health | CAD risk, blood pressure, cholesterol, inflammation, clotting |
| 🔬 Disease Risk | Alzheimer's, diabetes, autoimmune, iron overload, and more |
| ⚗️ Hormones & Metabolism | Insulin sensitivity, thyroid, cortisol, estrogen, testosterone |
| 🧠 Mental Health & Cognition | Dopamine, serotonin, COMT, stress response, mood regulation |
| ⏳ Longevity & Aging | Telomere biology, epigenetic aging, senescence pathways |
| 🛡️ Immune Function | Inflammatory cytokines, autoimmune susceptibility, infection response |
| ⚡ Neurological | Migraine, nerve conduction, neurodegeneration risk |
| 🦠 Gut Health | Microbiome-related variants, lactose, gluten sensitivity |
| 🦴 Bone & Joint | Osteoporosis risk, collagen genes, tendon and ligament susceptibility |
| 🎗️ Cancer Risk | BRCA, colorectal, melanoma, and other cancer predisposition variants |
| 💉 Drug Response | Pharmacogenomics — how you metabolize specific medications |

**Within each category tab:**

- Findings are sorted by severity: High Risk first, then Moderate, then Protective, then Normal
- If there are "Normal" findings (no risk allele detected), they are hidden by default to reduce noise
- Use the **Actionable only / All findings** filter buttons to toggle normal findings on or off

**Each finding card shows:**

- Trait name and severity badge (High Risk / Moderate / Protective / Normal)
- Gene name, rsID, and your specific genotype (e.g., `A/G`)
- A description of what the variant does and how your specific genotype is classified
- Your personalized recommendation
- Confidence level and a PubMed citation link to the source study

### 💉 Drug Response tab

A table showing all pharmacogenomic variants found in your file. Columns:

- **Drug / Category** — the medication or drug class affected
- **Gene** — the gene variant responsible, with its rsID
- **Your Effect** — what your specific genotype means for how you process this drug
- **Action Required** — what to discuss with your prescriber
- **Level** — severity classification (High Risk or Moderate)

Share this tab with your doctor or pharmacist before starting any new medications.

### 🩺 Doctor Summary tab

A plain-text summary of your most important findings, formatted for clinical communication. Contains:

- High-priority findings with gene, rsID, trait, and brief effect description
- Moderate findings (up to 10)
- Pharmacogenomic findings
- A standard disclaimer about statistical associations vs. clinical diagnosis

Use the **Copy to clipboard** button to paste into an email or patient portal message. Or click **💾 Save Report** in the toolbar to save it as a `.txt` file.

---

## The toolbar (report screen)

Three buttons appear in the top-right of the report screen:

| Button | Action |
|---|---|
| **⚠️ Risk** | Toggle disease risk findings on/off. When active (amber highlight), the Disease Risk tab and its findings are hidden |
| **💾 Save Report** | Opens a native save dialog to export the doctor-friendly plain-text summary |
| **↩ New Analysis** | Returns to the home screen to analyze a different file |

---

## Tips

- **Tab navigation**: The tab bar scrolls horizontally if you have many categories. Swipe or use your trackpad.
- **Finding links**: PubMed citation links in finding cards open in your default browser.
- **Zoom**: Use your OS zoom controls (`Ctrl +` / `Cmd +`) to increase text size if needed.
- **Multiple files**: Use **↩ New Analysis** to analyze a second file. Results from previous analyses are not saved — each session is fresh.
