package validators

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/netip"
	"strings"
)

type PoPRow struct {
	Prefix  netip.Prefix
	PoPCode string
	PoPIATA string
}

func ParseStarlinkPoPs(r io.Reader) ([]PoPRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	rows := []PoPRow{}
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
		if len(record) != 3 {
			return nil, fmt.Errorf("line %d: expected CIDR,pop_code,iata", line)
		}
		prefix, err := netip.ParsePrefix(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid prefix %q: %w", line, record[0], err)
		}
		pop := strings.TrimSpace(record[1])
		iata := strings.ToLower(strings.TrimSpace(record[2]))
		if pop == "" || iata == "" {
			return nil, fmt.Errorf("line %d: invalid pop mapping", line)
		}
		rows = append(rows, PoPRow{Prefix: prefix, PoPCode: pop, PoPIATA: iata})
	}
	return rows, nil
}
