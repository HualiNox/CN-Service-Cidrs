package generator

import (
	"net/netip"
	"reflect"
	"testing"
)

func TestMinimize(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "duplicate prefix",
			input: []string{"192.0.2.0/24", "192.0.2.0/24"},
			want:  []string{"192.0.2.0/24"},
		},
		{
			name:  "parent contains child",
			input: []string{"192.0.2.128/25", "192.0.2.0/24"},
			want:  []string{"192.0.2.0/24"},
		},
		{
			name:  "two /25 prefixes merge to /24",
			input: []string{"192.0.2.128/25", "192.0.2.0/25"},
			want:  []string{"192.0.2.0/24"},
		},
		{
			name:  "four /26 prefixes recursively merge to /24",
			input: []string{"192.0.2.192/26", "192.0.2.0/26", "192.0.2.128/26", "192.0.2.64/26"},
			want:  []string{"192.0.2.0/24"},
		},
		{
			name:  "adjacent differently sized prefixes do not merge",
			input: []string{"192.0.2.0/25", "192.0.2.128/26"},
			want:  []string{"192.0.2.0/25", "192.0.2.128/26"},
		},
		{
			name:  "IPv4 and IPv6 remain separate",
			input: []string{"::/0", "0.0.0.0/0"},
			want:  []string{"0.0.0.0/0", "::/0"},
		},
		{
			name:  "zero prefix",
			input: []string{"203.0.113.1/32", "0.0.0.0/0"},
			want:  []string{"0.0.0.0/0"},
		},
		{
			name:  "maximum boundary addresses",
			input: []string{"255.255.255.255/32", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
			want:  []string{"255.255.255.255/32", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
		},
		{
			name:  "output sorting is stable",
			input: []string{"2001:db8::/32", "192.0.2.128/25", "10.0.0.0/8", "192.0.2.0/25"},
			want:  []string{"10.0.0.0/8", "192.0.2.0/24", "2001:db8::/32"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := parsePrefixes(t, tt.input)
			want := parsePrefixes(t, tt.want)
			got := minimize(input)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("minimize() = %v, want %v", got, want)
			}
		})
	}

	t.Run("output sorting does not depend on input order", func(t *testing.T) {
		first := parsePrefixes(t, []string{"2001:db8::/32", "192.0.2.128/25", "10.0.0.0/8", "192.0.2.0/25"})
		second := parsePrefixes(t, []string{"192.0.2.0/25", "10.0.0.0/8", "192.0.2.128/25", "2001:db8::/32"})

		if got, want := minimize(first), minimize(second); !reflect.DeepEqual(got, want) {
			t.Fatalf("minimize() varies with input order: %v != %v", got, want)
		}
	})
}

func parsePrefixes(t *testing.T, values []string) []netip.Prefix {
	t.Helper()

	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			t.Fatalf("parse prefix %q: %v", value, err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
