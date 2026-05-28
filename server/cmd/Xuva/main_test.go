package main

import "testing"

// TestPortFromHTTPAddr covers the three shapes config.HTTPAddr can take:
// pure ":N" (Go's default unspecified-bind notation), explicit IPv4
// host:port, and an empty/garbage string. The firewall hookup skips when
// this returns 0, so the test mainly proves we don't accidentally short
// the firewall code path for the common shapes.
func TestPortFromHTTPAddr(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{":8097", 8097},
		{"0.0.0.0:8097", 8097},
		{"192.168.1.5:8097", 8097},
		{"[::]:8097", 8097},
		{"", 0},
		{"not-a-host-port", 0},
		{":notnumeric", 0},
	}
	for _, tc := range cases {
		if got := portFromHTTPAddr(tc.in); got != tc.want {
			t.Errorf("portFromHTTPAddr(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
