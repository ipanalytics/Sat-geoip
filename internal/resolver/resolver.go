package resolver

import (
	"math"
	"strings"
)

var cityCountry = map[string]string{
	"riyadh":      "SA",
	"vancouver":   "CA",
	"manila":      "PH",
	"carlsbad":    "US",
	"farnborough": "GB",
	"torino":      "IT",
	"colombo":     "LK",
}

func Resolve(ev PrefixEvidence) ResolvedPrefix {
	return ResolveWithReference(ev, nil)
}

func ResolveWithReference(ev PrefixEvidence, ref GeoReference) ResolvedPrefix {
	op, originASN := resolveOperator(ev)
	cfg := Registry[op]
	flags := qualityFlags(ev, op, ref)
	state := resolveBGPState(ev)
	popCode, popIATA, popSource := resolvePoP(ev)

	geoSemantics := GeoIPSemanticsNone
	if ev.GeoIPFeed != nil && ev.GeoIPFeed.Country != "" {
		geoSemantics = GeoIPSemanticsCustomerSubnet
	}

	var originPtr *int
	originName := ""
	if originASN != 0 {
		originPtr = &originASN
		originName = ASNName(originASN)
	}

	ptrState := ""
	if hasFlag(flags, FlagPTRMatchesPoP) {
		ptrState = FlagPTRMatchesPoP
	} else if hasFlag(flags, FlagPTRConflictsWithPoP) {
		ptrState = FlagPTRConflictsWithPoP
	} else if hasFlag(flags, FlagPTRMissing) {
		ptrState = FlagPTRMissing
	}

	res := ResolvedPrefix{
		Prefix:             ev.Prefix,
		Operator:           string(op),
		OperatorGroup:      cfg.OperatorGroup,
		ServiceType:        cfg.ServiceType,
		OrbitClass:         string(cfg.OrbitClass),
		OriginASN:          originPtr,
		OriginASName:       originName,
		GeoIPSemantics:     geoSemantics,
		PoPCode:            popCode,
		PoPIATA:            popIATA,
		PoPSource:          popSource,
		BGPState:           string(state),
		BGPOriginMatch:     bgpOriginMatches(ev, op),
		PTRState:           ptrState,
		GroundStationClaim: false,
		ActiveUserClaim:    ev.BGP != nil && ev.BGP.Announced,
		QualityFlags:       flags,
		DataConfidence: DataConfidence{
			Attribution: computeAttributionConfidence(ev, op, flags),
			Geo:         computeGeoConfidence(ev, op, flags),
		},
		SourcePriority: SourcePriority{
			Operator:    sourceForOperator(originASN),
			GeoLocation: sourceIf(ev.GeoIPFeed != nil, "geoip_feed"),
			PoP:         popSource,
			Routing:     "bgp_observed",
		},
		FirstSeen: ev.FirstSeen,
		LastSeen:  ev.LastSeen,
	}

	if ev.GeoIPFeed != nil {
		res.GeoIPCountry = ev.GeoIPFeed.Country
		res.GeoIPRegion = ev.GeoIPFeed.Region
		res.GeoIPCity = ev.GeoIPFeed.City
		res.GeoIPSource = ev.GeoIPFeed.Source
		if res.GeoIPSource == "" {
			res.GeoIPSource = defaultGeoSource(op)
		}
	}
	if popCode != "" {
		res.PoPSemantics = PopSemanticsSubnetToPoP
	}
	if ev.PTR != nil {
		res.PTRRecord = ev.PTR.Record
		res.PTRPoPCode = ev.PTR.PoPCode
	}

	return res
}

func resolveOperator(ev PrefixEvidence) (Operator, int) {
	if ev.BGP != nil {
		for _, asn := range ev.BGP.OriginASNs {
			if op := OperatorForASN(asn); op != OperatorUnknown {
				return op, asn
			}
		}
		if len(ev.BGP.OriginASNs) > 0 {
			return OperatorUnknown, ev.BGP.OriginASNs[0]
		}
	}
	if ev.RIR != nil {
		for op := range Registry {
			if OrgMatchesOperator(ev.RIR.Org, op) {
				return op, 0
			}
		}
	}
	return OperatorUnknown, 0
}

func resolveBGPState(ev PrefixEvidence) BGPState {
	inFeed := ev.GeoIPFeed != nil
	announced := ev.BGP != nil && ev.BGP.Announced
	ever := ev.BGP != nil && ev.BGP.EverAnnounced

	switch {
	case inFeed && announced:
		return BGPAnnounced
	case inFeed && !announced && ever:
		return BGPWithdrawnRecently
	case inFeed && !announced:
		return BGPPlannedNotAnnounced
	case !inFeed && announced:
		return BGPOnly
	default:
		return BGPGeoIPOnly
	}
}

func resolvePoP(ev PrefixEvidence) (string, string, string) {
	if ev.PoPFeed != nil && ev.PoPFeed.PoPCode != "" {
		source := ev.PoPFeed.Source
		if source == "" {
			source = "starlink_pops_csv"
		}
		return ev.PoPFeed.PoPCode, ev.PoPFeed.PoPIATA, source
	}
	if ev.PTR != nil && ev.PTR.PoPCode != "" {
		return ev.PTR.PoPCode, "", "ptr_observed"
	}
	return "", "", ""
}

