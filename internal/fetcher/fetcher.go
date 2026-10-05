package fetcher

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

var httpClient = http.Client{
	Timeout: 30 * time.Second,
}

func Fetch(output string, sourceFiles []parser.SourceFile) error {

	if err := mkdirAll(output, true); err != nil {
		return fmt.Errorf("prepare output directory %q: %w", output, err)
	}

	tables := filepath.Join(output, "tables")
	if err := os.MkdirAll(tables, 0o755); err != nil {
		return fmt.Errorf("create tables directory %q: %w", tables, err)
	}

	for _, sourceFile := range sourceFiles {
		sources := sourceFile.Group.Sources

		var ipPrefixes IPPrefixes
		for _, source := range sources {
			var sourcePrefixes *IPPrefixes

			switch source.Type {
			case parser.ClashList:
				var err error
				sourcePrefixes, err = fetchClashList(source.Value)
				if err != nil {
					return fmt.Errorf("fetch source %q in group %q: %w", source.Value, sourceFile.Group.Name, err)
				}
			default:
				log.Printf("warning: unsupported source type %q; skipping", source.Type)
				continue
			}

			ipPrefixes.IPv4 = append(ipPrefixes.IPv4, sourcePrefixes.IPv4...)
			ipPrefixes.IPv6 = append(ipPrefixes.IPv6, sourcePrefixes.IPv6...)
		}

		ipPrefixes.IPv4 = dedup(ipPrefixes.IPv4)
		ipPrefixes.IPv6 = dedup(ipPrefixes.IPv6)

		if err := writeIPCIDRs(
			&ipPrefixes,
			filepath.Join(tables, sourceFile.Directory),
			sourceFile.Group.Name,
		); err != nil {
			return fmt.Errorf("write group %q: %w", sourceFile.Group.Name, err)
		}
	}

	return nil
}

func mkdirAll(path string, removeExisting bool) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("inspect directory %q: %w", path, err)
		}
	} else if removeExisting {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove directory %q: %w", path, err)
		}
	} else {
		return nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", path, err)
	}
	return nil
}

func dedup(prefixes []netip.Prefix) []netip.Prefix {
	set := make(map[netip.Prefix]struct{}, len(prefixes))

	for _, prefix := range prefixes {
		set[prefix.Masked()] = struct{}{}
	}

	result := make([]netip.Prefix, 0, len(set))

	for prefix := range set {
		result = append(result, prefix)
	}

	return result
}

func writeIPCIDRs(ipPrefixes *IPPrefixes, path, name string) error {
	if len(ipPrefixes.IPv4) == 0 && len(ipPrefixes.IPv6) == 0 {
		log.Printf("no IP prefixes found for %q; skipping output", filepath.Join(path, name))
		return nil
	}

	if err := mkdirAll(path, false); err != nil {
		return fmt.Errorf("prepare output directory %q: %w", path, err)
	}

	ipv4Name := name + "-ipv4.txt"
	ipv6Name := name + "-ipv6.txt"
	allIPName := name + ".txt"

	allIPFile, err := os.OpenFile(
		filepath.Join(path, allIPName),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", filepath.Join(path, allIPName), err)
	}
	defer func() {
		_ = allIPFile.Close()
	}()

	if len(ipPrefixes.IPv4) != 0 {
		ipv4File, err := os.OpenFile(
			filepath.Join(path, ipv4Name),
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", filepath.Join(path, ipv4Name), err)
		}
		defer func() {
			_ = ipv4File.Close()
		}()

		for _, ipv4 := range ipPrefixes.IPv4 {
			ipv4Str := ipv4.String() + "\n"

			_, err := io.WriteString(ipv4File, ipv4Str)
			if err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, ipv4Name), err)
			}

			_, err = io.WriteString(allIPFile, ipv4Str)
			if err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, allIPName), err)
			}
		}
	} else {
		log.Printf("no IPv4 prefixes found for %q; skipping output", filepath.Join(path, ipv4Name))
	}

	if len(ipPrefixes.IPv6) != 0 {
		ipv6File, err := os.OpenFile(
			filepath.Join(path, ipv6Name),
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", filepath.Join(path, ipv6Name), err)
		}
		defer func() {
			_ = ipv6File.Close()
		}()
		for _, ipv6 := range ipPrefixes.IPv6 {
			ipv6Str := ipv6.String() + "\n"

			_, err := io.WriteString(ipv6File, ipv6Str)
			if err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, ipv6Name), err)
			}

			_, err = io.WriteString(allIPFile, ipv6Str)
			if err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, allIPName), err)
			}
		}
	} else {
		log.Printf("no IPv6 prefixes found for %q; skipping output", filepath.Join(path, ipv6Name))
	}

	return nil
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
