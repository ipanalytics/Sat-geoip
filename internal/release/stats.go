package release

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

type Stats struct {
	TotalPrefixes       int            `json:"total_prefixes"`
	AnnouncedPrefixes   int            `json:"announced_prefixes"`
	GeoFeedOnly         int            `json:"geofeed_only_prefixes"`
	BGPOnly             int            `json:"bgp_only_prefixes"`
	WithPoP             int            `json:"prefixes_with_pop"`
	Operators           map[string]int `json:"operators"`
	OrbitClasses        map[string]int `json:"orbit_classes"`
	QualityFlags        map[string]int `json:"quality_flags"`
	GroundStationClaims int            `json:"ground_station_claims"`
	ActiveUserClaims    int            `json:"active_user_claims"`
}

func ComputeStats(records []resolver.ResolvedPrefix) Stats {
	stats := Stats{
		TotalPrefixes: len(records),
		Operators:     map[string]int{},
		OrbitClasses:  map[string]int{},
		QualityFlags:  map[string]int{},
	}
	for _, record := range records {
		stats.Operators[record.Operator]++
		stats.OrbitClasses[record.OrbitClass]++
		if record.ActiveUserClaim {
			stats.AnnouncedPrefixes++
		}
		if record.BGPState == string(resolver.BGPOnly) {
			stats.BGPOnly++
		}
		if hasQualityFlag(record, resolver.FlagPrefixOnlyInGeoFeed) {
			stats.GeoFeedOnly++
		}
		if record.PoPCode != "" {
			stats.WithPoP++
		}
		if record.GroundStationClaim {
			stats.GroundStationClaims++
		}
		if record.ActiveUserClaim {
			stats.ActiveUserClaims++
		}
		for _, flag := range record.QualityFlags {
			stats.QualityFlags[flag]++
		}
	}
	return stats
}

func WriteStatsJSON(w io.Writer, stats Stats) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}

func WriteStatsMarkdown(w io.Writer, stats Stats) error {
	_, err := fmt.Fprintf(w, `# sat-geoip Dataset Release

Generated artifacts for the current sat-geoip dataset build.

| Metric | Count |
|---|---:|
| Prefixes | %d |
| Announced prefixes | %d |
| GeoFeed-only prefixes | %d |
| BGP-only prefixes | %d |
| Prefixes with PoP assignment | %d |
| Active user claims | %d |
| Ground station claims | %d |

## Operators

%s

## Orbit Classes

%s

## Quality Flags

%s

`, stats.TotalPrefixes, stats.AnnouncedPrefixes, stats.GeoFeedOnly, stats.BGPOnly, stats.WithPoP, stats.ActiveUserClaims, stats.GroundStationClaims, markdownCountTable(stats.Operators), markdownCountTable(stats.OrbitClasses), markdownCountTable(stats.QualityFlags))
	return err
}

func READMEStatsBlock(stats Stats) string {
	var b strings.Builder
	b.WriteString("<!-- SAT_GEOIP_STATS_START -->\n")
	b.WriteString("| Dataset metric | Count |\n")
	b.WriteString("|---|---:|\n")
	fmt.Fprintf(&b, "| Prefixes | %d |\n", stats.TotalPrefixes)
	fmt.Fprintf(&b, "| Announced prefixes | %d |\n", stats.AnnouncedPrefixes)
	fmt.Fprintf(&b, "| GeoFeed-only prefixes | %d |\n", stats.GeoFeedOnly)
	fmt.Fprintf(&b, "| BGP-only prefixes | %d |\n", stats.BGPOnly)
	fmt.Fprintf(&b, "| Prefixes with PoP assignment | %d |\n", stats.WithPoP)
	fmt.Fprintf(&b, "| Ground station claims | %d |\n", stats.GroundStationClaims)
	b.WriteString("\n### Operators\n\n")
	b.WriteString(markdownCountTable(stats.Operators))
	b.WriteString("\n### Orbit Classes\n\n")
	b.WriteString(markdownCountTable(stats.OrbitClasses))
	b.WriteString("<!-- SAT_GEOIP_STATS_END -->")
	return b.String()
}

func UpdateREADMEStats(path string, stats Stats) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	start := "<!-- SAT_GEOIP_STATS_START -->"
	end := "<!-- SAT_GEOIP_STATS_END -->"
	text := string(content)
	startIdx := strings.Index(text, start)
	endIdx := strings.Index(text, end)
	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		return fmt.Errorf("README stats markers not found")
	}
	endIdx += len(end)
	next := text[:startIdx] + READMEStatsBlock(stats) + text[endIdx:]
	return os.WriteFile(path, []byte(next), 0o644)
}

func markdownCountTable(counts map[string]int) string {
	if len(counts) == 0 {
		return "_No records._\n"
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("| Name | Count |\n")
	b.WriteString("|---|---:|\n")
	for _, key := range keys {
		fmt.Fprintf(&b, "| `%s` | %d |\n", key, counts[key])
	}
	return b.String()
}

func hasQualityFlag(record resolver.ResolvedPrefix, flag string) bool {
	for _, got := range record.QualityFlags {
		if got == flag {
			return true
		}
	}
	return false
}
