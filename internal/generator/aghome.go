package generator

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/idna"
)

const defaultAGHomeUpstreamDNS = "https://dns.alidns.com/dns-query https://doh.pub/dns-query"
const defaultTechnitiumUpstreamDNS = "https://dns.alidns.com/dns-query (223.5.5.5)"

const (
	domainUpstreamMaxDomainsPerRule = 100
	domainUpstreamMaxRuleBytes      = 4096
)

func writeAGHomeUpstream(rules []string, path, upstream string) error {
	return writeDomainUpstream("AGH", rules, path, upstream, defaultAGHomeUpstreamDNS)
}

func writeTechnitiumUpstream(rules []string, path, upstream string) error {
	return writeDomainUpstream("Technitium", rules, path, upstream, defaultTechnitiumUpstreamDNS)
}

func writeDomainUpstream(product string, rules []string, path, upstream, defaultUpstream string) error {
	upstream = strings.TrimSpace(upstream)
	if upstream == "" {
		upstream = defaultUpstream
	}
	if strings.ContainsAny(upstream, "\r\n") {
		return fmt.Errorf("%s upstream DNS must be a single line", product)
	}
	upstreams := strings.Fields(upstream)
	if len(upstreams) == 0 {
		upstream = defaultUpstream
	} else {
		upstream = strings.Join(upstreams, " ")
	}

	var normalizedRules []string
	for _, rule := range rules {
		kind, value, ok := strings.Cut(rule, ":")
		if ok && kind == "domain" {
			ascii, err := idna.Lookup.ToASCII(value)
			if err != nil {
				return fmt.Errorf("convert domain %q to IDNA ASCII: %w", value, err)
			}
			ascii = strings.TrimSuffix(strings.ToLower(ascii), ".")
			if ascii == "" {
				return fmt.Errorf("convert domain %q to IDNA ASCII: empty result", value)
			}
			normalizedRules = append(normalizedRules, "domain:"+ascii)
		}
	}
	normalizedRules = minimizeDomainRules(normalizedRules)
	domains := make([]string, 0, len(normalizedRules))
	for _, rule := range normalizedRules {
		_, value, _ := strings.Cut(rule, ":")
		domains = append(domains, value)
	}
	sort.Strings(domains)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create %s upstream file %q: %w", product, path, err)
	}
	defer func() { _ = file.Close() }()

	baseLength := len("[/") + len("/]") + len(upstream)
	if baseLength >= domainUpstreamMaxRuleBytes {
		return fmt.Errorf("%s upstream DNS config is too long (%d bytes)", product, len(upstream))
	}

	group := make([]string, 0, domainUpstreamMaxDomainsPerRule)
	groupLength := baseLength
	groupCount := 0
	writeGroup := func() error {
		if len(group) == 0 {
			return nil
		}
		line := "[/" + strings.Join(group, "/") + "/]" + upstream + "\n"
		if _, err := io.WriteString(file, line); err != nil {
			return fmt.Errorf("write %s upstream file %q: %w", product, path, err)
		}
		groupCount++
		group = group[:0]
		groupLength = baseLength
		return nil
	}

	for _, domain := range domains {
		addedLength := len(domain)
		if len(group) > 0 {
			addedLength++ // separator slash
		}
		if len(group) >= domainUpstreamMaxDomainsPerRule || groupLength+addedLength > domainUpstreamMaxRuleBytes {
			if err := writeGroup(); err != nil {
				return err
			}
			addedLength = len(domain)
		}
		if groupLength+addedLength > domainUpstreamMaxRuleBytes {
			return fmt.Errorf("domain %q is too long for a %s upstream rule", domain, product)
		}
		group = append(group, domain)
		groupLength += addedLength
	}
	if err := writeGroup(); err != nil {
		return err
	}
	log.Printf("%s upstream: wrote %d domain suffixes in %d rules using %q to %q", product, len(domains), groupCount, upstream, path)
	return nil
}
