package backend

// SNPRecord is one entry in the curated database.
type SNPRecord struct {
	RSID       string `json:"rsid"`
	Gene       string `json:"gene"`
	Variant    string `json:"variant"`
	Chrom      string `json:"chrom"`
	Category   string `json:"cat"`
	SubCat     string `json:"subcat"`
	RiskAllele string `json:"riskAllele"`
	RiskLevel  string `json:"riskLevel"` // high, moderate, low, protective
	Trait      string `json:"trait"`
	Desc       string `json:"desc"`
	Rec        string `json:"rec"`
	Confidence string `json:"conf"`
	PMID       string `json:"pmid"`
}

// DBStats summarises the loaded SNP database.
type DBStats struct {
	TotalSNPs      int            `json:"totalSNPs"`
	TotalRecords   int            `json:"totalRecords"`
	ByCategory     map[string]int `json:"byCategory"`
}

// ParseResult is the output of parsing a raw DNA file.
type ParseResult struct {
	Provider       string               `json:"provider"`
	TotalSNPs      int                  `json:"totalSNPs"`
	SNPs           map[string][2]string  `json:"snps"`           // rsid → [a1, a2]
	RefBuild       string               `json:"refBuild"`        // "GRCh37" or "GRCh38" if detected
	UnresolvedSNPs map[string][2]string  `json:"unresolvedSNPs"` // "chrom:pos" → [a1, a2] for VCF variants with no rsID
}

// Finding is a matched SNP with interpreted genotype status.
type Finding struct {
	SNPRecord
	A1     string `json:"a1"`
	A2     string `json:"a2"`
	Status string `json:"status"` // homozygous_risk, heterozygous, protective, normal
	Color  string `json:"color"`  // red, amber, green
	Badge  string `json:"badge"`
	Effect string `json:"effect"`
}

// ActionItem is a single recommendation bullet.
type ActionItem struct {
	Text string `json:"text"`
	Gene string `json:"gene"`
}

// ActionPlan groups recommendations by domain.
type ActionPlan struct {
	Diet        []ActionItem `json:"diet"`
	Supplements []ActionItem `json:"supplements"`
	Exercise    []ActionItem `json:"exercise"`
	Sleep       []ActionItem `json:"sleep"`
	Monitoring  []ActionItem `json:"monitoring"`
}

// Summary holds the top-level risk groupings.
type Summary struct {
	High       []Finding `json:"high"`
	Moderate   []Finding `json:"moderate"`
	Protective []Finding `json:"protective"`
	Drugs      []Finding `json:"drugs"`
}

// AnalysisResult is the full output returned to the frontend.
type AnalysisResult struct {
	Parsed     ParseResult            `json:"parsed"`
	Matched    int                    `json:"matched"`
	Categories map[string][]Finding   `json:"categories"`
	Summary    Summary                `json:"summary"`
	ActionPlan ActionPlan             `json:"actionPlan"`
	DoctorText string                 `json:"doctorText"`
	GeneratedAt string               `json:"generatedAt"`
}
