package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ipanalytics/Sat-geoip/internal/export"
	"github.com/ipanalytics/Sat-geoip/internal/live"
	"github.com/ipanalytics/Sat-geoip/internal/mmdb"
	"github.com/ipanalytics/Sat-geoip/internal/reference"
	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func FromEvidenceFile(evidencePath, outDir string) error {
	input, err := os.Open(evidencePath)
	if err != nil {
		return err
	}
	defer input.Close()

	var evidence []resolver.PrefixEvidence
	if err := json.NewDecoder(input).Decode(&evidence); err != nil {
		return err
	}
	records := make([]resolver.ResolvedPrefix, 0, len(evidence))
	ref, err := reference.LoadIfAvailable("data/reference")
	if err != nil {
		return err
	}
	for _, ev := range evidence {
		records = append(records, resolver.ResolveWithReference(ev, ref))
	}
	return Write(outDir, records)
}

func FromLive(ctx context.Context, outDir string) error {
	evidence, err := live.Evidence(ctx, live.Options{})
	if err != nil {
		return err
	}
	records := make([]resolver.ResolvedPrefix, 0, len(evidence))
	ref, err := reference.LoadIfAvailable("data/reference")
	if err != nil {
		return err
	}
	for _, ev := range evidence {
		records = append(records, resolver.ResolveWithReference(ev, ref))
	}
	return Write(outDir, records)
}

func Write(outDir string, records []resolver.ResolvedPrefix) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	stats := ComputeStats(records)

	writers := []struct {
		name string
		fn   func(io.Writer) error
	}{
		{"sat-geoip-prefixes.jsonl", func(w io.Writer) error { return export.WriteJSONL(w, records) }},
		{"sat-geoip-prefixes.csv", func(w io.Writer) error { return export.WriteResolvedCSV(w, records) }},
		{"satellite-asns.csv", export.WriteSatelliteASNs},
		{"operator-geofeeds.csv", export.WriteOperatorGeoFeeds},
		{"operator-gateway-reference.csv", export.WriteGatewayReference},
		{"starlink-geoip-vs-bgp.csv", func(w io.Writer) error { return export.WriteStarlinkGeoIPVsBGP(w, records) }},
		{"starlink-pop-mapping.csv", func(w io.Writer) error { return export.WriteStarlinkPoPMapping(w, records) }},
		{"pops-vs-ptr-mismatch.csv", func(w io.Writer) error { return export.WritePoPVsPTRMismatch(w, records) }},
		{"stats.json", func(w io.Writer) error { return WriteStatsJSON(w, stats) }},
		{"RELEASE_NOTES.md", func(w io.Writer) error { return WriteStatsMarkdown(w, stats) }},
	}
	for _, writer := range writers {
		if err := writeFile(filepath.Join(outDir, writer.name), writer.fn); err != nil {
			return err
		}
	}
	if err := mmdb.Write(filepath.Join(outDir, "sat-geoip.mmdb"), records); err != nil {
		return fmt.Errorf("write mmdb: %w", err)
	}
	return nil
}

func writeFile(path string, fn func(io.Writer) error) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
