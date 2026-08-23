package domain

import "testing"

func TestNormalizeQuotaSnapshotPreservesUsedValueAndState(t *testing.T) {
	used := 333.68
	snapshot := NormalizeQuotaSnapshot(QuotaSnapshot{Limits: []QuotaLimit{{
		ID:        "on_demand",
		UsedValue: &used,
		State:     QuotaLimitActive,
	}}})
	if snapshot.Limits[0].UsedValue == nil || *snapshot.Limits[0].UsedValue != used {
		t.Fatalf("used value = %v", snapshot.Limits[0].UsedValue)
	}
	if snapshot.Limits[0].State != QuotaLimitActive {
		t.Fatalf("state = %q", snapshot.Limits[0].State)
	}
}

func TestQuotaLimitStatesAreStableWireValues(t *testing.T) {
	want := []QuotaLimitState{"active", "unlimited", "disabled", "unavailable"}
	got := []QuotaLimitState{QuotaLimitActive, QuotaLimitUnlimited, QuotaLimitDisabled, QuotaLimitUnavailable}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("state %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeQuotaSnapshotDefaultsAbsoluteNumericLimitToActive(t *testing.T) {
	used, total := 12.0, 20.0
	snapshot := NormalizeQuotaSnapshot(QuotaSnapshot{Limits: []QuotaLimit{{
		ID: "spend", UsedValue: &used, TotalValue: &total,
	}}})
	if snapshot.Limits[0].State != QuotaLimitActive {
		t.Fatalf("state = %q, want active", snapshot.Limits[0].State)
	}
}
