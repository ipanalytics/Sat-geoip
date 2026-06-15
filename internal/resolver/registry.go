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
		},
		OrgTokens: []string{"viasat"},
		GeoIPFeed: "https://raw.githubusercontent.com/Viasat/geofeed/refs/heads/main/geofeed.csv",
	},
	OperatorInmarsat: {
		OperatorGroup: "viasat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			31515: "Inmarsat Global Limited",
		},
		OrgTokens: []string{"inmarsat", "inmarsat global"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"Inmarsat is modeled separately from Viasat AS7155 but remains under the Viasat operator group after acquisition.",
			"No public RFC 8805 geofeed known for AS31515; Viasat group geofeed may cover adjacent networks.",
		},
	},
	OperatorThuraya: {
		OperatorGroup: "space42",
		ServiceType:   "mss_narrowband",
		OrbitClass:    OrbitGEOMSS,
		ASNs: map[int]string{
			44703: "Thuraya Telecommunications",
		},
		OrgTokens: []string{"thuraya", "yahsat", "space42"},
		DataLayers: []string{
			"bgp_origin_prefixes",
			"rdap",
			"rpki",
		},
		Notes: []string{
			"GEO mobile satellite services operator; treat as MSS/narrowband rather than broadband LEO.",
			"No public geofeed known; gateway regions are not customer GeoIP.",
		},
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
	OperatorIridium: {
		OperatorGroup: "iridium",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			22184:  "IRIDIUM SATELLITE, LLC",
			206171: "Iridium Communications",
		},
		OrgTokens:  []string{"iridium", "iridium satellite", "iridium communications"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"LEO MSS / Certus satellite connectivity; BGP-derived prefixes only."},
	},
	OperatorKuiper: {
		OperatorGroup: "amazon",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			398031: "Amazon Kuiper",
		},
		OrgTokens:  []string{"kuiper", "amazon kuiper", "amazon leo"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Early LEO satellite internet network seed; do not include broad shared Amazon backbone ASNs as Kuiper-specific evidence."},
	},
	OperatorTelesat: {
		OperatorGroup: "telesat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			11919:  "Telesat Network Services Inc.",
			19036:  "Telesat Network Services Inc.",
			16212:  "Telesat International Ltd",
			57888:  "Telesat",
			268656: "Telesat Telecomunicacoes Ltda",
		},
		OrgTokens:  []string{"telesat"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO/hybrid satellite operator; future LEO Lightspeed resources should be discovered separately."},
	},
	OperatorYahsat: {
		OperatorGroup: "space42",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			206283: "YAHSAT-FRANKFURT",
			198381: "Star Satellite Communications Company",
			198247: "Yahsat-AbuDhabi",
			208428: "YAHSAT-DUBAI",
		},
		OrgTokens:  []string{"yahsat", "yahclick", "star satellite communications", "space42"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO satellite internet operator under the Space42/Yahsat group; keep separate from Thuraya MSS."},
	},
	OperatorHispasat: {
		OperatorGroup: "hispasat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			197991: "Hispasat S.A.",
			28273:  "Hispamar Satelites S/A",
			265554: "Hispasat Mexico",
		},
		OrgTokens:  []string{"hispasat", "hispamar"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO satellite internet operator for Europe and Latin America; BGP-derived prefixes only."},
	},
	OperatorKacific: {
		OperatorGroup: "kacific",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			135409: "Kacific Broadband Satellites Pte Ltd",
		},
		OrgTokens:  []string{"kacific"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Asia-Pacific GEO satellite broadband operator; BGP-derived prefixes only."},
	},
	OperatorKVH: {
		OperatorGroup: "kvh",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			10021: "KVH Co.,Ltd",
			25687: "KVH INDUSTRIES, INC",
		},
		OrgTokens:  []string{"kvh", "kvh industries"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Maritime VSAT and mobility service provider; mixed satellite and terrestrial infrastructure expected."},
	},
	OperatorAnuvu: {
		OperatorGroup: "anuvu",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			400110: "Anuvu Operations LLC",
			32806:  "MTNSAT Holdings LLC",
		},
		OrgTokens:  []string{"anuvu", "mtnsat", "global eagle", "row 44", "emerging markets comms"},
		GeoIPFeed:  "https://geoip.mtnsat.net/",
		DataLayers: append(bgpDerivedLayers(), "geoip_feed"),
		Notes:      []string{"Aviation and maritime satellite connectivity provider; MTNSAT geofeed is tracked as operator-declared GeoIP when parseable."},
	},
	OperatorPanasonic: {
		OperatorGroup: "panasonic",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			64294:  "Panasonic Avionics Corporation",
			394603: "Panasonic Avionics Corporation",
		},
		OrgTokens:  []string{"panasonic avionics", "panasonic"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"In-flight connectivity satellite service provider; BGP-derived prefixes only."},
	},
	OperatorSatcomDirect: {
		OperatorGroup: "satcom_direct",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			30252:  "Satcom Direct",
			394145: "Satcom Direct",
			54095:  "Satcom Direct",
		},
		OrgTokens:  []string{"satcom direct"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Business aviation, maritime, and enterprise satellite connectivity provider; BGP-derived prefixes only."},
	},
	OperatorGogo: {
		OperatorGroup: "gogo",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			33023: "Gogo Business Aviation",
		},
		OrgTokens:  []string{"gogo", "aircell"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Aviation connectivity provider; may include satellite and air-to-ground infrastructure."},
	},
	OperatorRigNet: {
		OperatorGroup: "viasat",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			18634:  "RigNet",
			20468:  "RigNet",
			394348: "RigNet",
		},
		OrgTokens:  []string{"rignet", "rig net"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Remote energy and maritime connectivity provider now associated with Viasat; BGP-derived prefixes only."},
	},
	OperatorNSSLGlobal: {
		OperatorGroup: "nsslglobal",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			42106: "NSSLGlobal",
		},
		OrgTokens:  []string{"nsslglobal", "nssl global"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Maritime, government, and enterprise satellite connectivity provider; BGP-derived prefixes only."},
	},
	OperatorSkyPerfectJSAT: {
		OperatorGroup: "sky_perfect_jsat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			17944: "SKY Perfect JSAT",
		},
		OrgTokens:  []string{"sky perfect", "jsat", "skyperfect"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Japanese GEO satellite operator and connectivity provider; BGP-derived prefixes only."},
	},
	OperatorTelespazio: {
		OperatorGroup: "telespazio",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			8947:   "Telespazio S.p.A.",
			196963: "Telespazio Germany",
			60471:  "Telespazio France",
			16013:  "Telespazio Argentina",
		},
		OrgTokens:  []string{"telespazio"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"European satellite teleport and connectivity operator; BGP-derived prefixes only."},
	},
	OperatorThaicom: {
		OperatorGroup: "thaicom",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			4653:  "Thaicom",
			17462: "Thaicom",
		},
		OrgTokens:  []string{"thaicom", "ipstar"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Thai GEO satellite internet operator; BGP-derived prefixes only."},
	},
	OperatorRSCC: {
		OperatorGroup: "rscc",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			8751: "Russian Satellite Communications Company",
		},
		OrgTokens:  []string{"rscc", "russian satellite communications", "space communication"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO satellite communications operator; BGP-derived prefixes only."},
	},
	OperatorGazpromSpace: {
		OperatorGroup: "gazprom_space_systems",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			35408: "Gazprom Space Systems",
			42846: "Gazprom Space Systems",
		},
		OrgTokens:  []string{"gazprom space", "yamal"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO satellite communications operator; BGP-derived prefixes only."},
	},
	OperatorNBNSkyMuster: {
		OperatorGroup: "nbn",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			4764:   "Aussie Broadband / NBN adjacency",
			133378: "NBN Co",
		},
		OrgTokens:  []string{"nbn co", "sky muster", "nbn"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Australian Sky Muster satellite internet seed; AS4764 may include non-satellite infrastructure and should be refined by evidence over time."},
	},
	OperatorTurksat: {
		OperatorGroup: "turksat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			47524: "Turksat",
		},
		OrgTokens:  []string{"turksat"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Turkish GEO satellite communications operator; BGP-derived prefixes only."},
	},
	OperatorOmniAccess: {
		OperatorGroup: "omniaccess",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			44431: "OmniAccess S.L.",
		},
		OrgTokens:  []string{"omniaccess", "omni access"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Maritime and managed satellite connectivity provider; BGP-derived prefixes only."},
	},
	OperatorCastorMarine: {
		OperatorGroup: "castor_marine",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			60647: "Castor Marine B.V.",
		},
		OrgTokens:  []string{"castor marine"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Offshore, maritime, and remote connectivity provider; BGP-derived prefixes only."},
	},
	OperatorNavarino: {
		OperatorGroup: "navarino",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			203101: "Navarino Single Member S.A.",
		},
		OrgTokens:  []string{"navarino"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Maritime satellite connectivity integrator; BGP-derived prefixes only."},
	},
	OperatorTampnet: {
		OperatorGroup: "tampnet",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitOffshore,
		ASNs: map[int]string{
			35310:  "Tampnet",
			394334: "Tampnet",
		},
		OrgTokens:  []string{"tampnet"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Hybrid offshore connectivity provider; may include satellite, microwave, and subsea infrastructure."},
	},
	OperatorGilatTelecom: {
		OperatorGroup: "gilat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			12491: "Gilat Telecom",
			56804: "Gilat Telecom",
		},
		OrgTokens:  []string{"gilat", "gilat telecom"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO VSAT and rural connectivity provider; BGP-derived prefixes only."},
	},
	OperatorSwarm: {
		OperatorGroup: "spacex",
		ServiceType:   "satellite_iot",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			400263: "Swarm Technologies",
		},
		OrgTokens:  []string{"swarm technologies", "swarm"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"LEO satellite IoT network under SpaceX; keep separate from Starlink broadband semantics."},
	},
	OperatorSpaceXInfra: {
		OperatorGroup: "spacex",
		ServiceType:   "satellite_infrastructure",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			36936: "Space Exploration Technologies",
			46485: "SpaceX Services",
		},
		OrgTokens:  []string{"spacex services"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"SpaceX infrastructure ASNs; do not merge with Starlink customer-subnet GeoIP semantics."},
	},
	OperatorCarnival: {
		OperatorGroup: "carnival",
		ServiceType:   "satellite_cruise_line",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			36151: "Carnival Corporation",
		},
		OrgTokens:  []string{"carnival corporation", "carnival"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Cruise-line operated network with satellite connectivity dependencies; BGP-derived prefixes only."},
	},
	OperatorRoyalCaribbean: {
		OperatorGroup: "royal_caribbean",
		ServiceType:   "satellite_cruise_line",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			40492: "Royal Caribbean Cruises Ltd.",
		},
		OrgTokens:  []string{"royal caribbean"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Cruise-line operated network with satellite connectivity dependencies; BGP-derived prefixes only."},
	},
	OperatorUSAP: {
		OperatorGroup: "usap",
		ServiceType:   "satellite_research",
		OrbitClass:    OrbitGEOMSS,
		ASNs: map[int]string{
			35266: "United States Antarctic Program",
		},
		OrgTokens:  []string{"united states antarctic program", "usap"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Polar research network with satellite connectivity dependencies; BGP-derived prefixes only."},
	},
	OperatorKSAT: {
		OperatorGroup: "ksat",
		ServiceType:   "satellite_research",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			42055: "Kongsberg Satellite Services AS",
		},
		OrgTokens:  []string{"kongsberg satellite", "ksat"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Satellite ground-station and data services operator; BGP-derived prefixes only."},
	},
	OperatorThalesAvionics: {
		OperatorGroup: "thales",
		ServiceType:   "satellite_aero",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			398918: "Thales Avionics",
		},
		OrgTokens:  []string{"thales avionics", "thales inflyt", "thales"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Aviation connectivity network; BGP-derived prefixes only."},
	},
	OperatorLufthansaSystems: {
		OperatorGroup: "lufthansa",
		ServiceType:   "satellite_aero",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			31022: "Lufthansa Systems",
		},
		OrgTokens:  []string{"lufthansa systems", "lufthansa"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Aviation connectivity and airline network seed; BGP-derived prefixes only."},
	},
	OperatorCapRock: {
		OperatorGroup: "speedcast",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			14264: "CapRock Communications",
		},
		OrgTokens:  []string{"caprock", "caprock communications"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Legacy offshore and remote-site satellite connectivity provider now associated with Speedcast."},
	},
	OperatorBentleyWalker: {
		OperatorGroup: "bentley_walker",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			39122: "Bentley Walker Ltd",
		},
		OrgTokens:  []string{"bentley walker", "freedomsat"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"GEO satellite internet and VSAT provider; BGP-derived prefixes only."},
	},
	OperatorITCGlobal: {
		OperatorGroup: "panasonic",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			53907: "ITC Global",
		},
		OrgTokens:  []string{"itc global"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Legacy energy and mining satellite connectivity provider now associated with Panasonic."},
	},
	OperatorChinaSatcom: {
		OperatorGroup: "china_satcom",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			134762: "China Satellite Communications",
			140902: "Sino Satellite",
		},
		OrgTokens:  []string{"china satcom", "china satellite", "sino satellite"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Chinese GEO satellite communications operator; BGP-derived prefixes only."},
	},
	OperatorAPSTAR: {
		OperatorGroup: "apstar",
		ServiceType:   "satellite_service_provider",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			4636:   "APT Satellite / APSTAR",
			132214: "APT Satellite / APSTAR",
		},
		OrgTokens:  []string{"apt satellite", "apstar"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"APSTAR/APT satellite connectivity provider; BGP-derived prefixes only."},
	},
	OperatorMorsviazsputnik: {
		OperatorGroup: "morsviazsputnik",
		ServiceType:   "satellite_maritime_strategic",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			25542: "Morsviazsputnik",
		},
		OrgTokens:  []string{"morsviazsputnik", "morsvyazsputnik"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Maritime satellite communications operator; BGP-derived prefixes only."},
	},
	OperatorRuSatAmTel: {
		OperatorGroup: "rusat_amtel",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			28811: "RuSat",
			34812: "AmTel-Svyaz",
		},
		OrgTokens:  []string{"rusat", "amtel", "amtel-svyaz"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Russian VSAT and remote connectivity providers; BGP-derived prefixes only."},
	},
	OperatorIntelsatGeneral: {
		OperatorGroup: "intelsat",
		ServiceType:   "satellite_military",
		OrbitClass:    OrbitHybrid,
		ASNs: map[int]string{
			22557: "Intelsat General Corporation",
		},
		OrgTokens:  []string{"intelsat general"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Government satellite communications network; modeled as infrastructure attribution only."},
	},
	OperatorNASAJPL: {
		OperatorGroup: "nasa",
		ServiceType:   "space_agency_telemetry",
		OrbitClass:    OrbitDeepSpace,
		ASNs: map[int]string{
			10461: "NASA Jet Propulsion Laboratory",
		},
		OrgTokens:  []string{"nasa", "jet propulsion laboratory", "jpl"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Space agency network seed; not customer GeoIP and not satellite ISP consumer infrastructure."},
	},
	OperatorRocketLab: {
		OperatorGroup: "rocket_lab",
		ServiceType:   "spaceport",
		OrbitClass:    OrbitLEO,
		ASNs: map[int]string{
			398322: "Rocket Lab",
		},
		OrgTokens:  []string{"rocket lab"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Commercial space infrastructure network seed; BGP-derived prefixes only."},
	},
	OperatorCNES: {
		OperatorGroup: "cnes",
		ServiceType:   "spaceport",
		OrbitClass:    OrbitGroundInfrastructure,
		ASNs: map[int]string{
			27702: "Centre National d'Etudes Spatiales",
		},
		OrgTokens:  []string{"cnes", "centre national d'etudes spatiales"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Space agency and launch-site infrastructure seed; BGP-derived prefixes only."},
	},
	OperatorESA: {
		OperatorGroup: "esa",
		ServiceType:   "space_agency_telemetry",
		OrbitClass:    OrbitMixed,
		ASNs: map[int]string{
			28958: "European Space Agency",
			29076: "ESA ESRIN",
		},
		OrgTokens:  []string{"european space agency", "esa", "esrin"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"Space agency infrastructure seed; BGP-derived prefixes only."},
	},
	OperatorKTSat: {
		OperatorGroup: "kt_sat",
		ServiceType:   "satellite_internet",
		OrbitClass:    OrbitGEO,
		ASNs: map[int]string{
			38053: "KT Sat",
		},
		OrgTokens:  []string{"kt sat", "korea telecom sat"},
		DataLayers: bgpDerivedLayers(),
		Notes:      []string{"South Korean GEO satellite communications operator; BGP-derived prefixes only."},
	},
}

func bgpDerivedLayers() []string {
	return []string{
		"bgp_origin_prefixes",
		"rdap",
		"rpki",
	}
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
