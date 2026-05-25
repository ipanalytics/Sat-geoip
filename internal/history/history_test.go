package history

import (
	"bytes"
	"testing"
	"time"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func TestApplyRecordsTracksNewChangedAndWithdrawn(t *testing.T) {
	asn1 := 14593
	asn2 := 45700
	previous := map[string]resolver.ResolvedPrefix{
		"1.1.1.0/24": {
			Prefix:    "1.1.1.0/24",
			Operator:  "starlink",
			BGPState:  "announced",
			OriginASN: &asn1,
			FirstSeen: "2026-05-20",
			LastSeen:  "2026-05-24",
		},
		"2.2.2.0/24": {
			Prefix:    "2.2.2.0/24",
			Operator:  "viasat",
			BGPState:  "announced",
			FirstSeen: "2026-05-21",
			LastSeen:  "2026-05-24",
		},
	}
	current := []resolver.ResolvedPrefix{
		{
			Prefix:    "1.1.1.0/24",
			Operator:  "starlink",
			BGPState:  "bgp_only",
			OriginASN: &asn2,
		},
		{
			Prefix:   "3.3.3.0/24",
			Operator: "oneweb",
			BGPState: "bgp_only",
		},
	}

	result := ApplyRecords(previous, current, time.Date(2026, 5, 25, 12, 30, 0, 0, time.UTC))
	if result.Summary.New != 1 || result.Summary.Changed != 1 || result.Summary.Withdrawn != 1 {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
	changed := findRecord(t, result.Records, "1.1.1.0/24")
	if changed.FirstSeen != "2026-05-20" || changed.LastSeen != "2026-05-25" {
		t.Fatalf("history dates not preserved: %#v", changed)
	}
	if changed.ChangeType != ChangeChanged || changed.PreviousBGPState != "announced" || changed.PreviousOriginASN == nil || *changed.PreviousOriginASN != asn1 {
		t.Fatalf("previous fields not set: %#v", changed)
	}
}

func TestWriteChangesJSONLEmptyIsNonZeroForReleaseAsset(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteChangesJSONL(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty change log asset must be non-zero for GitHub Releases")
	}
	if got := buf.String(); got != "\n" {
		t.Fatalf("unexpected empty change log payload %q", got)
	}
}

func findRecord(t *testing.T, records []resolver.ResolvedPrefix, prefix string) resolver.ResolvedPrefix {
	t.Helper()
	for _, record := range records {
		if record.Prefix == prefix {
			return record
		}
	}
	t.Fatalf("missing record %s", prefix)
	return resolver.ResolvedPrefix{}
}
