package collectors

import (
	"encoding/json"
	"fmt"
	"io"
)

type RIPEAnnouncedPrefixesResponse struct {
	Data struct {
		Prefixes []struct {
			Prefix string `json:"prefix"`
		} `json:"prefixes"`
	} `json:"data"`
}

func ParseRIPEAnnouncedPrefixes(r io.Reader) ([]string, error) {
	var payload RIPEAnnouncedPrefixesResponse
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, err
	}
	prefixes := make([]string, 0, len(payload.Data.Prefixes))
	for _, item := range payload.Data.Prefixes {
		if item.Prefix == "" {
			return nil, fmt.Errorf("RIPEstat announced-prefixes response contains empty prefix")
		}
		prefixes = append(prefixes, item.Prefix)
	}
	return prefixes, nil
}
