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
	IPPrefixes        IPPrefixes
	RejectedCIDRCount int
	SHA256            string
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
	var rejectedCIDRCount int
	switch source.Type {
	case parser.ClashList:
		prefixes, rejectedCIDRCount, err = parseClashList(source.Value, content)
	case parser.SourceCIDR:
		prefixes, rejectedCIDRCount, err = parseSourceCIDR(source.Value, content)
	}
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(content)
	return &Result{
		IPPrefixes:        *prefixes,
		RejectedCIDRCount: rejectedCIDRCount,
		SHA256:            hex.EncodeToString(hash[:]),
	}, nil
}

func fetchSource(url string) ([]byte, error) {
	retryDelays := [...]time.Duration{time.Second, 3 * time.Second, 8 * time.Second}
	var lastErr error
	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request for source %q: %w", url, err)
		}
		req.Header.Set("User-Agent", "cn-service-cidrs (+https://github.com/HualiNox/cn-service-cidrs)")

		resp, err := httpClient.Do(req)
		retryable := false
		if err != nil {
			lastErr = err
			retryable = true
		} else if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected HTTP status %s", resp.Status)
			retryable = retryableStatus(resp.StatusCode)
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
			_ = resp.Body.Close()
		} else {
			content, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr == nil {
				return content, nil
			}
			lastErr = fmt.Errorf("read response body: %w", readErr)
			retryable = true
		}

		if !retryable {
			return nil, fmt.Errorf("request source %q: %w", url, lastErr)
		}
		if attempt == len(retryDelays) {
			return nil, fmt.Errorf("request source %q failed after %d attempts: %w", url, attempt+1, lastErr)
		}

		delay := retryDelays[attempt]
		log.Printf(
			"warning: request source %q failed on attempt %d/%d: %v; retrying in %s",
			url,
			attempt+1,
			len(retryDelays)+1,
			lastErr,
			delay,
		)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("request source %q failed: %w", url, lastErr)
}

func retryableStatus(statusCode int) bool {
	return statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests ||
		statusCode >= http.StatusInternalServerError
}

func parseClashList(url string, content []byte) (*IPPrefixes, int, error) {
	var ipPrefixes IPPrefixes
	rejectedCIDRCount := 0
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
				return nil, rejectedCIDRCount, fmt.Errorf("source %q line %d: invalid CIDR %q: %w", url, lineNo, value, err)
			}

			masked := prefix.Masked()
			if prefix != masked {
				rejectedCIDRCount++
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
					return nil, rejectedCIDRCount, fmt.Errorf(
						"source %q line %d: IP-CIDR contains IPv6 prefix %s",
						url,
						lineNo,
						masked,
					)
				}
				ipPrefixes.IPv4 = append(ipPrefixes.IPv4, masked)

			case "IP-CIDR6":
				if !masked.Addr().Is6() {
					return nil, rejectedCIDRCount, fmt.Errorf(
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
		return nil, rejectedCIDRCount, fmt.Errorf("scan source %q after line %d: %w", url, lineNo, err)
	}

	return &ipPrefixes, rejectedCIDRCount, nil
}

func parseSourceCIDR(url string, content []byte) (*IPPrefixes, int, error) {
	var ipPrefixes IPPrefixes
	rejectedCIDRCount := 0
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
			rejectedCIDRCount++
			log.Printf("%s:%d: warning: invalid CIDR %q; skipping: %v", url, lineNo, line, err)
			continue
		}

		masked := prefix.Masked()
		if masked != prefix {
			rejectedCIDRCount++
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
		return nil, rejectedCIDRCount, fmt.Errorf("scan source %q after line %d: %w", url, lineNo, err)
	}

	return &ipPrefixes, rejectedCIDRCount, nil
}
