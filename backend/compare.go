package backend

import "strings"

// ComparisonRow describes a single rsID present in both files with differing genotypes.
type ComparisonRow struct {
	RSID       string   `json:"rsid"`
	Gene       string   `json:"gene"`
	Trait      string   `json:"trait"`
	Category   string   `json:"cat"`
	A1Geno     string   `json:"a1Geno"` // e.g. "A/G"
	A2Geno     string   `json:"a2Geno"`
	A1Status   string   `json:"a1Status"`
	A2Status   string   `json:"a2Status"`
	RiskAllele string   `json:"riskAllele"`
	Notes      []string `json:"notes"` // e.g. "inheritance: heterozygous→homozygous_risk"
}

// ComparisonResult is the output of comparing two parsed DNA files.
type ComparisonResult struct {
	A           ParseResult     `json:"a"`
	B           ParseResult     `json:"b"`
	CommonSNPs  int             `json:"commonSNPs"`
	Identical   int             `json:"identical"`
	Differ      int             `json:"differ"`
	OnlyA       int             `json:"onlyA"`
	OnlyB       int             `json:"onlyB"`
	Rows        []ComparisonRow `json:"rows"` // only rsIDs that differ AND are in the curated/reference set
}

// CompareTwoFiles parses both files and returns a diff of their annotated variants.
func CompareTwoFiles(pathA, pathB string) (*ComparisonResult, error) {
	a, err := ParseDNAFile(pathA)
	if err != nil {
		return nil, err
	}
	b, err := ParseDNAFile(pathB)
	if err != nil {
		return nil, err
	}
	out := &ComparisonResult{A: *a, B: *b}

	// Count overlap / differences across all rsIDs (cheap set math).
	for rsid, aAll := range a.SNPs {
		if bAll, ok := b.SNPs[rsid]; ok {
			out.CommonSNPs++
			if sortedGeno(aAll) == sortedGeno(bAll) {
				out.Identical++
			} else {
				out.Differ++
			}
		} else {
			out.OnlyA++
		}
	}
	for rsid := range b.SNPs {
		if _, ok := a.SNPs[rsid]; !ok {
			out.OnlyB++
		}
	}

	// Build interpreted diff rows only for rsIDs that have reference annotations,
	// so the result is meaningful instead of millions of no-op rows.
	runA := RunAnalysis(a)
	runB := RunAnalysis(b)

	indexByRSID := func(res *AnalysisResult) map[string]Finding {
		m := make(map[string]Finding)
		for _, fs := range res.Categories {
			for _, f := range fs {
				if _, ok := m[f.RSID]; !ok {
					m[f.RSID] = f
				}
			}
		}
		return m
	}
	idxA := indexByRSID(runA)
	idxB := indexByRSID(runB)

	for rsid, fa := range idxA {
		fb, ok := idxB[rsid]
		if !ok {
			continue
		}
		if fa.Status == fb.Status && fa.A1 == fb.A1 && fa.A2 == fb.A2 {
			continue
		}
		row := ComparisonRow{
			RSID:       rsid,
			Gene:       fa.Gene,
			Trait:      fa.Trait,
			Category:   fa.Category,
			A1Geno:     fa.A1 + "/" + fa.A2,
			A2Geno:     fb.A1 + "/" + fb.A2,
			A1Status:   fa.Status,
			A2Status:   fb.Status,
			RiskAllele: fa.RiskAllele,
		}
		if fa.Status != fb.Status {
			row.Notes = append(row.Notes, "Status: "+fa.Status+" → "+fb.Status)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

func sortedGeno(alleles [2]string) string {
	x, y := strings.ToUpper(alleles[0]), strings.ToUpper(alleles[1])
	if x > y {
		return y + x
	}
	return x + y
}