func qualityFlags(ev PrefixEvidence, op Operator, ref GeoReference) []string {
	flags := make([]string, 0, 8)
	inFeed := ev.GeoIPFeed != nil
	announced := ev.BGP != nil && ev.BGP.Announced
	ever := ev.BGP != nil && ev.BGP.EverAnnounced

	if ev.GeoIPFeed != nil && ev.GeoIPFeed.Country != "" {
		flags = append(flags, FlagGeoIPValid)
	}
	if inFeed && !announced && !ever {
		flags = append(flags, FlagPrefixOnlyInGeoFeed)
	}
	if announced && !inFeed {
		flags = append(flags, FlagPrefixOnlyInBGP)
	}
	if ever && !announced {
		flags = append(flags, FlagWithdrawnRecently)
	}
	if announced {
		flags = append(flags, FlagBGPAnnounced)
	} else {
		flags = append(flags, FlagBGPNotAnnounced)
	}

	if inFeed && (ev.PoPFeed == nil || ev.PoPFeed.PoPCode == "") {
		flags = append(flags, FlagPopMissing)
	} else if ev.PoPFeed != nil && ev.PoPFeed.PoPCode != "" {
		flags = append(flags, FlagPopPresent)
	}
	if ev.PoPFeed != nil && ev.PoPFeed.PoPCode != "" {
		if ev.PTR == nil || ev.PTR.PoPCode == "" {
			flags = append(flags, FlagPTRMissing)
		} else if ev.PTR.PoPCode == ev.PoPFeed.PoPCode {
			flags = append(flags, FlagPTRMatchesPoP)
		} else {
			flags = append(flags, FlagPTRConflictsWithPoP)
		}
	}

	if announced && ev.BGP != nil && len(ev.BGP.OriginASNs) > 0 {
		if bgpOriginMatches(ev, op) {
			flags = append(flags, FlagOriginASNExpected)
		} else if op != OperatorUnknown {
			flags = append(flags, FlagOriginASNUnexpected)
		}
	}
	if ev.RPKI != nil && ev.RPKI.HasVRP && ev.RPKI.ValidForOrigin != nil && !*ev.RPKI.ValidForOrigin {
		flags = append(flags, FlagRPKIInvalid)
	}
	if ev.GeoIPFeed != nil && ev.GeoIPFeed.Country != "" && ev.RIR != nil && ev.RIR.Country != "" && ev.GeoIPFeed.Country != ev.RIR.Country {
		flags = append(flags, FlagCountryMismatchRIR)
	}
	if ev.GeoIPFeed != nil && ev.GeoIPFeed.Country != "" && ev.GeoIPFeed.City != "" {
		if ref != nil {
			if known, valid := ref.CityCountry(ev.GeoIPFeed.Country, ev.GeoIPFeed.City); known && !valid {
				flags = append(flags, FlagGeoIPInvalidCountryCityPair)
			}
		} else {
			if known, ok := cityCountry[strings.ToLower(strings.TrimSpace(ev.GeoIPFeed.City))]; ok && known != ev.GeoIPFeed.Country {
				flags = append(flags, FlagGeoIPInvalidCountryCityPair)
			}
		}
	}

	return flags
}

func computeAttributionConfidence(ev PrefixEvidence, op Operator, flags []string) float64 {
	if op == OperatorUnknown {
		return 0.1
	}
	signals := []float64{}
	if ev.RIR != nil && OrgMatchesOperator(ev.RIR.Org, op) {
		signals = append(signals, 0.85)
	}
	if ev.BGP != nil && ev.BGP.Announced && bgpOriginMatches(ev, op) {
		signals = append(signals, 0.80)
	}
	if ev.RPKI != nil && ev.RPKI.HasVRP && ev.RPKI.ValidForOrigin != nil && *ev.RPKI.ValidForOrigin {
		signals = append(signals, 0.80)
	}
	if ev.GeoIPFeed != nil {
		signals = append(signals, 0.50)
	}

	conf := noisyOR(signals)
	if hasFlag(flags, FlagOriginASNUnexpected) {
		conf -= 0.40
	}
	if hasFlag(flags, FlagRPKIInvalid) {
		conf -= 0.30
	}
	if len(signals) == 1 && signals[0] == 0.50 {
		conf -= 0.15
	}
	return round3(clamp(conf))
}

func computeGeoConfidence(ev PrefixEvidence, op Operator, flags []string) float64 {
	if ev.GeoIPFeed == nil || ev.GeoIPFeed.Country == "" {
		return 0
	}
	conf := 0.60
	if ev.PoPFeed != nil && ev.PoPFeed.PoPCode != "" {
		conf += 0.15
	}
	if !hasFlag(flags, FlagPTRConflictsWithPoP) && ev.PTR != nil && ev.PTR.PoPCode != "" {
		conf += 0.10
	}
	if hasFlag(flags, FlagPTRConflictsWithPoP) {
		conf -= 0.15
	}
	if hasFlag(flags, FlagCountryMismatchRIR) && op == OperatorUnknown {
		conf -= 0.20
	}
	if hasFlag(flags, FlagGeoIPInvalidCountryCityPair) {
		conf -= 0.55
	}
	return round3(clamp(conf))
}

func bgpOriginMatches(ev PrefixEvidence, op Operator) bool {
	if ev.BGP == nil || op == OperatorUnknown {
		return false
	}
	for _, asn := range ev.BGP.OriginASNs {
		if OperatorForASN(asn) == op {
			return true
		}
	}
	return false
}

func noisyOR(weights []float64) float64 {
	p := 1.0
	for _, w := range weights {
		p *= 1.0 - clamp(w)
	}
	return 1.0 - p
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func hasFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

func sourceIf(ok bool, source string) string {
	if ok {
		return source
	}
	return ""
}

func sourceForOperator(originASN int) string {
	if originASN != 0 {
		return "bgp_origin"
	}
	return "rir_org"
}

func defaultGeoSource(op Operator) string {
	switch op {
	case OperatorStarlink:
		return "starlink_feed_csv"
	case OperatorViasat:
		return "viasat_geofeed_csv"
	default:
		return "operator_geofeed"
	}
}
