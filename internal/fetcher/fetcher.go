package fetcher

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

var httpClient = http.Client{
	Timeout: 30 * time.Second,
}

func Fetch(source parser.Source) (*IPPrefixes, error) {
	switch source.Type {
	case parser.ClashList:
		return fetchClashList(source.Value)
	default:
		log.Printf("warning: unsupported source type %q; skipping", source.Type)
		return &IPPrefixes{}, nil
	}
}

func fetchClashList(url string) (*IPPrefixes, error) {
	var ipPrefixes IPPrefixes

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request source %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request source %q: unexpected HTTP status %s", url, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 2 {
			continue
		}

		ruleType := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch ruleType {
		case "IP-CIDR", "IP-CIDR6":
			prefix, err := netip.ParsePrefix(value)
			if err != nil {
				return nil, fmt.Errorf("source %q line %d: invalid CIDR %q: %w", url, lineNo, value, err)
			}

			masked := prefix.Masked()
			if prefix != masked {
				log.Printf(
					"%s:%d: non-canonical CIDR %s, expected %s",
					url,
					lineNo,
					prefix,
					masked,
				)
				continue
			}

			switch ruleType {
			case "IP-CIDR":
				if !masked.Addr().Is4() {
					return nil, fmt.Errorf(
						"source %q line %d: IP-CIDR contains IPv6 prefix %s",
						url,
						lineNo,
						masked,
					)
				}
				ipPrefixes.IPv4 = append(ipPrefixes.IPv4, masked)

			case "IP-CIDR6":
				if !masked.Addr().Is6() {
					return nil, fmt.Errorf(
						"source %q line %d: IP-CIDR6 contains IPv4 prefix %s",
						url,
						lineNo,
						masked,
					)
				}
				ipPrefixes.IPv6 = append(ipPrefixes.IPv6, masked)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan source %q after line %d: %w", url, lineNo, err)
	}

	return &ipPrefixes, nil
}
