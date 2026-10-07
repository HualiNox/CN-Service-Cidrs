package generator

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeDomainRules(rules []string, path, name string) error {
	parsedCount := len(rules)
	rules = minimizeDomainRules(rules)
	counts := countDomainRuleTypes(rules)
	log.Printf(
		"domain output %q: parsed=%d, after dedup/merge=%d (domain=%d, full=%d, regexp=%d, keyword=%d)",
		filepath.Join(path, name),
		parsedCount,
		len(rules),
		counts["domain"],
		counts["full"],
		counts["regexp"],
		counts["keyword"],
	)
	if len(rules) == 0 {
		log.Printf("no domain rules found for %q; skipping output", filepath.Join(path, name))
		return nil
	}
	if err := mkdirAll(path, false); err != nil {
		return fmt.Errorf("prepare output directory %q: %w", path, err)
	}
	outputPath := filepath.Join(path, name+"-domain.txt")
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer func() { _ = file.Close() }()
	for _, rule := range rules {
		if _, err := io.WriteString(file, rule+"\n"); err != nil {
			return fmt.Errorf("write output file %q: %w", outputPath, err)
		}
	}
	return nil
}

func countDomainRuleTypes(rules []string) map[string]int {
	counts := map[string]int{
		"domain":  0,
		"full":    0,
		"regexp":  0,
		"keyword": 0,
	}
	for _, rule := range rules {
		kind, _, ok := strings.Cut(rule, ":")
		if ok {
			if _, supported := counts[kind]; supported {
				counts[kind]++
			}
		}
	}
	return counts
}

func minimizeDomainRules(rules []string) []string {
	domainValues := make(map[string]struct{})
	for _, rule := range rules {
		kind, value, ok := strings.Cut(rule, ":")
		if ok && kind == "domain" {
			domainValues[strings.ToLower(value)] = struct{}{}
		}
	}

	orderedDomains := make([]string, 0, len(domainValues))
	for value := range domainValues {
		orderedDomains = append(orderedDomains, value)
	}
	sort.Slice(orderedDomains, func(i, j int) bool {
		iLabels := strings.Count(orderedDomains[i], ".")
		jLabels := strings.Count(orderedDomains[j], ".")
		if iLabels != jLabels {
			return iLabels < jLabels
		}
		return orderedDomains[i] < orderedDomains[j]
	})

	keptDomains := make(map[string]struct{}, len(orderedDomains))
	for _, candidate := range orderedDomains {
		covered := false
		for dot := strings.IndexByte(candidate, '.'); dot >= 0; {
			parent := candidate[dot+1:]
			if _, exists := keptDomains[parent]; exists {
				covered = true
				break
			}
			nextDot := strings.IndexByte(candidate[dot+1:], '.')
			if nextDot < 0 {
				break
			}
			dot += nextDot + 1
		}
		if !covered {
			keptDomains[candidate] = struct{}{}
		}
	}

	result := make([]string, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		kind, value, ok := strings.Cut(rule, ":")
		if !ok {
			if _, exists := seen[rule]; !exists {
				seen[rule] = struct{}{}
				result = append(result, rule)
			}
			continue
		}

		valueKey := value
		switch kind {
		case "domain":
			valueKey = strings.ToLower(value)
			if _, keep := keptDomains[valueKey]; !keep {
				continue
			}
		case "full":
			valueKey = strings.ToLower(value)
			if coveredByDomainSuffix(valueKey, keptDomains) {
				continue
			}
		}

		key := kind + ":" + valueKey
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, rule)
	}
	return result
}

func coveredByDomainSuffix(host string, suffixes map[string]struct{}) bool {
	if _, exists := suffixes[host]; exists {
		return true
	}
	for dot := strings.IndexByte(host, '.'); dot >= 0; {
		if _, exists := suffixes[host[dot+1:]]; exists {
			return true
		}
		nextDot := strings.IndexByte(host[dot+1:], '.')
		if nextDot < 0 {
			break
		}
		dot += nextDot + 1
	}
	return false
}

func writeDomainDirectorySummary(directoryRules map[string][]string, path string) error {
	for dir, rules := range directoryRules {
		if len(rules) == 0 {
			continue
		}
		targetDir := filepath.Join(path, filepath.Dir(dir))
		name := filepath.Base(dir)
		if err := writeDomainRules(rules, targetDir, name); err != nil {
			return fmt.Errorf("write domain directory summary %q: %w", dir, err)
		}
	}
	return nil
}
