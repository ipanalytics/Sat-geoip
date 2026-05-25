package validators

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/netip"
	"strings"
)

type GeoFeedRow struct {
	Prefix  netip.Prefix
	Country string
	Region  string
	City    string
	Postal  string
}

func ParseRFC8805(r io.Reader) ([]GeoFeedRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	rows := []GeoFeedRow{}
	for line := 1; ; line++ {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if len(record) == 0 || strings.HasPrefix(strings.TrimSpace(record[0]), "#") {
			continue
		}
		if len(record) < 2 {
			return nil, fmt.Errorf("line %d: expected at least prefix and country", line)
		}
		prefix, err := netip.ParsePrefix(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid prefix %q: %w", line, record[0], err)
		}
		country := strings.TrimSpace(record[1])
		if len(country) != 2 {
			return nil, fmt.Errorf("line %d: invalid ISO country %q", line, country)
		}
		row := GeoFeedRow{Prefix: prefix, Country: strings.ToUpper(country)}
		if len(record) > 2 {
			row.Region = strings.TrimSpace(record[2])
		}
		if len(record) > 3 {
			row.City = strings.TrimSpace(record[3])
		}
		if len(record) > 4 {
			row.Postal = strings.TrimSpace(record[4])
		}
		rows = append(rows, row)
	}
	return rows, nil
}
