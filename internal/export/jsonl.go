package export

import (
	"encoding/json"
	"io"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func WriteJSONL(w io.Writer, records []resolver.ResolvedPrefix) error {
	enc := json.NewEncoder(w)
	for _, record := range records {
		if err := enc.Encode(record); err != nil {
			return err
		}
	}
	return nil
}
