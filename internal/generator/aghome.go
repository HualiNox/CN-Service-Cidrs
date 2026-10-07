package generator

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

const defaultAGHomeUpstreamDNS = "https://dns.alidns.com/dns-query https://doh.pub/dns-query"

const (
	agHomeMaxDomainsPerRule = 100
	agHomeMaxRuleBytes      = 4096
)

func writeAGHomeUpstream(rules []string, path, upstream string) error {
	upstream = strings.TrimSpace(upstream)
	if upstream == "" {
		upstream = defaultAGHomeUpstreamDNS
	}
	if strings.ContainsAny(upstream, "\r\n") {
		return fmt.Errorf("AGH upstream DNS must be a single line")
	}
	upstreams := strings.Fields(upstream)
	if len(upstreams) == 0 {
		upstream = defaultAGHomeUpstreamDNS
	} else {
		upstream = strings.Join(upstreams, " ")
	}

	rules = minimizeDomainRules(rules)
	domains := make([]string, 0, len(rules))
	for _, rule := range rules {
		kind, value, ok := strings.Cut(rule, ":")
		if ok && kind == "domain" {
			domains = append(domains, value)
		}
	}
	sort.Strings(domains)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create AGH upstream file %q: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	baseLength := len("[/") + len("/]") + len(upstream)
	if baseLength >= agHomeMaxRuleBytes {
		return fmt.Errorf("AGH upstream DNS config is too long (%d bytes)", len(upstream))
	}

	group := make([]string, 0, agHomeMaxDomainsPerRule)
	groupLength := baseLength
	groupCount := 0
	writeGroup := func() error {
		if len(group) == 0 {
			return nil
		}
		line := "[/" + strings.Join(group, "/") + "/]" + upstream + "\n"
		if _, err := io.WriteString(file, line); err != nil {
			return fmt.Errorf("write AGH upstream file %q: %w", path, err)
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
		if len(group) >= agHomeMaxDomainsPerRule || groupLength+addedLength > agHomeMaxRuleBytes {
			if err := writeGroup(); err != nil {
				return err
			}
			addedLength = len(domain)
		}
		if groupLength+addedLength > agHomeMaxRuleBytes {
			return fmt.Errorf("domain %q is too long for an AGH upstream rule", domain)
		}
		group = append(group, domain)
		groupLength += addedLength
	}
	if err := writeGroup(); err != nil {
		return err
	}
	log.Printf("AGH upstream: wrote %d domain suffixes in %d rules using %q to %q", len(domains), groupCount, upstream, path)
	return nil
}
