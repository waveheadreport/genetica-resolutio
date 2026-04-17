package backend

import (
	"fmt"
	"strings"
	"time"
)

// DB is the package-level curated SNP database, initialised once at startup.
var DB map[string][]SNPRecord

// InitDB seeds the in-memory curated database. Call once at app startup.
func InitDB() {
	DB = initSNPDB()
}

// GetDBStats returns summary statistics about the loaded databases.
func GetDBStats() DBStats {
	stats := DBStats{
		TotalSNPs:  len(DB),
		ByCategory: make(map[string]int),
	}
	for _, records := range DB {
		stats.TotalRecords += len(records)
		for _, r := range records {
			stats.ByCategory[r.Category]++
		}
	}
	if sqlCounts, err := SQLiteDBStats(); err == nil {
		for src, n := range sqlCounts {
			stats.ByCategory["db:"+src] += n
			stats.TotalRecords += n
			stats.TotalSNPs += n
		}
	}
	return stats
}

// interpretGenotype classifies a user's alleles against a SNP record.
func interpretGenotype(rec SNPRecord, a1, a2 string) (status, color, badge, effect string) {
	if rec.RiskAllele == "" {
		return "normal", "green", "Normal", rec.Desc
	}

	riskCount := 0
	if a1 == rec.RiskAllele {
		riskCount++
	}
	if a2 == rec.RiskAllele {
		riskCount++
	}

	switch {
	case riskCount == 2:
		return "homozygous_risk", "red", "High Risk",
			fmt.Sprintf("Homozygous for %s risk allele (%s/%s). %s", rec.RiskAllele, a1, a2, rec.Desc)
	case riskCount == 1:
		return "heterozygous", "amber", "Moderate",
			fmt.Sprintf("Heterozygous carrier of %s risk allele (%s/%s). %s", rec.RiskAllele, a1, a2, rec.Desc)
	default:
		if rec.RiskLevel == "protective" {
			return "protective", "green", "Protective",
				fmt.Sprintf("Favorable genotype detected (%s/%s). %s", a1, a2, rec.Desc)
		}
		return "normal", "green", "Normal",
			fmt.Sprintf("No risk allele detected (%s/%s). %s", a1, a2, rec.Desc)
	}
}

// RunAnalysis performs the full analysis, merging curated and SQLite databases.
func RunAnalysis(parsed *ParseResult) *AnalysisResult {
	categories := make(map[string][]Finding)
	summary := Summary{}

	// dedup key: rsid+trait to avoid showing the same finding twice
	// when both the curated DB and a downloaded DB cover the same SNP
	seen := make(map[string]bool)

	addFinding := func(rec SNPRecord, a1, a2 string) {
		key := rec.RSID + "|" + rec.Trait
		if seen[key] {
			return
		}
		seen[key] = true

		status, color, badge, effect := interpretGenotype(rec, a1, a2)
		f := Finding{
			SNPRecord: rec,
			A1:        a1,
			A2:        a2,
			Status:    status,
			Color:     color,
			Badge:     badge,
			Effect:    effect,
		}

		if rec.Category == "pharmacogenomics" {
			summary.Drugs = append(summary.Drugs, f)
			return
		}
		categories[rec.Category] = append(categories[rec.Category], f)
		switch status {
		case "homozygous_risk":
			summary.High = append(summary.High, f)
		case "heterozygous":
			summary.Moderate = append(summary.Moderate, f)
		case "protective":
			summary.Protective = append(summary.Protective, f)
		}
	}

	matched := 0

	// ── Pass 1: curated in-memory database (always available) ──────────
	for rsid, records := range DB {
		alleles, ok := parsed.SNPs[rsid]
		if !ok {
			continue
		}
		matched++
		for _, rec := range records {
			addFinding(rec, alleles[0], alleles[1])
		}
	}

	// ── Pass 2: downloaded databases (if any are installed) ─────────────
	if KV != nil {
		for rsid, alleles := range parsed.SNPs {
			sqlRecords, err := QuerySNPsByRSID(rsid)
			if err != nil || len(sqlRecords) == 0 {
				continue
			}
			if _, inCurated := DB[rsid]; !inCurated {
				matched++
			}
			for _, rec := range sqlRecords {
				addFinding(rec, alleles[0], alleles[1])
			}
		}

		// ── Pass 3: positional lookup for VCF variants with no rsID ─────
		for posKey, alleles := range parsed.UnresolvedSNPs {
			parts := strings.SplitN(posKey, ":", 2)
			if len(parts) != 2 {
				continue
			}
			rsid, err := QueryRSIDByPosition(parts[0], parts[1], parsed.RefBuild)
			if err != nil || rsid == "" {
				continue
			}
			if _, already := parsed.SNPs[rsid]; already {
				continue
			}
			records, err := QuerySNPsByRSID(rsid)
			if err != nil || len(records) == 0 {
				continue
			}
			matched++
			for _, rec := range records {
				addFinding(rec, alleles[0], alleles[1])
			}
		}
	}

	actionPlan := BuildActionPlan(summary, categories)
	doctorText := buildDoctorText(parsed, matched, summary)

	return &AnalysisResult{
		Parsed:      *parsed,
		Matched:     matched,
		Categories:  categories,
		Summary:     summary,
		ActionPlan:  actionPlan,
		DoctorText:  doctorText,
		GeneratedAt: time.Now().Format("January 2, 2006"),
	}
}

func buildDoctorText(parsed *ParseResult, matched int, summary Summary) string {
	var sb strings.Builder
	sb.WriteString("PATIENT GENETIC SUMMARY — For Clinical Reference\n")
	sb.WriteString(fmt.Sprintf("Provider: %s | Generated: %s\n", parsed.Provider, time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("SNPs in file: %d | Variants matched: %d\n\n", parsed.TotalSNPs, matched))

	sb.WriteString(fmt.Sprintf("HIGH-PRIORITY FINDINGS (%d):\n", len(summary.High)))
	for _, f := range summary.High {
		sb.WriteString(fmt.Sprintf("• %s (%s) — %s: %s\n", f.Gene, f.RSID, f.Trait, f.Effect))
	}

	sb.WriteString(fmt.Sprintf("\nMODERATE FINDINGS (%d):\n", len(summary.Moderate)))
	limit := len(summary.Moderate)
	if limit > 10 {
		limit = 10
	}
	for _, f := range summary.Moderate[:limit] {
		sb.WriteString(fmt.Sprintf("• %s (%s) — %s\n", f.Gene, f.RSID, f.Trait))
	}

	sb.WriteString(fmt.Sprintf("\nPHARMACOGENOMIC FINDINGS (%d):\n", len(summary.Drugs)))
	for _, f := range summary.Drugs {
		sb.WriteString(fmt.Sprintf("• %s (%s) — %s: %s\n", f.Gene, f.RSID, f.Trait, f.Badge))
	}

	sb.WriteString("\nNote: These are statistical associations from population genetics research. " +
		"They are not diagnoses. Many people with risk variants never develop the associated condition. " +
		"Consult with a genetic counselor or physician for clinical interpretation.\n")
	return sb.String()
}
