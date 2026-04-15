package backend

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// ── GWAS CATALOG ─────────────────────────────────────────────────────────────
// Source: https://ftp.ebi.ac.uk/pub/databases/gwas/releases/latest/gwas-catalog-associations_ontology-annotated-full.zip
// Format: ZIP archive containing a single tab-separated file with header.
// Key columns: SNPS, REPORTED GENE(S), MAPPED_GENE, DISEASE/TRAIT,
//              STRONGEST SNP-RISK ALLELE, PUBMEDID, P-VALUE, MAPPED_TRAIT

func ImportGWASCatalog(ctx context.Context, filePath string, fn ProgressFn) (int, error) {
	fn(DownloadProgress{Phase: "importing", Message: "Extracting GWAS Catalog ZIP…"})

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot open GWAS zip: %w", err)
	}
	defer zr.Close()

	var tsvFile *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".tsv") {
			tsvFile = f
			break
		}
	}
	if tsvFile == nil {
		return 0, fmt.Errorf("no .tsv file found in GWAS zip")
	}

	r, err := tsvFile.Open()
	if err != nil {
		return 0, err
	}
	defer r.Close()

	fn(DownloadProgress{Phase: "importing", Message: "Parsing GWAS Catalog…"})

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))

	// Parse header
	if !scanner.Scan() {
		return 0, fmt.Errorf("empty GWAS file")
	}
	headers := strings.Split(scanner.Text(), "\t")
	idx := makeIndex(headers)

	col := func(row []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var batch []SNPRecord
	total := 0
	lineNo := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		lineNo++
		if lineNo%50000 == 0 {
			fn(DownloadProgress{Phase: "importing", Message: fmt.Sprintf("GWAS: imported %d records…", total)})
			if err := BulkInsertSNPs(batch, "gwas"); err != nil {
				return total, err
			}
			total += len(batch)
			batch = batch[:0]
		}

		row := strings.Split(scanner.Text(), "\t")

		// Extract rsID from "SNPS" column (may be "rs123" or "rs123 x rs456")
		snpsField := col(row, "SNPS")
		rsid := extractFirstRSID(snpsField)
		if rsid == "" {
			continue
		}

		trait := col(row, "DISEASE/TRAIT")
		if trait == "" {
			trait = col(row, "MAPPED_TRAIT")
		}
		if trait == "" {
			continue
		}

		gene := col(row, "MAPPED_GENE")
		if gene == "" {
			gene = col(row, "REPORTED GENE(S)")
		}
		// Truncate multi-gene fields
		if i := strings.IndexAny(gene, ", "); i > 0 {
			gene = gene[:i]
		}

		// Risk allele: "STRONGEST SNP-RISK ALLELE" looks like "rs123-A"
		riskAllele := ""
		strongest := col(row, "STRONGEST SNP-RISK ALLELE")
		if i := strings.LastIndex(strongest, "-"); i >= 0 && i < len(strongest)-1 {
			a := strings.ToUpper(strongest[i+1:])
			if len(a) == 1 && strings.ContainsAny(a, "ATCG") {
				riskAllele = a
			}
		}

		pval := col(row, "P-VALUE")
		confidence := pvalToConfidence(pval)

		batch = append(batch, SNPRecord{
			RSID:       strings.ToLower(rsid),
			Gene:       gene,
			Category:   gwasTraitToCategory(trait),
			SubCat:     "gwas_association",
			RiskAllele: riskAllele,
			RiskLevel:  "moderate",
			Trait:      truncate(trait, 120),
			Desc:       fmt.Sprintf("GWAS association: %s (p=%s)", truncate(trait, 80), pval),
			Confidence: confidence,
			PMID:       col(row, "PUBMEDID"),
		})
	}

	if len(batch) > 0 {
		if err := BulkInsertSNPs(batch, "gwas"); err != nil {
			return total, err
		}
		total += len(batch)
	}

	return total, scanner.Err()
}

// ── CLINVAR ──────────────────────────────────────────────────────────────────
// Source: https://ftp.ncbi.nlm.nih.gov/pub/clinvar/vcf_GRCh38/clinvar.vcf.gz
// Format: gzip VCF; INFO contains CLNSIG, CLNDN, GENEINFO, RS

