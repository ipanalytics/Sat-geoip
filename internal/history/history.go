package history

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

const (
	ChangeNew       = "new"
	ChangeUnchanged = "unchanged"
	ChangeChanged   = "changed"
	ChangeWithdrawn = "withdrawn"
)

type Change struct {
	Prefix               string `json:"prefix"`
	ChangeType           string `json:"change_type"`
	ChangedAt            string `json:"changed_at"`
	Operator             string `json:"operator,omitempty"`
	PreviousOperator     string `json:"previous_operator,omitempty"`
	BGPState             string `json:"bgp_state,omitempty"`
	PreviousBGPState     string `json:"previous_bgp_state,omitempty"`
	OriginASN            *int   `json:"origin_asn,omitempty"`
	PreviousOriginASN    *int   `json:"previous_origin_asn,omitempty"`
	PoPCode              string `json:"pop_code,omitempty"`
	PreviousPoPCode      string `json:"previous_pop_code,omitempty"`
	GeoIPCountry         string `json:"geoip_country,omitempty"`
	PreviousGeoIPCountry string `json:"previous_geoip_country,omitempty"`
	GeoIPRegion          string `json:"geoip_region,omitempty"`
	PreviousGeoIPRegion  string `json:"previous_geoip_region,omitempty"`
	GeoIPCity            string `json:"geoip_city,omitempty"`
	PreviousGeoIPCity    string `json:"previous_geoip_city,omitempty"`
}

type Summary struct {
	GeneratedAt     int64          `json:"generated_at"`
	GeneratedDate   string         `json:"generated_date"`
	PreviousRecords int            `json:"previous_records"`
	CurrentRecords  int            `json:"current_records"`
	New             int            `json:"new"`
	Changed         int            `json:"changed"`
	Withdrawn       int            `json:"withdrawn"`
	Unchanged       int            `json:"unchanged"`
	ChangesByType   map[string]int `json:"changes_by_type"`
}

type Result struct {
	Records []resolver.ResolvedPrefix
	Changes []Change
	Summary Summary
}

func Apply(previousPath string, current []resolver.ResolvedPrefix, now time.Time) (Result, error) {
	previous, err := ReadResolved(previousPath)
	if err != nil {
		return Result{}, err
	}
	return ApplyRecords(previous, current, now), nil
}

func ApplyRecords(previous map[string]resolver.ResolvedPrefix, current []resolver.ResolvedPrefix, now time.Time) Result {
	date := now.UTC().Format("2006-01-02")
	ts := now.UTC().Format(time.RFC3339)
	currentMap := make(map[string]resolver.ResolvedPrefix, len(current))
	changes := []Change{}

	for i := range current {
		record := current[i]
		prev, seen := previous[record.Prefix]
		if !seen {
			if record.FirstSeen == "" {
				record.FirstSeen = date
			}
			record.LastSeen = date
			record.ChangedAt = ts
			record.ChangeType = ChangeNew
			changes = append(changes, changeFrom(record, resolver.ResolvedPrefix{}, ChangeNew, ts))
		} else {
			record.FirstSeen = nonEmpty(prev.FirstSeen, date)
			record.LastSeen = date
			if materiallyChanged(prev, record) {
				record.ChangedAt = ts
				record.ChangeType = ChangeChanged
				copyPrevious(&record, prev)
				changes = append(changes, changeFrom(record, prev, ChangeChanged, ts))
			} else {
				record.ChangedAt = prev.ChangedAt
				record.ChangeType = ChangeUnchanged
			}
		}
		current[i] = record
		currentMap[record.Prefix] = record
	}

	for prefix, prev := range previous {
		if _, ok := currentMap[prefix]; ok {
			continue
		}
		withdrawn := prev
		withdrawn.LastSeen = date
		withdrawn.ChangedAt = ts
		withdrawn.ChangeType = ChangeWithdrawn
		withdrawn.PreviousBGPState = prev.BGPState
		withdrawn.BGPState = string(resolver.BGPWithdrawnRecently)
		withdrawn.ActiveUserClaim = false
		changes = append(changes, changeFrom(withdrawn, prev, ChangeWithdrawn, ts))
	}

	sort.Slice(current, func(i, j int) bool { return current[i].Prefix < current[j].Prefix })
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].ChangeType == changes[j].ChangeType {
			return changes[i].Prefix < changes[j].Prefix
		}
		return changes[i].ChangeType < changes[j].ChangeType
	})

	summary := Summary{
		GeneratedAt:     now.UTC().Unix(),
		GeneratedDate:   date,
		PreviousRecords: len(previous),
		CurrentRecords:  len(current),
		ChangesByType:   map[string]int{},
	}
	for _, change := range changes {
		summary.ChangesByType[change.ChangeType]++
	}
	summary.New = summary.ChangesByType[ChangeNew]
	summary.Changed = summary.ChangesByType[ChangeChanged]
	summary.Withdrawn = summary.ChangesByType[ChangeWithdrawn]
	summary.Unchanged = len(current) - summary.New - summary.Changed
	return Result{Records: current, Changes: changes, Summary: summary}
}

