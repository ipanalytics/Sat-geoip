package live

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ipanalytics/Sat-geoip/internal/collectors"
	"github.com/ipanalytics/Sat-geoip/internal/resolver"
	"github.com/ipanalytics/Sat-geoip/internal/validators"
)

type Options struct {
	Client  *http.Client
	Timeout time.Duration
}

func Evidence(ctx context.Context, opts Options) ([]resolver.PrefixEvidence, error) {
	client := opts.Client
	if client == nil {
		client = http.DefaultClient
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 180 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	today := time.Now().UTC().Format("2006-01-02")
	evidence := map[string]resolver.PrefixEvidence{}

	if err := addStarlink(ctx, client, evidence, today); err != nil {
		return nil, err
	}
	if err := addViasat(ctx, client, evidence, today); err != nil {
		return nil, err
	}
	if err := addBGP(ctx, client, evidence, today); err != nil {
		return nil, err
	}

	prefixes := make([]string, 0, len(evidence))
	for prefix := range evidence {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)

	out := make([]resolver.PrefixEvidence, 0, len(prefixes))
	for _, prefix := range prefixes {
		out = append(out, evidence[prefix])
	}
	return out, nil
}

func addStarlink(ctx context.Context, client *http.Client, evidence map[string]resolver.PrefixEvidence, today string) error {
	cfg := resolver.Registry[resolver.OperatorStarlink]
	feed, err := fetch(ctx, client, cfg.GeoIPFeed)
	if err != nil {
		return fmt.Errorf("starlink geofeed: %w", err)
	}
	defer feed.Close()
	rows, err := validators.ParseRFC8805(feed)
	if err != nil {
		return fmt.Errorf("parse starlink geofeed: %w", err)
	}
	for _, row := range rows {
		prefix := row.Prefix.String()
		ev := evidence[prefix]
		ev.Prefix = prefix
		ev.GeoIPFeed = &resolver.GeoIPFeedEvidence{
			Country: row.Country,
			Region:  row.Region,
			City:    row.City,
			Source:  "starlink_feed_csv",
		}
		ev.RIR = &resolver.RIREvidence{Org: "Space Exploration Technologies"}
		ev.FirstSeen = today
		ev.LastSeen = today
		evidence[prefix] = ev
	}

	pops, err := fetch(ctx, client, cfg.PoPFeed)
	if err != nil {
		return fmt.Errorf("starlink pops: %w", err)
	}
	defer pops.Close()
	popRows, err := validators.ParseStarlinkPoPs(pops)
	if err != nil {
		return fmt.Errorf("parse starlink pops: %w", err)
	}
	for _, row := range popRows {
		prefix := row.Prefix.String()
		ev := evidence[prefix]
		ev.Prefix = prefix
		ev.PoPFeed = &resolver.PopFeedEvidence{
			PoPCode: row.PoPCode,
			PoPIATA: row.PoPIATA,
			Source:  "starlink_pops_csv",
		}
		if ev.FirstSeen == "" {
			ev.FirstSeen = today
		}
		ev.LastSeen = today
		evidence[prefix] = ev
	}
	return nil
}

func addViasat(ctx context.Context, client *http.Client, evidence map[string]resolver.PrefixEvidence, today string) error {
	cfg := resolver.Registry[resolver.OperatorViasat]
	feed, err := fetch(ctx, client, cfg.GeoIPFeed)
	if err != nil {
		return fmt.Errorf("viasat geofeed: %w", err)
	}
	defer feed.Close()
	rows, err := validators.ParseRFC8805(feed)
	if err != nil {
		return fmt.Errorf("parse viasat geofeed: %w", err)
	}
	for _, row := range rows {
		prefix := row.Prefix.String()
		ev := evidence[prefix]
		ev.Prefix = prefix
		ev.GeoIPFeed = &resolver.GeoIPFeedEvidence{
			Country: row.Country,
			Region:  row.Region,
			City:    row.City,
			Source:  "viasat_geofeed_csv",
		}
		ev.RIR = &resolver.RIREvidence{Org: "ViaSat, Inc."}
		ev.FirstSeen = today
		ev.LastSeen = today
		evidence[prefix] = ev
	}
	return nil
}

func addBGP(ctx context.Context, client *http.Client, evidence map[string]resolver.PrefixEvidence, today string) error {
	for _, op := range resolver.Operators() {
		cfg := resolver.Registry[op]
		asns := make([]int, 0, len(cfg.ASNs))
		for asn := range cfg.ASNs {
			asns = append(asns, asn)
		}
		sort.Ints(asns)
		for _, asn := range asns {
			url := fmt.Sprintf("https://stat.ripe.net/data/announced-prefixes/data.json?resource=AS%d", asn)
			body, err := fetch(ctx, client, url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: RIPEstat AS%d skipped: %v\n", asn, err)
				continue
			}
			prefixes, err := collectors.ParseRIPEAnnouncedPrefixes(body)
			body.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: RIPEstat AS%d parse skipped: %v\n", asn, err)
				continue
			}
			for _, prefix := range prefixes {
				ev := evidence[prefix]
				ev.Prefix = prefix
				ev.BGP = mergeBGP(ev.BGP, asn)
				if ev.RIR == nil {
					ev.RIR = &resolver.RIREvidence{Org: cfg.ASNs[asn]}
				}
				if ev.FirstSeen == "" {
					ev.FirstSeen = today
				}
				ev.LastSeen = today
				evidence[prefix] = ev
			}
		}
	}
	return nil
}

func mergeBGP(existing *resolver.BGPEvidence, asn int) *resolver.BGPEvidence {
	if existing == nil {
		return &resolver.BGPEvidence{Announced: true, OriginASNs: []int{asn}, EverAnnounced: true}
	}
	existing.Announced = true
	existing.EverAnnounced = true
	for _, seen := range existing.OriginASNs {
		if seen == asn {
			return existing
		}
	}
	existing.OriginASNs = append(existing.OriginASNs, asn)
	sort.Ints(existing.OriginASNs)
	return existing
}

func fetch(ctx context.Context, client *http.Client, url string) (io.ReadCloser, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		body, err := fetchOnce(ctx, client, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retryable(err) || attempt == 2 {
			break
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func fetchOnce(ctx context.Context, client *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "sat-geoip/0.1 (+https://github.com/ipanalytics/Sat-geoip)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: HTTP %d", strings.TrimSpace(url), resp.StatusCode)
	}
	return resp.Body, nil
}

func retryable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 429") ||
		strings.Contains(msg, "HTTP 500") ||
		strings.Contains(msg, "HTTP 502") ||
		strings.Contains(msg, "HTTP 503") ||
		strings.Contains(msg, "HTTP 504") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout")
}
