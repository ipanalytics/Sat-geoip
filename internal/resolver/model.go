package resolver

type Operator string

const (
	OperatorStarlink         Operator = "starlink"
	OperatorViasat           Operator = "viasat"
	OperatorSESO3B           Operator = "ses_o3b"
	OperatorMarlink          Operator = "marlink"
	OperatorHughes           Operator = "hughes"
	OperatorOneWeb           Operator = "oneweb"
	OperatorIntelsat         Operator = "intelsat"
	OperatorAvanti           Operator = "avanti"
	OperatorSpeedcast        Operator = "speedcast"
	OperatorEutelsatSkylogic Operator = "eutelsat_skylogic"
	OperatorKuiper           Operator = "kuiper"
	OperatorUnknown          Operator = "unknown"
)

type OrbitClass string

const (
	OrbitLEO      OrbitClass = "leo"
	OrbitMEO      OrbitClass = "meo"
	OrbitGEO      OrbitClass = "geo"
	OrbitHybrid   OrbitClass = "geo_or_hybrid_satellite"
	OrbitGeoMulti OrbitClass = "geo_or_multi_orbit"
	OrbitMixed    OrbitClass = "mixed_satellite"
	OrbitUnknown  OrbitClass = "unknown"
)

type BGPState string

const (
	BGPPlannedNotAnnounced BGPState = "planned_not_announced"
	BGPAnnounced           BGPState = "announced"
	BGPWithdrawnRecently   BGPState = "withdrawn_recently"
	BGPOnly                BGPState = "bgp_only"
	BGPGeoIPOnly           BGPState = "geoip_only"
)

const (
	GeoIPSemanticsCustomerSubnet = "customer_subnet_geoip_location"
	GeoIPSemanticsNone           = "no_geolocation"
	PopSemanticsSubnetToPoP      = "subnet_to_pop_assignment"
)

const (
	FlagGeoIPValid                  = "geoip_valid"
	FlagGeoIPInvalidCountryCityPair = "geoip_invalid_country_city_pair"
	FlagCountryMismatchRIR          = "country_mismatch_rir"
	FlagPopPresent                  = "pop_present"
	FlagPopMissing                  = "pop_missing"
	FlagPTRMissing                  = "ptr_missing"
	FlagPTRMatchesPoP               = "ptr_matches_pop"
	FlagPTRConflictsWithPoP         = "ptr_conflicts_with_pop"
	FlagBGPAnnounced                = "bgp_announced"
	FlagBGPNotAnnounced             = "bgp_not_announced"
	FlagWithdrawnRecently           = "withdrawn_recently"
	FlagOriginASNExpected           = "origin_asn_expected"
	FlagOriginASNUnexpected         = "origin_asn_unexpected"
	FlagPrefixOnlyInGeoFeed         = "prefix_only_in_geofeed"
	FlagPrefixOnlyInBGP             = "prefix_only_in_bgp"
	FlagRPKIInvalid                 = "rpki_invalid"
)

type GeoIPFeedEvidence struct {
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	Source  string `json:"source,omitempty"`
}

type PopFeedEvidence struct {
	PoPCode string `json:"pop_code,omitempty"`
	PoPIATA string `json:"pop_iata,omitempty"`
	Source  string `json:"source,omitempty"`
}

type PTREvidence struct {
	Record  string `json:"record,omitempty"`
	PoPCode string `json:"pop_code,omitempty"`
}

type BGPEvidence struct {
	Announced     bool  `json:"announced"`
	OriginASNs    []int `json:"origin_asns,omitempty"`
	EverAnnounced bool  `json:"ever_announced"`
}

type RIREvidence struct {
	Org     string `json:"org,omitempty"`
	Country string `json:"country,omitempty"`
}

type RPKIEvidence struct {
	HasVRP         bool  `json:"has_vrp"`
	ValidForOrigin *bool `json:"valid_for_origin,omitempty"`
}

type PrefixEvidence struct {
	Prefix    string             `json:"prefix"`
	GeoIPFeed *GeoIPFeedEvidence `json:"geoip_feed,omitempty"`
	PoPFeed   *PopFeedEvidence   `json:"pop_feed,omitempty"`
	PTR       *PTREvidence       `json:"ptr,omitempty"`
	BGP       *BGPEvidence       `json:"bgp,omitempty"`
	RIR       *RIREvidence       `json:"rir,omitempty"`
	RPKI      *RPKIEvidence      `json:"rpki,omitempty"`
	FirstSeen string             `json:"first_seen,omitempty"`
	LastSeen  string             `json:"last_seen,omitempty"`
}

type DataConfidence struct {
	Attribution float64 `json:"attribution"`
	Geo         float64 `json:"geo"`
}

type GeoReference interface {
	CityCountry(country, city string) (known bool, valid bool)
	RegionCountry(country, region string) (known bool, valid bool)
}

type SourcePriority struct {
	Operator    string `json:"operator,omitempty"`
	GeoLocation string `json:"geolocation,omitempty"`
	PoP         string `json:"pop,omitempty"`
	Routing     string `json:"routing,omitempty"`
}

type ResolvedPrefix struct {
	Prefix             string         `json:"prefix"`
	Operator           string         `json:"operator"`
	OperatorGroup      string         `json:"operator_group"`
	ServiceType        string         `json:"service_type"`
	OrbitClass         string         `json:"orbit_class"`
	OriginASN          *int           `json:"origin_asn,omitempty"`
	OriginASName       string         `json:"origin_as_name,omitempty"`
	GeoIPCountry       string         `json:"geoip_country,omitempty"`
	GeoIPRegion        string         `json:"geoip_region,omitempty"`
	GeoIPCity          string         `json:"geoip_city,omitempty"`
	GeoIPSource        string         `json:"geoip_source,omitempty"`
	GeoIPSemantics     string         `json:"geoip_semantics"`
	PoPCode            string         `json:"pop_code,omitempty"`
	PoPIATA            string         `json:"pop_iata,omitempty"`
	PoPSource          string         `json:"pop_source,omitempty"`
	PoPSemantics       string         `json:"pop_semantics,omitempty"`
	BGPState           string         `json:"bgp_state"`
	BGPOriginMatch     bool           `json:"bgp_origin_match"`
	PTRRecord          string         `json:"ptr_record,omitempty"`
	PTRPoPCode         string         `json:"ptr_pop_code,omitempty"`
	PTRState           string         `json:"ptr_state,omitempty"`
	GroundStationClaim bool           `json:"ground_station_claim"`
	ActiveUserClaim    bool           `json:"active_user_claim"`
	QualityFlags       []string       `json:"quality_flags"`
	DataConfidence     DataConfidence `json:"data_confidence"`
	SourcePriority     SourcePriority `json:"source_priority"`
	FirstSeen          string         `json:"first_seen,omitempty"`
	LastSeen           string         `json:"last_seen,omitempty"`
}
