package fetcher

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/netip"
	"strings"
)

func parseCountryCSV(url string, content []byte, countryCode string) (*IPPrefixes, int, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("read country CSV header from %q: %w", url, err)
	}
	columns := make(map[string]int, len(header))
	for i, column := range header {
		column = strings.TrimSpace(strings.TrimPrefix(column, "\ufeff"))
		columns[column] = i
	}

	startColumn, hasStart := columns["ip_range_start"]
	endColumn, hasEnd := columns["ip_range_end"]
	countryColumn, hasCountry := columns["country_code"]
	if !hasStart || !hasEnd || !hasCountry {
		return nil, 0, fmt.Errorf("country CSV %q must contain ip_range_start, ip_range_end, and country_code columns", url)
	}

	var ipPrefixes IPPrefixes
	rejectedCIDRCount := 0
	rowNo := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNo++
		if err != nil {
			return nil, rejectedCIDRCount, fmt.Errorf("read country CSV %q row %d: %w", url, rowNo, err)
		}
		maxColumn := max(startColumn, endColumn, countryColumn)
		if len(row) <= maxColumn {
			return nil, rejectedCIDRCount, fmt.Errorf("country CSV %q row %d has %d columns, expected at least %d", url, rowNo, len(row), maxColumn+1)
		}
		if !strings.EqualFold(strings.TrimSpace(row[countryColumn]), strings.TrimSpace(countryCode)) {
			continue
		}

		start, startErr := netip.ParseAddr(strings.TrimSpace(row[startColumn]))
		end, endErr := netip.ParseAddr(strings.TrimSpace(row[endColumn]))
		if startErr != nil || endErr != nil || start.Is4In6() || end.Is4In6() || start.Is4() != end.Is4() {
			rejectedCIDRCount++
			log.Printf("%s: warning: invalid IP range in country CSV row %d; skipping", url, rowNo)
			continue
		}
		if start.Compare(end) > 0 {
			rejectedCIDRCount++
			log.Printf("%s: warning: reversed IP range in country CSV row %d; skipping", url, rowNo)
			continue
		}

		prefixes := rangeToPrefixes(start, end)
		if start.Is4() {
			ipPrefixes.IPv4 = append(ipPrefixes.IPv4, prefixes...)
		} else {
			ipPrefixes.IPv6 = append(ipPrefixes.IPv6, prefixes...)
		}
	}

	return &ipPrefixes, rejectedCIDRCount, nil
}

func rangeToPrefixes(start, end netip.Addr) []netip.Prefix {
	width := 128
	var startBytes, endBytes []byte
	if start.Is4() {
		width = 32
		v4Start := start.As4()
		v4End := end.As4()
		startBytes = v4Start[:]
		endBytes = v4End[:]
	} else {
		v6Start := start.As16()
		v6End := end.As16()
		startBytes = v6Start[:]
		endBytes = v6End[:]
	}

	current := new(big.Int).SetBytes(startBytes)
	last := new(big.Int).SetBytes(endBytes)
	var prefixes []netip.Prefix
	for current.Cmp(last) <= 0 {
		alignmentBits := 0
		if current.Sign() == 0 {
			alignmentBits = width
		} else {
			alignmentBits = min(int(current.TrailingZeroBits()), width)
		}
		remaining := new(big.Int).Sub(last, current)
		remaining.Add(remaining, big.NewInt(1))
		blockHostBits := min(alignmentBits, remaining.BitLen()-1)

		addressBytes := make([]byte, width/8)
		currentBytes := current.Bytes()
		copy(addressBytes[len(addressBytes)-len(currentBytes):], currentBytes)
		var addr netip.Addr
		if width == 32 {
			var bytes [4]byte
			copy(bytes[:], addressBytes)
			addr = netip.AddrFrom4(bytes)
		} else {
			var bytes [16]byte
			copy(bytes[:], addressBytes)
			addr = netip.AddrFrom16(bytes)
		}
		prefixes = append(prefixes, netip.PrefixFrom(addr, width-blockHostBits))

		blockSize := new(big.Int).Lsh(big.NewInt(1), uint(blockHostBits))
		current.Add(current, blockSize)
	}
	return prefixes
}
