package web

import "testing"

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
