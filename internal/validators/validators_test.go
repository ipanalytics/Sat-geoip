package validators

import (
	"strings"
	"testing"
)

func TestParseRFC8805(t *testing.T) {
	rows, err := ParseRFC8805(strings.NewReader("# prefix/ip,country_code,region_code,city,postal\n14.1.64.0/24,PH,,Manila,\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Country != "PH" || rows[0].City != "Manila" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestParseStarlinkPoPs(t *testing.T) {
	rows, err := ParseStarlinkPoPs(strings.NewReader("14.1.64.0/24,mnlaphl1,mnl\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PoPCode != "mnlaphl1" || rows[0].PoPIATA != "mnl" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
