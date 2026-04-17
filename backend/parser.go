package backend

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var rsidRe = regexp.MustCompile(`^rs\d+$`)
var alleleRe = regexp.MustCompile(`^[ATCG]$`)
var infoRsIDRe = regexp.MustCompile(`(?:^|;)rsID=(rs\d+)`)

func extractRsIDFromInfo(info string) string {
	m := infoRsIDRe.FindStringSubmatch(info)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

func gtToNuc(idx string, allNucs []string) string {
	if idx == "." {
		return ""
	}
	i, err := strconv.Atoi(idx)
	if err != nil || i < 0 || i >= len(allNucs) {
		return ""
	}
	return allNucs[i]
}

// ParseDNAFile reads a raw DNA export file and returns a ParseResult.
// Supports: 23andMe, AncestryDNA, MyHeritage, FamilyTreeDNA, LivingDNA (tab/csv).
// Supports gzip-compressed files transparently.
func ParseDNAFile(path string) (*ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f

	// Transparent gzip decompression
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	return parseReader(reader)
}

func parseReader(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{
		Provider: "Unknown",
		SNPs:     make(map[string][2]string),
	}

	scanner := bufio.NewScanner(r)
	// 10 MB buffer for very large lines (some providers have long headers)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, cap(buf))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Comment/header lines: detect provider, skip
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			lower := strings.ToLower(line)
			switch {
			case strings.HasPrefix(lower, "##fileformat=vcf"):
				result.Provider = "VCF"
			case strings.Contains(lower, "ancestrydna"):
				result.Provider = "AncestryDNA"
			case strings.Contains(lower, "23andme"):
				result.Provider = "23andMe"
			case strings.Contains(lower, "myheritage"):
				result.Provider = "MyHeritage"
			case strings.Contains(lower, "familytreedna"), strings.Contains(lower, "ftdna"):
				result.Provider = "FamilyTreeDNA"
			case strings.Contains(lower, "livingdna"):
				result.Provider = "LivingDNA"
			case strings.Contains(lower, "nebula"):
				result.Provider = "Nebula Genomics"
			}
			continue
		}

		lower := strings.ToLower(line)
		// Skip column-header rows
		if strings.HasPrefix(lower, "rsid") ||
			strings.HasPrefix(lower, "name") ||
			strings.HasPrefix(lower, "snpid") ||
			lower == "#rsid" {
			continue
		}

		parts := strings.Split(line, "\t")

		// VCF format: CHROM POS ID REF ALT QUAL FILTER INFO FORMAT SAMPLE
		// Must come before generic tab checks since VCF also has 10+ tab fields.
		if len(parts) >= 10 {
			formatCols := strings.Split(parts[8], ":")
			hasGT := false
			for _, f := range formatCols {
				if f == "GT" {
					hasGT = true
					break
				}
			}
			if hasGT {
				rsid := strings.ToLower(strings.TrimSpace(parts[2]))
				if !rsidRe.MatchString(rsid) {
					rsid = extractRsIDFromInfo(parts[7])
					if rsid == "" {
						continue
					}
				}
				ref := strings.ToUpper(strings.TrimSpace(parts[3]))
				if !alleleRe.MatchString(ref) {
					continue
				}
				altField := strings.TrimSpace(parts[4])
				if altField == "." {
					continue
				}
				altAlleles := strings.Split(strings.ToUpper(altField), ",")
				allNucs := append([]string{ref}, altAlleles...)

				sample := strings.Split(parts[9], ":")
				gt := ""
				gtIdx := -1
				for i, f := range formatCols {
					if f == "GT" {
						gtIdx = i
						break
					}
				}
				if gtIdx >= 0 && gtIdx < len(sample) {
					gt = sample[gtIdx]
				}
				if gt == "" || gt == "." || gt == "./." {
					continue
				}
				gt = strings.ReplaceAll(gt, "|", "/")
				alleles := strings.Split(gt, "/")
				if len(alleles) != 2 {
					continue
				}
				a1 := gtToNuc(alleles[0], allNucs)
				a2 := gtToNuc(alleles[1], allNucs)
				if alleleRe.MatchString(a1) && alleleRe.MatchString(a2) {
					result.SNPs[rsid] = [2]string{a1, a2}
				}
				continue
			}
		}

		if len(parts) >= 5 {
			// AncestryDNA: rsid, chr, pos, allele1, allele2
			rsid := strings.ToLower(strings.TrimSpace(parts[0]))
			if !rsidRe.MatchString(rsid) {
				continue
			}
			a1 := strings.ToUpper(strings.TrimSpace(parts[3]))
			a2 := strings.ToUpper(strings.TrimSpace(parts[4]))
			if a1 == "-" || a2 == "-" || a1 == "0" || a2 == "0" {
				continue
			}
			if alleleRe.MatchString(a1) && alleleRe.MatchString(a2) {
				result.SNPs[rsid] = [2]string{a1, a2}
			}
			continue
		}

		if len(parts) == 4 {
			// 23andMe: rsid, chr, pos, genotype
			rsid := strings.ToLower(strings.TrimSpace(parts[0]))
			if !rsidRe.MatchString(rsid) {
				continue
			}
			geno := strings.ToUpper(strings.TrimSpace(parts[3]))
			if len(geno) == 2 && geno != "--" {
				a1 := string(geno[0])
				a2 := string(geno[1])
				if alleleRe.MatchString(a1) && alleleRe.MatchString(a2) {
					result.SNPs[rsid] = [2]string{a1, a2}
				}
			}
			continue
		}

		// Try comma-delimited (MyHeritage)
		cparts := strings.Split(line, ",")
		if len(cparts) >= 4 {
			rsid := strings.ToLower(strings.Trim(strings.TrimSpace(cparts[0]), `"`))
			if !rsidRe.MatchString(rsid) {
				continue
			}
			geno := strings.ToUpper(strings.Trim(strings.TrimSpace(cparts[3]), `"`))
			if len(geno) >= 2 {
				a1 := string(geno[0])
				a2 := string(geno[1])
				if alleleRe.MatchString(a1) && alleleRe.MatchString(a2) {
					result.SNPs[rsid] = [2]string{a1, a2}
				}
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	result.TotalSNPs = len(result.SNPs)
	return result, nil
}
