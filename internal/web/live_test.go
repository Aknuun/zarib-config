package web

import (
	"testing"

	"github.com/Aknuun/zarib-config/internal/ratio"
)

func TestAddFailSeparatesExpectedErrors(t *testing.T) {
	var bucket statBucket
	bucket.addFail("subscription_expired")
	bucket.addFail("probe_failed")
	bucket.addFail("xray_incompatible")
	bucket.addFail("fetch_failed")

	if bucket.Failed.Ignored != 2 {
		t.Fatalf("ignored = %d, want 2", bucket.Failed.Ignored)
	}
	if bucket.Failed.Total != 2 {
		t.Fatalf("total = %d, want 2", bucket.Failed.Total)
	}
	if bucket.Failed.Connect != 1 {
		t.Fatalf("connect = %d, want 1", bucket.Failed.Connect)
	}
	if bucket.Failed.Fetch != 1 {
		t.Fatalf("fetch = %d, want 1", bucket.Failed.Fetch)
	}
}

func TestStatClaimedAverageTreatsMissingClaimAsOne(t *testing.T) {
	claim := 1.2
	var bucket statBucket
	bucket.addResult(&ratio.Result{Ratio: 1.5}, nil)
	bucket.addResult(&ratio.Result{Ratio: 1.0, ClaimedRatio: &claim}, nil)

	view := statView(bucket, 0)
	if got := view["claimed_tests"].(int); got != 2 {
		t.Fatalf("claimed_tests = %d, want 2", got)
	}
	if got := view["claimed_avg"].(float64); got < 1.099 || got > 1.101 {
		t.Fatalf("claimed_avg = %v, want 1.1", got)
	}
	if got := view["claimed_real_avg"].(float64); got < 1.249 || got > 1.251 {
		t.Fatalf("claimed_real_avg = %v, want 1.25", got)
	}
	if got := view["claimed_diff_pct"].(float64); got < 13.63 || got > 13.64 {
		t.Fatalf("claimed_diff_pct = %v, want 13.636", got)
	}
}
