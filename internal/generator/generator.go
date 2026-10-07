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

func Build(output string, sourceFiles []parser.SourceFile) ([]SourceStatus, error) {
	if err := mkdirAll(output, true); err != nil {
		return nil, fmt.Errorf("prepare output directory %q: %w", output, err)
	}

	tables := filepath.Join(output, "tables")
	if err := os.MkdirAll(tables, 0o755); err != nil {
		return nil, fmt.Errorf("create tables directory %q: %w", tables, err)
	}

	directoryPrefixes := make(map[string]fetcher.IPPrefixes)
	sources := make([]SourceStatus, 0)
	sourceIPv4 := make([][]netip.Prefix, 0)
	sourceIPv6 := make([][]netip.Prefix, 0)
	for _, sourceFile := range sourceFiles {
		var ipPrefixes fetcher.IPPrefixes
		for _, source := range sourceFile.Group.Sources {
			result, err := fetcher.Fetch(source)
			if err != nil {
				return nil, fmt.Errorf("fetch source %q in group %q: %w", source.Value, sourceFile.Group.Name, err)
			}

			group := sourceFile.Group.Name
			if sourceFile.Directory != "." {
				group = filepath.ToSlash(filepath.Join(sourceFile.Directory, group))
			}
			ipv4 := minimize(result.IPPrefixes.IPv4)
			ipv6 := minimize(result.IPPrefixes.IPv6)
			sources = append(sources, SourceStatus{
				Group:             group,
				Type:              source.Type,
				URL:               source.Value,
				SHA256:            result.SHA256,
				IPv4Count:         len(ipv4),
				IPv6Count:         len(ipv6),
				RejectedCIDRCount: result.RejectedCIDRCount,
			})
			sourceIPv4 = append(sourceIPv4, ipv4)
			sourceIPv6 = append(sourceIPv6, ipv6)

			ipPrefixes.IPv4 = append(ipPrefixes.IPv4, ipv4...)
			ipPrefixes.IPv6 = append(ipPrefixes.IPv6, ipv6...)
		}

		ipPrefixes.IPv4 = minimize(ipPrefixes.IPv4)
		ipPrefixes.IPv6 = minimize(ipPrefixes.IPv6)

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
			return nil, fmt.Errorf("write group %q: %w", sourceFile.Group.Name, err)
		}

		for dir := sourceFile.Directory; dir != "."; dir = filepath.Dir(dir) {
			prefixes := directoryPrefixes[dir]
			prefixes.IPv4 = minimize(append(prefixes.IPv4, ipPrefixes.IPv4...))
			prefixes.IPv6 = minimize(append(prefixes.IPv6, ipPrefixes.IPv6...))
			directoryPrefixes[dir] = prefixes
		}
	}

	if err := writeDirectorySummary(directoryPrefixes, tables); err != nil {
		return nil, err
	}

	v4Exclusive := exclusiveAddressCounts(sourceIPv4, 32)
	v6Exclusive := exclusiveAddressCounts(sourceIPv6, 128)
	for i := range sources {
		sources[i].ExclusiveIPv4AddressCount = v4Exclusive[i]
		sources[i].ExclusiveIPv6AddressCount = v6Exclusive[i]
	}

	return sources, nil
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

func minimize(prefixes []netip.Prefix) []netip.Prefix {
	var ipv4Root, ipv6Root *prefixNode
	for _, prefix := range prefixes {
		prefix = prefix.Masked()
		if prefix.Addr().Is4() {
			ipv4Root = insertPrefix(ipv4Root, prefix, 0)
		} else {
			ipv6Root = insertPrefix(ipv6Root, prefix, 0)
		}
	}

	result := make([]netip.Prefix, 0, len(prefixes))
	collectPrefixes(ipv4Root, make([]byte, 4), 0, true, &result)
	collectPrefixes(ipv6Root, make([]byte, 16), 0, false, &result)
	sort.Slice(result, func(i, j int) bool {
		if order := result[i].Addr().Compare(result[j].Addr()); order != 0 {
			return order < 0
		}
		return result[i].Bits() < result[j].Bits()
	})
	return result
}

type prefixNode struct {
	full     bool
	children [2]*prefixNode
}

func insertPrefix(node *prefixNode, prefix netip.Prefix, depth int) *prefixNode {
	if node == nil {
		node = &prefixNode{}
	}
	if node.full {
		return node
	}
	if depth == prefix.Bits() {
		node.full = true
		node.children = [2]*prefixNode{}
		return node
	}

	bit := addressBit(prefix.Addr(), depth)
	node.children[bit] = insertPrefix(node.children[bit], prefix, depth+1)
	if node.children[0] != nil && node.children[0].full && node.children[1] != nil && node.children[1].full {
		node.full = true
		node.children = [2]*prefixNode{}
	}
	return node
}

func addressBit(addr netip.Addr, bit int) int {
	if addr.Is4() {
		bytes := addr.As4()
		return int((bytes[bit/8] >> (7 - bit%8)) & 1)
	}
	bytes := addr.As16()
	return int((bytes[bit/8] >> (7 - bit%8)) & 1)
}

func collectPrefixes(node *prefixNode, address []byte, depth int, ipv4 bool, prefixes *[]netip.Prefix) {
	if node == nil {
		return
	}
	if node.full {
		var addr netip.Addr
		if ipv4 {
			var bytes [4]byte
			copy(bytes[:], address)
			addr = netip.AddrFrom4(bytes)
		} else {
			var bytes [16]byte
			copy(bytes[:], address)
			addr = netip.AddrFrom16(bytes)
		}
		*prefixes = append(*prefixes, netip.PrefixFrom(addr, depth))
		return
	}

	for bit, child := range node.children {
		if child == nil {
			continue
		}
		if bit == 1 {
			address[depth/8] |= 1 << (7 - depth%8)
		}
		collectPrefixes(child, address, depth+1, ipv4, prefixes)
		if bit == 1 {
			address[depth/8] &^= 1 << (7 - depth%8)
		}
	}
}

func writeIPCIDRs(ipPrefixes fetcher.IPPrefixes, path, name string) error {
	for _, prefix := range ipPrefixes.IPv4 {
		if prefix.Bits() == 0 {
			return fmt.Errorf("refusing to write IPv4 default route %s to %q", prefix, filepath.Join(path, name))
		}
	}
	for _, prefix := range ipPrefixes.IPv6 {
		if prefix.Bits() == 0 {
			return fmt.Errorf("refusing to write IPv6 default route %s to %q", prefix, filepath.Join(path, name))
		}
	}

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
