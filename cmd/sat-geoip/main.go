package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ipanalytics/Sat-geoip/internal/export"
	"github.com/ipanalytics/Sat-geoip/internal/release"
	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func main() {
	evidencePath := flag.String("evidence", "", "JSON array of prefix evidence records")
	format := flag.String("format", "jsonl", "output format: jsonl|csv|satellite-asns|release|live-release|update-readme-stats")
	outDir := flag.String("out", "outputs", "output directory for release format")
	readmePath := flag.String("readme", "README.md", "README path for update-readme-stats")
	statsPath := flag.String("stats", "", "stats.json path for update-readme-stats")
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
			for _, ev := range inputs {
				records = append(records, resolver.Resolve(ev))
			}
			stats = release.ComputeStats(records)
		}
		if err := release.UpdateREADMEStats(*readmePath, stats); err != nil {
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
	for _, ev := range inputs {
		records = append(records, resolver.Resolve(ev))
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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
