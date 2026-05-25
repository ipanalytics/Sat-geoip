package mmdb

import (
	"fmt"
	"net"
	"os"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"

	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

func Write(path string, records []resolver.ResolvedPrefix) error {
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "sat-geoip-Satellite-Internet",
		Description: map[string]string{
			"en": "sat-geoip satellite internet GeoIP PoP BGP intelligence database",
		},
		IPVersion:  6,
		RecordSize: 28,
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		_, network, err := net.ParseCIDR(record.Prefix)
		if err != nil {
			return fmt.Errorf("parse %s: %w", record.Prefix, err)
		}
		value := mmdbtype.Map{
			"network_type":   mmdbtype.String(record.ServiceType),
			"operator":       mmdbtype.String(record.Operator),
			"orbit_class":    mmdbtype.String(record.OrbitClass),
			"geoip_source":   mmdbtype.String(record.GeoIPSource),
			"geoip_country":  mmdbtype.String(record.GeoIPCountry),
			"geoip_region":   mmdbtype.String(record.GeoIPRegion),
			"geoip_city":     mmdbtype.String(record.GeoIPCity),
			"pop_code":       mmdbtype.String(record.PoPCode),
			"pop_iata":       mmdbtype.String(record.PoPIATA),
			"bgp_state":      mmdbtype.String(record.BGPState),
			"data_semantics": mmdbtype.String("satellite_customer_subnet_geoip"),
		}
		if record.OriginASN != nil {
			value["origin_asn"] = mmdbtype.Uint32(uint32(*record.OriginASN))
		}
		if err := tree.Insert(network, value); err != nil {
			return fmt.Errorf("insert %s: %w", record.Prefix, err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = tree.WriteTo(f)
	return err
}
