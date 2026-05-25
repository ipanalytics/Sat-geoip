package mmdb

import (
	"net/netip"
	"path/filepath"
	"testing"

	maxminddb "github.com/oschwald/maxminddb-golang/v2"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func TestWriteReadableMMDB(t *testing.T) {
	asn := 45700
	path := filepath.Join(t.TempDir(), "sat-geoip.mmdb")
	err := Write(path, []resolver.ResolvedPrefix{
		{
			Prefix:       "14.1.64.0/24",
			Operator:     "starlink",
			ServiceType:  "satellite_internet",
			OrbitClass:   "leo",
			OriginASN:    &asn,
			GeoIPSource:  "starlink_feed_csv",
			GeoIPCountry: "PH",
			GeoIPCity:    "Manila",
			PoPCode:      "mnlaphl1",
			PoPIATA:      "mnl",
			BGPState:     "announced",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	reader, err := maxminddb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	var got struct {
		Operator     string `maxminddb:"operator"`
		OriginASN    uint32 `maxminddb:"origin_asn"`
		GeoIPCountry string `maxminddb:"geoip_country"`
		BGPState     string `maxminddb:"bgp_state"`
	}
	if err := reader.Lookup(netip.MustParseAddr("14.1.64.1")).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Operator != "starlink" || got.OriginASN != 45700 || got.GeoIPCountry != "PH" || got.BGPState != "announced" {
		t.Fatalf("unexpected mmdb record: %#v", got)
	}
}
