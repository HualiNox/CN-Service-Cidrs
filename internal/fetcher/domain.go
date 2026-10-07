package fetcher

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

var domainRulePrefixes = map[string]string{
	"DOMAIN-SUFFIX":  "domain",
	"DOMAIN":         "full",
	"DOMAIN-REGEX":   "regexp",
	"DOMAIN-KEYWORD": "keyword",
}

var domainCommunityPrefixes = map[string]string{
	"domain":  "domain",
	"full":    "full",
	"keyword": "keyword",
	"regexp":  "regexp",
}

func parseDomainList(source string, content []byte) ([]string, error) {
	source = strings.ToLower(source)
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	rules := make([]string, 0)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.Contains(source, "dnsmasq-china-list") {
			if domain, ok := parseDNSMasqDomain(line); ok {
				rules = append(rules, "domain:"+domain)
			}
			continue
		}

		if strings.Contains(source, "ios_rule_script") && !strings.Contains(line, ",") {
			// Surge's ChinaMax_Domain list uses a leading dot for suffix
			// matching; unprefixed domains are exact matches.
			if strings.HasPrefix(line, ".") {
				if value := strings.TrimPrefix(line, "."); value != "" {
					rules = append(rules, "domain:"+value)
				}
			} else if strings.Contains(line, ".") {
				rules = append(rules, "full:"+line)
			}
			continue
		}

		if kind, value, ok := parseDomainCommunityRule(line); ok {
			rules = append(rules, kind+":"+value)
			continue
		}

		reader := csv.NewReader(strings.NewReader(line))
		reader.FieldsPerRecord = -1
		fields, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("source %q line %d: parse domain rule: %w", source, lineNo, err)
		}
		if len(fields) == 1 {
			value := strings.TrimSpace(fields[0])
			if value == "" {
				return nil, fmt.Errorf("source %q line %d: empty domain", source, lineNo)
			}
			if strings.HasPrefix(strings.ToUpper(value), "DOMAIN-") || strings.EqualFold(value, "DOMAIN") {
				return nil, fmt.Errorf("source %q line %d: missing value for domain rule %q", source, lineNo, value)
			}
			if !strings.Contains(value, ".") {
				// Ignore standalone non-domain directives such as MATCH in a
				// mixed Clash list; public bare domains contain a dot.
				continue
			}
			rules = append(rules, "domain:"+value)
			continue
		}

		ruleType := strings.ToUpper(strings.TrimSpace(fields[0]))
		prefix, isDomainRule := domainRulePrefixes[ruleType]
		if !isDomainRule {
			if strings.HasPrefix(ruleType, "DOMAIN-") || ruleType == "DOMAIN" {
				return nil, fmt.Errorf("source %q line %d: unsupported domain rule type %q", source, lineNo, ruleType)
			}
			// Domain sources may point at a mixed Clash rule list. Ignore IP
			// and other non-domain rules while extracting the domain entries.
			continue
		}
		valueFields := fields[1:]
		if ruleType == "DOMAIN-REGEX" && len(valueFields) > 1 {
			// Regexes may contain commas. In Clash rule lists, the final
			// column is the optional policy; join the preceding columns back
			// into the regex so its expression is preserved.
			valueFields = valueFields[:len(valueFields)-1]
		} else {
			valueFields = valueFields[:1]
		}
		value := strings.TrimSpace(strings.Join(valueFields, ","))
		if value == "" {
			return nil, fmt.Errorf("source %q line %d: empty value for %s", source, lineNo, ruleType)
		}
		rules = append(rules, prefix+":"+value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read domain list %q: %w", source, err)
	}
	return rules, nil
}

func parseDomainCommunityRule(line string) (kind, value string, ok bool) {
	colon := strings.IndexByte(line, ':')
	if colon < 0 {
		return "", "", false
	}

	kind = strings.ToLower(strings.TrimSpace(line[:colon]))
	prefix, supported := domainCommunityPrefixes[kind]
	if !supported {
		return "", "", false
	}

	value = strings.TrimSpace(line[colon+1:])
	if attr := strings.Index(value, ":@"); attr >= 0 {
		value = strings.TrimSpace(value[:attr])
	}
	if value == "" {
		return "", "", false
	}
	return prefix, value, true
}

func parseDNSMasqDomain(line string) (string, bool) {
	if !strings.HasPrefix(line, "server=/") && !strings.HasPrefix(line, "address=/") {
		return "", false
	}
	value := strings.TrimPrefix(strings.TrimPrefix(line, "server=/"), "address=/")
	separator := strings.IndexByte(value, '/')
	if separator <= 0 {
		return "", false
	}
	domain := strings.TrimSuffix(value[:separator], ".")
	return domain, domain != ""
}
