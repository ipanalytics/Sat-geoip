package resolver

import "testing"

func TestHealthyStarlinkManila(t *testing.T) {
	valid := true
	record := Resolve(PrefixEvidence{
		Prefix:    "14.1.64.0/24",
		GeoIPFeed: &GeoIPFeedEvidence{Country: "PH", City: "Manila", Source: "starlink_feed_csv"},
		PoPFeed:   &PopFeedEvidence{PoPCode: "mnlaphl1", PoPIATA: "mnl", Source: "starlink_pops_csv"},
		PTR:       &PTREvidence{Record: "customer.mnlaphl1.pop.starlinkisp.net", PoPCode: "mnlaphl1"},
		BGP:       &BGPEvidence{Announced: true, OriginASNs: []int{45700}, EverAnnounced: true},
		RIR:       &RIREvidence{Org: "PT Starlink Services Indonesia", Country: "ID"},
		RPKI:      &RPKIEvidence{HasVRP: true, ValidForOrigin: &valid},
	})

	if record.BGPState != string(BGPAnnounced) {
		t.Fatalf("BGPState = %q", record.BGPState)
	}
	if !record.ActiveUserClaim {
		t.Fatal("ActiveUserClaim must be true only for live BGP")
	}
	if record.GroundStationClaim {
		t.Fatal("GroundStationClaim must never be true")
	}
	assertFlag(t, record, FlagPTRMatchesPoP)
	assertFlag(t, record, FlagOriginASNExpected)
	assertFlag(t, record, FlagCountryMismatchRIR)
	if record.DataConfidence.Attribution < 0.99 {
		t.Fatalf("attribution confidence too low: %.3f", record.DataConfidence.Attribution)
	}
	if record.DataConfidence.Geo < 0.80 {
		t.Fatalf("known satellite RIR/feed country mismatch should be informational; geo=%.3f", record.DataConfidence.Geo)
	}
}

func TestStarlinkPlannedVancouverStalePTR(t *testing.T) {
	record := Resolve(PrefixEvidence{
		Prefix:    "170.203.201.0/24",
		GeoIPFeed: &GeoIPFeedEvidence{Country: "CA", Region: "CA-BC", City: "Vancouver"},
		PoPFeed:   &PopFeedEvidence{PoPCode: "seataws1", PoPIATA: "sea"},
		PTR:       &PTREvidence{Record: "customer.yvr.pop.starlinkisp.net", PoPCode: "yvrbcca1"},
		BGP:       &BGPEvidence{Announced: false, OriginASNs: nil, EverAnnounced: false},
		RIR:       &RIREvidence{Org: "Space Exploration Technologies", Country: "US"},
		RPKI:      &RPKIEvidence{HasVRP: true},
	})

	if record.BGPState != string(BGPPlannedNotAnnounced) {
		t.Fatalf("BGPState = %q", record.BGPState)
	}
	if record.ActiveUserClaim {
		t.Fatal("feed-only prefixes are not active user proof")
	}
	if record.PoPCode != "seataws1" {
		t.Fatalf("official PoP feed must win over PTR, got %q", record.PoPCode)
	}
	assertFlag(t, record, FlagPTRConflictsWithPoP)
	assertFlag(t, record, FlagPrefixOnlyInGeoFeed)
	if record.DataConfidence.Geo >= 0.75 {
		t.Fatalf("PTR conflict should reduce geo confidence, got %.3f", record.DataConfidence.Geo)
	}
}

func TestViasatFeedErrorSeparatesAttributionAndGeo(t *testing.T) {
	valid := true
	record := Resolve(PrefixEvidence{
		Prefix:    "184.63.131.8/29",
		GeoIPFeed: &GeoIPFeedEvidence{Country: "LK", City: "Riyadh", Source: "viasat_geofeed_csv"},
		BGP:       &BGPEvidence{Announced: true, OriginASNs: []int{7155}, EverAnnounced: true},
		RIR:       &RIREvidence{Org: "ViaSat, Inc.", Country: "US"},
		RPKI:      &RPKIEvidence{HasVRP: true, ValidForOrigin: &valid},
	})

	assertFlag(t, record, FlagGeoIPInvalidCountryCityPair)
	assertFlag(t, record, FlagOriginASNExpected)
	if record.DataConfidence.Attribution < 0.99 {
		t.Fatalf("attribution should remain high, got %.3f", record.DataConfidence.Attribution)
	}
	if record.DataConfidence.Geo > 0.10 {
		t.Fatalf("invalid country/city pair should heavily reduce geo confidence, got %.3f", record.DataConfidence.Geo)
	}
}

func TestBGPOnlyPrefixIsStillLiveBGP(t *testing.T) {
	record := Resolve(PrefixEvidence{
		Prefix: "129.222.10.0/24",
		BGP:    &BGPEvidence{Announced: true, OriginASNs: []int{14593}, EverAnnounced: true},
		RIR:    &RIREvidence{Org: "Space Exploration Technologies"},
	})

	if record.BGPState != string(BGPOnly) {
		t.Fatalf("BGPState = %q", record.BGPState)
	}
	if !record.ActiveUserClaim {
		t.Fatal("live BGP-only prefixes must set active_user_claim")
	}
	assertFlag(t, record, FlagPrefixOnlyInBGP)
}

func TestNewBGPDerivedSatelliteOperators(t *testing.T) {
	cases := []struct {
		name        string
		asn         int
		wantOp      string
		wantGroup   string
		wantService string
		wantOrbit   string
	}{
		{
			name:        "ses_o3b",
			asn:         60725,
			wantOp:      string(OperatorSESO3B),
			wantGroup:   "ses",
			wantService: "satellite_internet",
			wantOrbit:   string(OrbitMEO),
		},
		{
			name:        "marlink",
			asn:         55784,
			wantOp:      string(OperatorMarlink),
			wantGroup:   "marlink",
			wantService: "satellite_service_provider",
			wantOrbit:   string(OrbitMixed),
		},
		{
			name:        "hughes",
			asn:         63062,
			wantOp:      string(OperatorHughes),
			wantGroup:   "echostar",
			wantService: "satellite_internet",
			wantOrbit:   string(OrbitGEO),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := Resolve(PrefixEvidence{
				Prefix: "203.0.113.0/24",
				BGP:    &BGPEvidence{Announced: true, OriginASNs: []int{tc.asn}, EverAnnounced: true},
			})
			if record.Operator != tc.wantOp {
				t.Fatalf("operator = %q, want %q", record.Operator, tc.wantOp)
			}
			if record.OperatorGroup != tc.wantGroup {
				t.Fatalf("operator_group = %q, want %q", record.OperatorGroup, tc.wantGroup)
			}
			if record.ServiceType != tc.wantService {
				t.Fatalf("service_type = %q, want %q", record.ServiceType, tc.wantService)
			}
			if record.OrbitClass != tc.wantOrbit {
				t.Fatalf("orbit_class = %q, want %q", record.OrbitClass, tc.wantOrbit)
			}
			if record.GeoIPSemantics != GeoIPSemanticsNone {
				t.Fatalf("BGP-derived operators without geofeed must not synthesize GeoIP semantics")
			}
			if !record.ActiveUserClaim {
				t.Fatalf("live BGP-derived prefix must be marked active in BGP")
			}
		})
	}
}

func assertFlag(t *testing.T, record ResolvedPrefix, want string) {
	t.Helper()
	for _, got := range record.QualityFlags {
		if got == want {
			return
		}
	}
	t.Fatalf("missing flag %q in %#v", want, record.QualityFlags)
}
