package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func WriteSatelliteASNs(w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"operator", "asn", "asn_name", "orbit_class", "source", "confidence", "notes"}); err != nil {
		return err
	}
	for _, op := range resolver.Operators() {
		cfg := resolver.Registry[op]
		asns := sortedASNs(cfg.ASNs)
		for _, asn := range asns {
			name := cfg.ASNs[asn]
			if err := cw.Write([]string{string(op), fmt.Sprint(asn), name, string(cfg.OrbitClass), "verified_constant_plus_discovery_seed", "0.997", "ASN registry is discovered and expected to grow"}); err != nil {
				return err
			}
		}
	}
	return cw.Error()
}

func WriteResolvedCSV(w io.Writer, records []resolver.ResolvedPrefix) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	header := []string{
		"prefix", "operator", "operator_group", "service_type", "orbit_class", "origin_asn",
		"geoip_country", "geoip_region", "geoip_city", "geoip_source", "pop_code", "pop_iata",
		"bgp_state", "ground_station_claim", "active_user_claim", "quality_flags",
		"attribution_confidence", "geo_confidence",
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, r := range records {
		asn := ""
		if r.OriginASN != nil {
			asn = fmt.Sprint(*r.OriginASN)
		}
		if err := cw.Write([]string{
			r.Prefix, r.Operator, r.OperatorGroup, r.ServiceType, r.OrbitClass, asn,
			r.GeoIPCountry, r.GeoIPRegion, r.GeoIPCity, r.GeoIPSource, r.PoPCode, r.PoPIATA,
			r.BGPState, fmt.Sprint(r.GroundStationClaim), fmt.Sprint(r.ActiveUserClaim),
			fmt.Sprint(r.QualityFlags), fmt.Sprintf("%.3f", r.DataConfidence.Attribution), fmt.Sprintf("%.3f", r.DataConfidence.Geo),
		}); err != nil {
			return err
		}
	}
	return cw.Error()
}

func WriteOperatorGeoFeeds(w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"operator", "url", "type", "status", "format", "notes"}); err != nil {
		return err
	}
	for _, op := range resolver.Operators() {
		cfg := resolver.Registry[op]
		if cfg.GeoIPFeed != "" {
			if err := cw.Write([]string{string(op), cfg.GeoIPFeed, "geoip_feed", "active", "rfc8805", "operator-declared customer subnet GeoIP location"}); err != nil {
				return err
			}
		} else {
			if err := cw.Write([]string{string(op), "", "geoip_feed", "not_found", "", "BGP-derived operator; no public RFC 8805 geofeed known"}); err != nil {
				return err
			}
		}
		if cfg.PoPFeed != "" {
			if err := cw.Write([]string{string(op), cfg.PoPFeed, "pop_feed", "active", "custom_csv", "operator-declared subnet to PoP assignment"}); err != nil {
				return err
			}
		}
	}
	return cw.Error()
}

func WriteGatewayReference(w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"operator", "country", "source", "semantics", "notes"}); err != nil {
		return err
	}
	for _, op := range resolver.Operators() {
		cfg := resolver.Registry[op]
		for _, country := range cfg.GatewayCountries {
			if err := cw.Write([]string{
				string(op),
				country,
				"operator_gateway_reference",
				"gateway_country_reference_not_customer_geoip",
				"Reference location for satellite gateway architecture; do not use as IP GeoIP truth",
			}); err != nil {
				return err
			}
		}
	}
	return cw.Error()
}

func sortedASNs(asns map[int]string) []int {
	out := make([]int, 0, len(asns))
	for asn := range asns {
		out = append(out, asn)
	}
	sort.Ints(out)
	return out
}

func WriteStarlinkGeoIPVsBGP(w io.Writer, records []resolver.ResolvedPrefix) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"prefix", "geoip_country", "geoip_region", "geoip_city", "in_pops_csv", "pop_code", "bgp_announced", "origin_asn", "state"}); err != nil {
		return err
	}
	for _, r := range records {
		if r.Operator != string(resolver.OperatorStarlink) {
			continue
		}
		asn := ""
		if r.OriginASN != nil {
			asn = fmt.Sprint(*r.OriginASN)
		}
		if err := cw.Write([]string{
			r.Prefix,
			r.GeoIPCountry,
			r.GeoIPRegion,
			r.GeoIPCity,
			fmt.Sprint(r.PoPSource == "starlink_pops_csv"),
			r.PoPCode,
			fmt.Sprint(r.BGPState == string(resolver.BGPAnnounced)),
			asn,
			r.BGPState,
		}); err != nil {
			return err
		}
	}
	return cw.Error()
}

func WriteStarlinkPoPMapping(w io.Writer, records []resolver.ResolvedPrefix) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"prefix", "pop_code", "pop_iata", "source"}); err != nil {
		return err
	}
	for _, r := range records {
		if r.Operator == string(resolver.OperatorStarlink) && r.PoPCode != "" {
			if err := cw.Write([]string{r.Prefix, r.PoPCode, r.PoPIATA, r.PoPSource}); err != nil {
				return err
			}
		}
	}
	return cw.Error()
}

func WritePoPVsPTRMismatch(w io.Writer, records []resolver.ResolvedPrefix) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"prefix", "pop_code", "ptr_pop_code", "ptr_state"}); err != nil {
		return err
	}
	for _, r := range records {
		if strings.Contains(r.PTRState, "conflicts") || strings.Contains(r.PTRState, "missing") {
			if err := cw.Write([]string{r.Prefix, r.PoPCode, r.PTRPoPCode, r.PTRState}); err != nil {
				return err
			}
		}
	}
	return cw.Error()
}
