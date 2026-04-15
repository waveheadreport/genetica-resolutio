package backend

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var rsidRe = regexp.MustCompile(`^rs\d+$`)
var alleleRe = regexp.MustCompile(`^[ATCG]$`)

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

		// Try tab-delimited first (23andMe, AncestryDNA, most providers)
		parts := strings.Split(line, "\t")

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

		// VCF format: CHROM POS ID REF ALT QUAL FILTER INFO FORMAT SAMPLE
		if len(parts) >= 10 && !strings.HasPrefix(parts[0], "#") {
			rsid := strings.ToLower(strings.TrimSpace(parts[2]))
			if !rsidRe.MatchString(rsid) {
				continue
			}
			ref := strings.ToUpper(strings.TrimSpace(parts[3]))
			alt := strings.ToUpper(strings.TrimSpace(parts[4]))
			// Only handle simple single-nucleotide variants
			if len(ref) == 1 && len(alt) == 1 &&
				alleleRe.MatchString(ref) && alleleRe.MatchString(alt) {
				// Parse GT from sample column (format: GT:..., sample: 0/1:...)
				format := strings.Split(parts[8], ":")
				sample := strings.Split(parts[9], ":")
				gt := "0/0"
				for i, f := range format {
					if f == "GT" && i < len(sample) {
						gt = sample[i]
						break
					}
				}
				gt = strings.ReplaceAll(gt, "|", "/")
				alleles := strings.Split(gt, "/")
				if len(alleles) == 2 {
					toNuc := func(idx string) string {
						if idx == "0" {
							return ref
						}
						return alt
					}
					a1 := toNuc(alleles[0])
					a2 := toNuc(alleles[1])
					if alleleRe.MatchString(a1) && alleleRe.MatchString(a2) {
						result.SNPs[rsid] = [2]string{a1, a2}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	result.TotalSNPs = len(result.SNPs)
	return result, nil
}
