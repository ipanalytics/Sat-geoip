package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ipanalytics/Sat-geoip/internal/export"
	"github.com/ipanalytics/Sat-geoip/internal/reference"
	"github.com/ipanalytics/Sat-geoip/internal/release"
	"github.com/ipanalytics/Sat-geoip/internal/resolver"
	"github.com/ipanalytics/Sat-geoip/internal/site"
)

func main() {
	evidencePath := flag.String("evidence", "", "JSON array of prefix evidence records")
	format := flag.String("format", "jsonl", "output format: jsonl|csv|satellite-asns|release|live-release|site-data|update-readme-stats")
	outDir := flag.String("out", "outputs", "output directory for release format")
	readmePath := flag.String("readme", "README.md", "README path for update-readme-stats")
	statsPath := flag.String("stats", "", "stats.json path for update-readme-stats")
	recordsPath := flag.String("records", "outputs/sat-geoip-prefixes.jsonl", "resolved prefix JSONL path for site-data")
	gatewaysPath := flag.String("gateways", "outputs/operator-gateway-reference.csv", "gateway reference CSV path for site-data")
	referenceRoot := flag.String("reference", "data/reference", "reference dataset root for site-data")
	flag.Parse()

	if *format == "satellite-asns" {
		if err := export.WriteSatelliteASNs(os.Stdout); err != nil {
			fatal(err)
		}
		return
	}
	if *format == "live-release" {
		if err := release.FromLive(context.Background(), *outDir); err != nil {
			fatal(err)
		}
		return
	}
	if *format == "update-readme-stats" {
		var stats release.Stats
		if *statsPath != "" {
			f, err := os.Open(*statsPath)
			if err != nil {
				fatal(err)
			}
			defer f.Close()
			if err := json.NewDecoder(f).Decode(&stats); err != nil {
				fatal(err)
			}
		} else {
			if *evidencePath == "" {
				fatal(fmt.Errorf("-evidence or -stats is required for update-readme-stats"))
			}
			f, err := os.Open(*evidencePath)
			if err != nil {
				fatal(err)
			}
			defer f.Close()
			var inputs []resolver.PrefixEvidence
			if err := json.NewDecoder(f).Decode(&inputs); err != nil {
				fatal(err)
			}
			records := make([]resolver.ResolvedPrefix, 0, len(inputs))
			ref, err := reference.LoadIfAvailable("data/reference")
			if err != nil {
				fatal(err)
			}
			for _, ev := range inputs {
				records = append(records, resolver.ResolveWithReference(ev, ref))
			}
			stats = release.ComputeStats(records)
		}
		if err := release.UpdateREADMEStats(*readmePath, stats); err != nil {
			fatal(err)
		}
		return
	}
	if *format == "site-data" {
		if err := site.GenerateDashboard(site.GenerateOptions{
			RecordsPath:   *recordsPath,
			StatsPath:     nonEmpty(*statsPath, "outputs/stats.json"),
			GatewaysPath:  *gatewaysPath,
			ReferenceRoot: *referenceRoot,
			OutDir:        *outDir,
		}); err != nil {
			fatal(err)
		}
		return
	}
	if *format == "release" {
		if *evidencePath == "" {
			fatal(fmt.Errorf("-evidence is required for release output"))
		}
		if err := release.FromEvidenceFile(*evidencePath, *outDir); err != nil {
			fatal(err)
		}
		return
	}
	if *evidencePath == "" {
		fatal(fmt.Errorf("-evidence is required for %s output", *format))
	}

	f, err := os.Open(*evidencePath)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	var inputs []resolver.PrefixEvidence
	if err := json.NewDecoder(f).Decode(&inputs); err != nil {
		fatal(err)
	}
	records := make([]resolver.ResolvedPrefix, 0, len(inputs))
	ref, err := reference.LoadIfAvailable("data/reference")
	if err != nil {
		fatal(err)
	}
	for _, ev := range inputs {
		records = append(records, resolver.ResolveWithReference(ev, ref))
	}

	switch *format {
	case "jsonl":
		err = export.WriteJSONL(os.Stdout, records)
	case "csv":
		err = export.WriteResolvedCSV(os.Stdout, records)
	default:
		err = fmt.Errorf("unsupported format %q", *format)
	}
	if err != nil {
		fatal(err)
	}
}

func nonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