func ImportClinVar(ctx context.Context, filePath string, fn ProgressFn) (int, error) {
	r, err := OpenMaybeGzip(filePath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	fn(DownloadProgress{Phase: "importing", Message: "Parsing ClinVar VCF…"})

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))

	var batch []SNPRecord
	total := 0
	lineNo := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		lineNo++
		if lineNo%20000 == 0 {
			fn(DownloadProgress{Phase: "importing", Message: fmt.Sprintf("ClinVar: imported %d records…", total)})
			if err := BulkInsertSNPs(batch, "clinvar"); err != nil {
				return total, err
			}
			total += len(batch)
			batch = batch[:0]
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		chrom := strings.TrimPrefix(fields[0], "chr")
		// vcfID field may contain rsID
		vcfID := fields[2]
		ref := strings.ToUpper(fields[3])
		alt := strings.ToUpper(fields[4])
		info := fields[7]

		// Only process SNVs
		if len(ref) != 1 || len(alt) != 1 {
			continue
		}

		infoMap := parseVCFInfo(info)

		// Get rsID: prefer RS= tag in INFO, fall back to ID field
		rsid := ""
		if rs, ok := infoMap["RS"]; ok {
			rsid = "rs" + rs
		} else if strings.HasPrefix(strings.ToLower(vcfID), "rs") {
			rsid = strings.ToLower(vcfID)
		}
		if rsid == "" || !rsidRe.MatchString(rsid) {
			continue
		}

		clnsig := infoMap["CLNSIG"]
		clndn  := infoMap["CLNDN"]
		gene   := ""
		if gi, ok := infoMap["GENEINFO"]; ok {
			// GENEINFO = "GENE:GENEID|..."
			if i := strings.Index(gi, ":"); i > 0 {
				gene = gi[:i]
			}
		}

		if clndn == "" || clndn == "not_provided" || clndn == "not_specified" {
			continue
		}
		clndn = strings.ReplaceAll(clndn, "_", " ")
		clnsig = strings.ReplaceAll(clnsig, "_", " ")

		riskLevel := clinSigToRiskLevel(clnsig)
		if riskLevel == "" {
			continue // benign/likely benign — skip
		}

		batch = append(batch, SNPRecord{
			RSID:       strings.ToLower(rsid),
			Gene:       gene,
			Chrom:      chrom,
			Category:   "disease_risk",
			SubCat:     "clinvar",
			RiskAllele: alt,
			RiskLevel:  riskLevel,
			Trait:      truncate(clndn, 120),
			Desc:       fmt.Sprintf("ClinVar: %s. Clinical significance: %s.", truncate(clndn, 100), clnsig),
			Confidence: "high",
		})
	}

	if len(batch) > 0 {
		if err := BulkInsertSNPs(batch, "clinvar"); err != nil {
			return total, err
		}
		total += len(batch)
	}

	return total, scanner.Err()
}

// ── PHARMGKB ─────────────────────────────────────────────────────────────────
// Source: https://api.pharmgkb.org/v1/download/file/data/clinicalAnnotations.zip
// Contains: clinical_annotations.tsv inside the ZIP

