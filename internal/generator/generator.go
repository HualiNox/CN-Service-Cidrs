package generator

import (
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"sort"

	"github.com/HualiNox/cn-service-cidrs/internal/fetcher"
	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

func Build(output string, sourceFiles []parser.SourceFile) error {
	if err := mkdirAll(output, true); err != nil {
		return fmt.Errorf("prepare output directory %q: %w", output, err)
	}

	tables := filepath.Join(output, "tables")
	if err := os.MkdirAll(tables, 0o755); err != nil {
		return fmt.Errorf("create tables directory %q: %w", tables, err)
	}

	directoryPrefixes := make(map[string]fetcher.IPPrefixes)
	for _, sourceFile := range sourceFiles {
		var ipPrefixes fetcher.IPPrefixes
		for _, source := range sourceFile.Group.Sources {
			sourcePrefixes, err := fetcher.Fetch(source)
			if err != nil {
				return fmt.Errorf("fetch source %q in group %q: %w", source.Value, sourceFile.Group.Name, err)
			}
			ipPrefixes.IPv4 = append(ipPrefixes.IPv4, sourcePrefixes.IPv4...)
			ipPrefixes.IPv6 = append(ipPrefixes.IPv6, sourcePrefixes.IPv6...)
		}

		ipPrefixes.IPv4 = dedup(ipPrefixes.IPv4)
		ipPrefixes.IPv6 = dedup(ipPrefixes.IPv6)

		log.Printf(
			"group %q: IPv4=%d, IPv6=%d",
			sourceFile.Group.Name,
			len(ipPrefixes.IPv4),
			len(ipPrefixes.IPv6),
		)
		if err := writeIPCIDRs(
			ipPrefixes,
			filepath.Join(tables, sourceFile.Directory),
			sourceFile.Group.Name,
		); err != nil {
			return fmt.Errorf("write group %q: %w", sourceFile.Group.Name, err)
		}

		for dir := sourceFile.Directory; dir != "."; dir = filepath.Dir(dir) {
			prefixes := directoryPrefixes[dir]
			prefixes.IPv4 = dedup(append(prefixes.IPv4, ipPrefixes.IPv4...))
			prefixes.IPv6 = dedup(append(prefixes.IPv6, ipPrefixes.IPv6...))
			directoryPrefixes[dir] = prefixes
		}
	}

	if err := writeDirectorySummary(directoryPrefixes, tables); err != nil {
		return err
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
	sort.Slice(result, func(i, j int) bool {
		if order := result[i].Addr().Compare(result[j].Addr()); order != 0 {
			return order < 0
		}
		return result[i].Bits() < result[j].Bits()
	})
	return result
}

func writeIPCIDRs(ipPrefixes fetcher.IPPrefixes, path, name string) error {
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
	defer func() { _ = allIPFile.Close() }()

	if len(ipPrefixes.IPv4) == 0 {
		log.Printf("no IPv4 prefixes found for %q; skipping output", filepath.Join(path, ipv4Name))
	} else {
		ipv4Path := filepath.Join(path, ipv4Name)
		ipv4File, err := os.OpenFile(
			ipv4Path,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", ipv4Path, err)
		}
		defer func() { _ = ipv4File.Close() }()

		for _, prefix := range ipPrefixes.IPv4 {
			line := prefix.String() + "\n"
			if _, err := io.WriteString(ipv4File, line); err != nil {
				return fmt.Errorf("write output file %q: %w", ipv4Path, err)
			}
			if _, err := io.WriteString(allIPFile, line); err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, allIPName), err)
			}
		}
	}

	if len(ipPrefixes.IPv6) == 0 {
		log.Printf("no IPv6 prefixes found for %q; skipping output", filepath.Join(path, ipv6Name))
	} else {
		ipv6Path := filepath.Join(path, ipv6Name)
		ipv6File, err := os.OpenFile(
			ipv6Path,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", ipv6Path, err)
		}
		defer func() { _ = ipv6File.Close() }()

		for _, prefix := range ipPrefixes.IPv6 {
			line := prefix.String() + "\n"
			if _, err := io.WriteString(ipv6File, line); err != nil {
				return fmt.Errorf("write output file %q: %w", ipv6Path, err)
			}
			if _, err := io.WriteString(allIPFile, line); err != nil {
				return fmt.Errorf("write output file %q: %w", filepath.Join(path, allIPName), err)
			}
		}
	}

	return nil
}

func writeDirectorySummary(directoryPrefixes map[string]fetcher.IPPrefixes, path string) error {
	for dir, ipPrefixes := range directoryPrefixes {
		if len(ipPrefixes.IPv4) == 0 && len(ipPrefixes.IPv6) == 0 {
			continue
		}

		targetDir := filepath.Join(path, filepath.Dir(dir))
		log.Printf(
			"directory %q: IPv4=%d, IPv6=%d",
			dir,
			len(ipPrefixes.IPv4),
			len(ipPrefixes.IPv6),
		)
		if err := writeIPCIDRs(ipPrefixes, targetDir, filepath.Base(dir)); err != nil {
			return fmt.Errorf("write directory summary %q: %w", dir, err)
		}
	}

	return nil
}
