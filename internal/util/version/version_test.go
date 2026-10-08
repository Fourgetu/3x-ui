package version

import "testing"

func TestCompareFourgetuRevisions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"v3.8.5-fourgetu.1", "3.9.0-fourgetu.1", -1},
		{"v3.9.0-fourgetu.1", "3.9.0-fourgetu.1", 0},
		{"v3.9.0-fourgetu.10", "3.9.0-fourgetu.2", 1},
		{"v3.9.0-fourgetu.1", "3.8.0", 1},
		{"v3.7.0-fourgetu.4", "3.8.0", -1},
	} {
		if got, ok := Compare(tc.a, tc.b); !ok || got != tc.want {
			t.Fatalf("Compare(%q, %q) = %d, %v; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
}

func TestCompareRejectsUnexpectedFormats(t *testing.T) {
	if _, ok := Compare("latest", "2.9.3"); ok {
		t.Fatal("expected non-semver latest tag to be rejected")
	}
	if _, ok := Compare("v2.9", "2.9.3"); ok {
		t.Fatal("expected short version to be rejected")
	}
}
