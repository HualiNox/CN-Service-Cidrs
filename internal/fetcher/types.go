package fetcher

import "net/netip"

type IPPrefixes struct {
	IPv4 []netip.Prefix
	IPv6 []netip.Prefix
}
