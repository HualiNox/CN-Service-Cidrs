package fetcher

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

type Result struct {
	IPPrefixes IPPrefixes
	SHA256     string
}

func Fetch(source parser.Source) (*Result, error) {
	if source.Type != parser.ClashList && source.Type != parser.SourceCIDR {
		log.Printf("warning: unsupported source type %q; skipping", source.Type)
		return &Result{}, nil
	}

	content, err := fetchSource(source.Value)
	if err != nil {
		return nil, err
	}

	var prefixes *IPPrefixes
	switch source.Type {
	case parser.ClashList:
		prefixes, err = parseClashList(source.Value, content)
	case parser.SourceCIDR:
		prefixes, err = parseSourceCIDR(source.Value, content)
	}
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	return &Result{IPPrefixes: *prefixes, SHA256: hex.EncodeToString(hash[:])}, nil
}

func fetchSource(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request source %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request source %q: unexpected HTTP status %s", url, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read source %q: %w", url, err)
	}
	return content, nil
}

func parseClashList(url string, content []byte) (*IPPrefixes, error) {
	var ipPrefixes IPPrefixes
	scanner := bufio.NewScanner(bytes.NewReader(content))
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

func parseSourceCIDR(url string, content []byte) (*IPPrefixes, error) {
	var ipPrefixes IPPrefixes
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			log.Printf("%s:%d: warning: invalid CIDR %q; skipping: %v", url, lineNo, line, err)
			continue
		}

		masked := prefix.Masked()
		if masked != prefix {
			log.Printf(
				"%s:%d: non-canonical CIDR %s, expected %s",
				url,
				lineNo,
				prefix,
				masked,
			)
			continue
		}

		if masked.Addr().Is4() {
			ipPrefixes.IPv4 = append(ipPrefixes.IPv4, masked)
		} else if masked.Addr().Is6() {
			ipPrefixes.IPv6 = append(ipPrefixes.IPv6, masked)
		} else {

		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan source %q after line %d: %w", url, lineNo, err)
	}

	return &ipPrefixes, nil
}
