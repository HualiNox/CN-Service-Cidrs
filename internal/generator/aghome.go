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

	for _, domain := range domains {
		if _, err := io.WriteString(file, "[/"+domain+"/]"+upstream+"\n"); err != nil {
			return fmt.Errorf("write AGH upstream file %q: %w", path, err)
		}
	}
	log.Printf("AGH upstream: wrote %d domain suffixes using %q to %q", len(domains), upstream, path)
	return nil
}