func ImportPharmGKB(ctx context.Context, filePath string, fn ProgressFn) (int, error) {
	fn(DownloadProgress{Phase: "importing", Message: "Extracting PharmGKB ZIP…"})

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot open PharmGKB zip: %w", err)
	}
	defer zr.Close()

	var tsvFile *zip.File
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, "clinical_annotations.tsv") ||
			strings.HasSuffix(name, "clinicalannotations.tsv") {
			tsvFile = f
			break
		}
	}
	if tsvFile == nil {
		return 0, fmt.Errorf("clinical_annotations.tsv not found in PharmGKB zip")
	}

	rc, err := tsvFile.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	fn(DownloadProgress{Phase: "importing", Message: "Parsing PharmGKB clinical annotations…"})

	reader := csv.NewReader(rc)
	reader.Comma = '\t'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return 0, err
	}
	idx := makeIndex(headers)

	col := func(row []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var batch []SNPRecord
	total := 0
	lineNo := 0

	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		lineNo++
		if lineNo%5000 == 0 {
			fn(DownloadProgress{Phase: "importing", Message: fmt.Sprintf("PharmGKB: imported %d records…", total)})
			if err := BulkInsertSNPs(batch, "pharmgkb"); err != nil {
				return total, err
			}
			total += len(batch)
			batch = batch[:0]
		}

		// PharmGKB TSV columns vary by version; try common names
		variant := col(row, "Variant/Haplotypes")
		if variant == "" {
			variant = col(row, "Variant")
		}
		rsid := extractFirstRSID(variant)
		if rsid == "" {
			continue
		}

		drug := col(row, "Drug(s)")
		if drug == "" {
			drug = col(row, "Drugs")
		}
		gene := col(row, "Gene")
		pheno := col(row, "Phenotype(s)")
		if pheno == "" {
			pheno = col(row, "Phenotype")
		}
		level := col(row, "Level of Evidence")
		sig := col(row, "Clinical Annotation Types")

		trait := drug
		if pheno != "" {
			trait = drug + " — " + pheno
		}

		batch = append(batch, SNPRecord{
			RSID:       strings.ToLower(rsid),
			Gene:       gene,
			Category:   "pharmacogenomics",
			SubCat:     truncate(drug, 60),
			RiskAllele: "",
			RiskLevel:  pharmLevelToRisk(level),
			Trait:      truncate(trait, 120),
			Desc:       fmt.Sprintf("PharmGKB: %s. Type: %s.", truncate(pheno, 100), sig),
			Rec:        fmt.Sprintf("Discuss %s with your prescriber; this variant affects drug response.", truncate(drug, 60)),
			Confidence: pharmLevelToConfidence(level),
		})
	}

	if len(batch) > 0 {
		if err := BulkInsertSNPs(batch, "pharmgkb"); err != nil {
			return total, err
		}
		total += len(batch)
	}

	return total, nil
}

// ── dbSNP ────────────────────────────────────────────────────────────────────
// Source: per-chromosome VCFs from NCBI FTP (clinically-associated subset)
// https://ftp.ncbi.nlm.nih.gov/snp/latest_release/VCF/GCF_000001405.40.gz
// This is the full dbSNP — very large. We import only variants with
// clinical significance annotations (CLNSIG in INFO).

func ImportdbSNP(ctx context.Context, filePath string, fn ProgressFn) (int, error) {
	r, err := OpenMaybeGzip(filePath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	fn(DownloadProgress{Phase: "importing", Message: "Parsing dbSNP VCF (clinical variants only)…"})

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))

	var batch []SNPRecord
	total := 0
	lineNo := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		lineNo++
		if lineNo%100000 == 0 {
			fn(DownloadProgress{Phase: "importing", Message: fmt.Sprintf("dbSNP: processed %d lines, kept %d records…", lineNo, total)})
			if err := BulkInsertSNPs(batch, "dbsnp"); err != nil {
				return total, err
			}
			total += len(batch)
			batch = batch[:0]
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		vcfID := fields[2]
		if !strings.HasPrefix(strings.ToLower(vcfID), "rs") {
			continue
		}
		rsid := strings.ToLower(vcfID)
		if !rsidRe.MatchString(rsid) {
			continue
		}

		ref := strings.ToUpper(fields[3])
		alt := strings.ToUpper(fields[4])
		if len(ref) != 1 || len(alt) != 1 {
			continue // SNVs only
		}

		info := fields[7]
		infoMap := parseVCFInfo(info)

		// Only keep variants with clinical annotations
		clnsig, hasCLN := infoMap["CLNSIG"]
		clndn  := infoMap["CLNDN"]
		if !hasCLN || clnsig == "" {
			continue
		}

		riskLevel := clinSigToRiskLevel(clnsig)
		if riskLevel == "" {
			continue
		}

		gene := ""
		if gi, ok := infoMap["GENEINFO"]; ok {
			if i := strings.Index(gi, ":"); i > 0 {
				gene = gi[:i]
			}
		}

		clndn = strings.ReplaceAll(clndn, "_", " ")
		trait := truncate(clndn, 120)
		if trait == "" {
			trait = "Clinically associated variant"
		}

		batch = append(batch, SNPRecord{
			RSID:       rsid,
			Gene:       gene,
			Chrom:      strings.TrimPrefix(fields[0], "chr"),
			Category:   "disease_risk",
			SubCat:     "dbsnp_clinical",
			RiskAllele: alt,
			RiskLevel:  riskLevel,
			Trait:      trait,
			Desc:       fmt.Sprintf("dbSNP clinically-associated variant. Significance: %s.", strings.ReplaceAll(clnsig, "_", " ")),
			Confidence: "moderate",
		})
	}

	if len(batch) > 0 {
		if err := BulkInsertSNPs(batch, "dbsnp"); err != nil {
			return total, err
		}
		total += len(batch)
	}

	return total, scanner.Err()
}

