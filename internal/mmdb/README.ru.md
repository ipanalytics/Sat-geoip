_English version: [README.md](README.md)_

# Запись в MMDB

Экспортёр `.mmdb` реализован на Go с использованием официального пакета MaxMind
`github.com/maxmind/mmdbwriter`. Он изолирован от резолвера, чтобы логика
разрешения оставалась независимой от формата вывода базы данных.

Минимальная структура записи:

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