func ReadResolved(path string) (map[string]resolver.ResolvedPrefix, error) {
	records := map[string]resolver.ResolvedPrefix{}
	if path == "" {
		return records, nil
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return records, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 32*1024*1024)
	for scanner.Scan() {
		var record resolver.ResolvedPrefix
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		records[record.Prefix] = record
	}
	return records, scanner.Err()
}

func WriteChangesJSONL(w io.Writer, changes []Change) error {
	if len(changes) == 0 {
		_, err := io.WriteString(w, "\n")
		return err
	}
	enc := json.NewEncoder(w)
	for _, change := range changes {
		if err := enc.Encode(change); err != nil {
			return err
		}
	}
	return nil
}

func WriteChangesCSV(w io.Writer, changes []Change) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"prefix", "change_type", "changed_at", "operator", "previous_operator", "bgp_state", "previous_bgp_state", "origin_asn", "previous_origin_asn", "pop_code", "previous_pop_code", "geoip_country", "previous_geoip_country", "geoip_region", "previous_geoip_region", "geoip_city", "previous_geoip_city"}); err != nil {
		return err
	}
	for _, c := range changes {
		if err := cw.Write([]string{
			c.Prefix, c.ChangeType, c.ChangedAt,
			c.Operator, c.PreviousOperator,
			c.BGPState, c.PreviousBGPState,
			intPtr(c.OriginASN), intPtr(c.PreviousOriginASN),
			c.PoPCode, c.PreviousPoPCode,
			c.GeoIPCountry, c.PreviousGeoIPCountry,
			c.GeoIPRegion, c.PreviousGeoIPRegion,
			c.GeoIPCity, c.PreviousGeoIPCity,
		}); err != nil {
			return err
		}
	}
	return cw.Error()
}

func WriteSummaryJSON(w io.Writer, summary Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(summary)
}

func materiallyChanged(prev, curr resolver.ResolvedPrefix) bool {
	return prev.Operator != curr.Operator ||
		prev.BGPState != curr.BGPState ||
		ptrVal(prev.OriginASN) != ptrVal(curr.OriginASN) ||
		prev.GeoIPCountry != curr.GeoIPCountry ||
		prev.GeoIPRegion != curr.GeoIPRegion ||
		prev.GeoIPCity != curr.GeoIPCity ||
		prev.PoPCode != curr.PoPCode ||
		prev.PoPIATA != curr.PoPIATA
}

func copyPrevious(record *resolver.ResolvedPrefix, prev resolver.ResolvedPrefix) {
	record.PreviousBGPState = prev.BGPState
	record.PreviousOriginASN = prev.OriginASN
	record.PreviousOperator = prev.Operator
	record.PreviousPoPCode = prev.PoPCode
	record.PreviousGeoIPCountry = prev.GeoIPCountry
	record.PreviousGeoIPRegion = prev.GeoIPRegion
	record.PreviousGeoIPCity = prev.GeoIPCity
}

func changeFrom(curr, prev resolver.ResolvedPrefix, kind, ts string) Change {
	return Change{
		Prefix:               curr.Prefix,
		ChangeType:           kind,
		ChangedAt:            ts,
		Operator:             curr.Operator,
		PreviousOperator:     prev.Operator,
		BGPState:             curr.BGPState,
		PreviousBGPState:     prev.BGPState,
		OriginASN:            curr.OriginASN,
		PreviousOriginASN:    prev.OriginASN,
		PoPCode:              curr.PoPCode,
		PreviousPoPCode:      prev.PoPCode,
		GeoIPCountry:         curr.GeoIPCountry,
		PreviousGeoIPCountry: prev.GeoIPCountry,
		GeoIPRegion:          curr.GeoIPRegion,
		PreviousGeoIPRegion:  prev.GeoIPRegion,
		GeoIPCity:            curr.GeoIPCity,
		PreviousGeoIPCity:    prev.GeoIPCity,
	}
}

func nonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func ptrVal(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func intPtr(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(*v)
}
