package generator

import (
	"math/big"
	"net/netip"
	"sort"
)

type prefixEvent struct {
	position *big.Int
	source   int
	delta    int
}

// exclusiveAddressCounts returns, for each source, the number of addresses
// covered by that source and no other source. Prefixes should be minimized
// within each source before calling this function.
func exclusiveAddressCounts(prefixesBySource [][]netip.Prefix, addressBits int) []string {
	counts := make([]string, len(prefixesBySource))
	for i := range counts {
		counts[i] = "0"
	}
	if len(prefixesBySource) == 0 {
		return counts
	}

	events := make([]prefixEvent, 0)
	for source, prefixes := range prefixesBySource {
		for _, prefix := range prefixes {
			prefix = prefix.Masked()
			addr := prefix.Addr()
			var raw []byte
			if addr.Is4() {
				v := addr.As4()
				raw = v[:]
			} else {
				v := addr.As16()
				raw = v[:]
			}
			start := new(big.Int).SetBytes(raw)
			size := new(big.Int).Lsh(big.NewInt(1), uint(addressBits-prefix.Bits()))
			end := new(big.Int).Add(new(big.Int).Set(start), size)
			events = append(events,
				prefixEvent{position: start, source: source, delta: 1},
				prefixEvent{position: end, source: source, delta: -1},
			)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].position.Cmp(events[j].position) < 0
	})

	totals := make([]*big.Int, len(prefixesBySource))
	active := make([]int, len(prefixesBySource))
	for i := range totals {
		totals[i] = new(big.Int)
	}
	var previous *big.Int
	for i := 0; i < len(events); {
		position := events[i].position
		if previous != nil {
			activeSource, activeSources := -1, 0
			for source, n := range active {
				if n > 0 {
					activeSource = source
					activeSources++
				}
			}
			if activeSources == 1 {
				segment := new(big.Int).Sub(position, previous)
				totals[activeSource].Add(totals[activeSource], segment)
			}
		}

		j := i
		for j < len(events) && events[j].position.Cmp(position) == 0 {
			active[events[j].source] += events[j].delta
			j++
		}
		previous = position
		i = j
	}
	for i := range totals {
		counts[i] = totals[i].String()
	}
	return counts
}