// ── HELPERS ──────────────────────────────────────────────────────────────────

func makeIndex(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.TrimSpace(h)] = i
	}
	return m
}

func extractFirstRSID(s string) string {
	// find first rs\d+ token in a string
	s = strings.ToLower(s)
	for _, token := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';' || r == 'x' || r == '\t'
	}) {
		token = strings.TrimSpace(token)
		if rsidRe.MatchString(token) {
			return token
		}
	}
	return ""
}

func parseVCFInfo(info string) map[string]string {
	m := make(map[string]string)
	for _, kv := range strings.Split(info, ";") {
		if i := strings.Index(kv, "="); i > 0 {
			m[kv[:i]] = kv[i+1:]
		} else {
			m[kv] = "1"
		}
	}
	return m
}

func pvalToConfidence(p string) string {
	// Very rough: p < 5e-8 is genome-wide significant
	if strings.Contains(p, "E-") || strings.Contains(p, "e-") {
		return "high"
	}
	return "moderate"
}

func gwasTraitToCategory(trait string) string {
	t := strings.ToLower(trait)
	switch {
	case containsStr(t, "alzheimer") || containsStr(t, "dementia") || containsStr(t, "parkinson"):
		return "disease_risk"
	case containsStr(t, "coronary") || containsStr(t, "cardiac") || containsStr(t, "heart") || containsStr(t, "blood pressure") || containsStr(t, "cholesterol"):
		return "cardiovascular"
	case containsStr(t, "diabetes") || containsStr(t, "insulin") || containsStr(t, "glucose") || containsStr(t, "bmi") || containsStr(t, "obesity"):
		return "hormones"
	case containsStr(t, "cancer") || containsStr(t, "carcinoma") || containsStr(t, "tumor") || containsStr(t, "lymphoma"):
		return "cancer_risk"
	case containsStr(t, "sleep") || containsStr(t, "chronotype") || containsStr(t, "insomnia"):
		return "sleep"
	case containsStr(t, "depression") || containsStr(t, "anxiety") || containsStr(t, "bipolar") || containsStr(t, "schizophrenia"):
		return "mental_health"
	case containsStr(t, "vitamin") || containsStr(t, "folate") || containsStr(t, "iron") || containsStr(t, "calcium"):
		return "nutrition"
	case containsStr(t, "drug") || containsStr(t, "medication") || containsStr(t, "pharmacok"):
		return "pharmacogenomics"
	case containsStr(t, "bone") || containsStr(t, "osteoporosis") || containsStr(t, "fracture"):
		return "bone"
	case containsStr(t, "immune") || containsStr(t, "autoimmune") || containsStr(t, "lupus") || containsStr(t, "rheumatoid"):
		return "immune"
	case containsStr(t, "longevity") || containsStr(t, "aging") || containsStr(t, "telomere"):
		return "longevity"
	default:
		return "disease_risk"
	}
}

func clinSigToRiskLevel(sig string) string {
	s := strings.ToLower(sig)
	switch {
	case containsStr(s, "pathogenic") && !containsStr(s, "likely"):
		return "high"
	case containsStr(s, "likely_pathogenic") || containsStr(s, "likely pathogenic"):
		return "moderate"
	case containsStr(s, "risk_factor") || containsStr(s, "risk factor"):
		return "moderate"
	case containsStr(s, "benign") || containsStr(s, "likely_benign"):
		return "" // skip
	case containsStr(s, "protective"):
		return "protective"
	case containsStr(s, "uncertain") || containsStr(s, "conflicting"):
		return "low"
	default:
		return "" // skip unknown
	}
}

func pharmLevelToRisk(level string) string {
	switch strings.TrimSpace(level) {
	case "1A", "1B":
		return "high"
	case "2A", "2B":
		return "moderate"
	default:
		return "low"
	}
}

func pharmLevelToConfidence(level string) string {
	switch strings.TrimSpace(level) {
	case "1A":
		return "high"
	case "1B", "2A":
		return "moderate"
	default:
		return "low"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// TempFilePath returns a path for a temporary download file with the given extension.
func TempFilePath(ext string) (string, error) {
	f, err := os.CreateTemp("", "gr-import-*"+ext)
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
