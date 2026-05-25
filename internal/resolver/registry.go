package resolver

import (
	"sort"
	"strings"
)

type OperatorConfig struct {
	OperatorGroup    string
	ServiceType      string
	OrbitClass       OrbitClass
	ASNs             map[int]string
	OrgTokens        []string
	GeoIPFeed        string
	PoPFeed          string
	IRRSets          []string
	DataLayers       []string
	Notes            []string
	GatewayCountries []string
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
	OperatorSESO3B: {
		OperatorGroup: "ses",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitMEO,
		ASNs: map[int]string{
			60725: "O3B-AS",
		},
		OrgTokens: []string{"o3b", "ses networks", "ses"},
		IRRSets:   []string{"AS-O3B", "AS-O3B-TX-US"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
			"gateway_reference_locations",
		},
		GatewayCountries: []string{"ZA", "PE", "BR", "PT", "AU", "GR", "US", "CL", "AE", "SN"},
		Notes: []string{
			"SES/O3b is modeled as BGP-derived MEO satellite internet; no public RFC 8805 geofeed is known.",
			"Gateway countries are reference locations and must not be treated as customer GeoIP.",
			"Do not include SES ASTRA AS12684; it is broadcast/media infrastructure rather than satellite internet.",
		},
	},
	OperatorMarlink: {
		OperatorGroup: "marlink",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			5377:  "Marlink AS",
			55784: "Marlink AS APNIC region",
		},
		OrgTokens: []string{"marlink", "vizada"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Marlink is modeled as a satellite connectivity service provider, not a constellation owner.",
			"Do not classify as LEO; expect mixed satellite plus terrestrial backbone infrastructure.",
		},
	},
	OperatorHughes: {
		OperatorGroup: "echostar",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			6621:  "Hughes Network Systems",
			63062: "Hughes Network Systems, LLC",
		},
		OrgTokens: []string{"hughes", "echostar", "hughesnet"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Hughes/HughesNet is modeled as GEO satellite internet using BGP-derived evidence.",
			"Regional Hughes ASNs should be discovered and appended over time.",
		},
	},
	OperatorOneWeb: {
		OperatorGroup: "eutelsat_oneweb",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			800: "ONEWEB",
		},
		OrgTokens: []string{"oneweb", "eutelsat oneweb", "network access associates", "worldvu"},
		IRRSets:   []string{"AS-OW"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
			"facility_reference_locations",
		},
		Notes: []string{
			"Core LEO satellite internet operator; keep separate from Eutelsat/Skylogic GEO infrastructure.",
			"PeeringDB facilities are interconnection/gateway references, not customer GeoIP.",
			"No public Starlink-style RFC 8805 geofeed known; use BGP-derived prefixes only.",
		},
	},
	OperatorIntelsat: {
		OperatorGroup: "intelsat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGeoMulti,
		ASNs: map[int]string{
			22351: "INTELSAT GLOBAL SERVICE CORPORATION",
			26243: "Intelsat",
		},
		OrgTokens: []string{"intelsat", "intelsat global", "intelsat global service"},
		IRRSets:   []string{"AS-INTELSATUS", "RS-INTELSAT"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Large GEO / multi-orbit satellite connectivity operator.",
			"Use as BGP-derived satellite operator; do not infer customer location without geofeed.",
			"SES acquisition may consolidate Intelsat under the SES group over time.",
		},
	},
	OperatorAvanti: {
		OperatorGroup: "avanti",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			39356:  "Avanti Broadband Ltd",
			328306: "Avanti Communications South Africa",
		},
		OrgTokens: []string{"avanti", "avanti communications", "avanti broadband", "iwayafrica"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"GEO satellite broadband / VSAT provider for Africa and EMEA coverage.",
			"No public operator geofeed known; BGP-derived prefixes only.",
		},
	},
	OperatorSpeedcast: {
		OperatorGroup: "speedcast",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			4913:   "Speedcast Communications, Inc",
			5666:   "Speedcast Communications, Inc",
			38456:  "SpeedCast Australia",
			132160: "Oceanic Broadband Solutions",
		},
		OrgTokens: []string{"speedcast", "speedcast australia", "speedcast managed services", "oceanic broadband"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Satellite and hybrid connectivity provider, not a pure LEO/GEO constellation owner.",
			"Use as maritime, energy, remote-site and enterprise satellite connectivity layer.",
			"May include terrestrial/hybrid infrastructure; keep service_type as satellite_service_provider.",
		},
	},
	OperatorEutelsatSkylogic: {
		OperatorGroup: "eutelsat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			29286: "SKYLOGIC S.P.A.",
		},
		OrgTokens: []string{"eutelsat", "skylogic", "skylogic s.p.a"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"peeringdb",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Separate from Eutelsat OneWeb AS800.",
			"Use for GEO/hybrid satellite ISP prefixes and legacy Eutelsat/Skylogic infrastructure.",
			"Do not merge with OneWeb LEO semantics.",
		},
	},
}

func Operators() []Operator {
	ops := make([]Operator, 0, len(Registry))
	for op := range Registry {
		ops = append(ops, op)
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i] < ops[j] })
	return ops
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
