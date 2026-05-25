# MMDB Writer

The `.mmdb` exporter is implemented in Go with MaxMind's official
`github.com/maxmind/mmdbwriter` package. It is isolated from the resolver so
resolution logic stays independent from the database output format.

Minimal record shape:

```json
{
  "network_type": "satellite_internet",
  "operator": "starlink",
  "orbit_class": "leo",
  "origin_asn": 14593,
  "geoip_source": "operator_geofeed",
  "geoip_country": "US",
  "geoip_region": "US-WA",
  "geoip_city": "Seattle",
  "pop_code": "sttlwax1",
  "pop_iata": "sea",
  "bgp_state": "announced",
  "data_semantics": "satellite_customer_subnet_geoip"
}
```
