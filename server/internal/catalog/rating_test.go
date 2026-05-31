package catalog

import "testing"

func TestRatingWithinCeiling(t *testing.T) {
	cases := []struct {
		name        string
		itemRating  string
		ceiling     string
		wantAllowed bool
	}{
		{"no ceiling allows everything", "NC-17", "", true},
		{"equal rating allowed", "PG-13", "PG-13", true},
		{"below ceiling allowed", "G", "PG-13", true},
		{"above ceiling blocked", "R", "PG-13", false},
		{"NC-17 blocked under R", "NC-17", "R", false},
		{"tv rating below ceiling allowed", "TV-Y", "TV-14", true},
		{"tv rating above ceiling blocked", "TV-MA", "TV-14", false},
		{"unknown item rating passes through", "NR", "PG-13", true},
		{"unknown ceiling does not enforce", "R", "BOGUS", true},
		{"empty item rating passes through", "", "G", true},
		{"case-insensitive comparison", "pg-13", "PG-13", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RatingWithinCeiling(tc.itemRating, tc.ceiling); got != tc.wantAllowed {
				t.Fatalf("RatingWithinCeiling(%q, %q) = %v, want %v", tc.itemRating, tc.ceiling, got, tc.wantAllowed)
			}
		})
	}
}
