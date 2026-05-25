package resolver

import "strings"

type OperatorConfig struct {
	OperatorGroup string
	ServiceType   string
	OrbitClass    OrbitClass
	ASNs          map[int]string
	OrgTokens     []string
	GeoIPFeed     string
	PoPFeed       string
}

var Registry = map[Operator]OperatorConfig{
	OperatorStarlink: {
		OperatorGroup: "spacex",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			14593: "SPACEX-STARLINK",
			45700: "IDNIC-STARLINK-AS-ID",
		},
		OrgTokens: []string{"starlink", "spacex", "space exploration"},
		GeoIPFeed: "https://geoip.starlinkisp.net/feed.csv",
		PoPFeed:   "https://geoip.starlinkisp.net/pops.csv",
	},
	OperatorViasat: {
		OperatorGroup: "viasat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			7155:  "VIASAT-SP-BACKBONE",
			40306: "Viasat Inc.",
			31515: "Inmarsat Global Limited",
		},
		OrgTokens: []string{"viasat", "inmarsat"},
		GeoIPFeed: "https://raw.githubusercontent.com/Viasat/geofeed/refs/heads/main/geofeed.csv",
	},
}

func OperatorForASN(asn int) Operator {
	for op, cfg := range Registry {
		if _, ok := cfg.ASNs[asn]; ok {
			return op
		}
	}
	return OperatorUnknown
}

func ASNName(asn int) string {
	for _, cfg := range Registry {
		if name, ok := cfg.ASNs[asn]; ok {
			return name
		}
	}
	return ""
}

func OrgMatchesOperator(org string, op Operator) bool {
	cfg, ok := Registry[op]
	if !ok || org == "" {
		return false
	}
	lower := strings.ToLower(org)
	for _, token := range cfg.OrgTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
