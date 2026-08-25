package utils

import (
	"net/url"
	"strings"
	"sync"
	"testing"
)

func TestGetValidFloatAcceptsJSONNumbers(t *testing.T) {
	// encoding/json decodes numbers into float64, which used to panic here.
	cases := []struct {
		in   interface{}
		want float64
	}{
		{nil, 0},
		{float64(1.5), 1.5},
		{"100.00", 100},
		{int64(7), 7},
		{true, 0},
		{"not-a-number", 0},
	}
	for _, c := range cases {
		if got := GetValidFloat(c.in); got != c.want {
			t.Errorf("GetValidFloat(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestGetValidStringAcceptsNonStrings(t *testing.T) {
	if got := GetValidString(nil); got != "" {
		t.Errorf("nil => %q", got)
	}
	if got := GetValidString("abc"); got != "abc" {
		t.Errorf("string => %q", got)
	}
	if got := GetValidString(123); got != "123" {
		t.Errorf("int => %q", got)
	}
}

func TestEscapePathSegmentBlocksTraversal(t *testing.T) {
	got := EscapePathSegment("../../auth/token")
	if strings.Contains(got, "/") {
		t.Fatalf("escaped segment still contains a path separator: %q", got)
	}
	// The escaped value must round-trip back to the original id.
	back, err := url.PathUnescape(got)
	if err != nil || back != "../../auth/token" {
		t.Fatalf("round trip failed: %q / %v", back, err)
	}
}

func TestBuildCallbackURLPreservesExistingQuery(t *testing.T) {
	got := BuildCallbackURL("https://shop.mn/cb?order=7", map[string]string{"payment_id": "42"})
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("result is not a valid URL: %q (%v)", got, err)
	}
	if strings.Count(got, "?") != 1 {
		t.Fatalf("expected exactly one '?', got %q", got)
	}
	if u.Query().Get("order") != "7" {
		t.Errorf("pre-existing param lost: %q", got)
	}
	if u.Query().Get("payment_id") != "42" {
		t.Errorf("new param missing: %q", got)
	}
}

func TestBuildCallbackURLWithoutParams(t *testing.T) {
	// No params must not leave a dangling "?".
	if got := BuildCallbackURL("https://shop.mn/cb", nil); got != "https://shop.mn/cb" {
		t.Fatalf("got %q", got)
	}
}

func TestRandStringIsConcurrencySafeAndUnique(t *testing.T) {
	const workers, each = 16, 64

	var mu sync.Mutex
	seen := map[string]struct{}{}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < each; j++ {
				s := RandStringBytesMaskImprSrcSB(24)
				mu.Lock()
				seen[s] = struct{}{}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(seen) != workers*each {
		t.Fatalf("expected %d unique strings, got %d", workers*each, len(seen))
	}
	for s := range seen {
		if len(s) != 24 {
			t.Fatalf("unexpected length %d for %q", len(s), s)
		}
	}
}

func TestRandStringNonPositive(t *testing.T) {
	if got := RandStringBytesMaskImprSrcSB(0); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := RandStringBytesMaskImprSrcSB(-5); got != "" {
		t.Fatalf("got %q", got)
	}
}
