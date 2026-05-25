# Reference Datasets

These datasets are used for validation only. They do not replace operator
geofeeds and are not used to infer customer, dish, or gateway locations.

| Dataset | File | Purpose | License |
|---|---|---|---|
| GeoNames countryInfo | `geonames/countryInfo.txt` | ISO country reference | Creative Commons Attribution |
| GeoNames admin1CodesASCII | `geonames/admin1CodesASCII.txt` | ISO-like first-level subdivision validation | Creative Commons Attribution |
| GeoNames cities1000 | `geonames/cities1000.txt` | city-to-country sanity checks | Creative Commons Attribution |
| OurAirports airports | `ourairports/airports.csv` | IATA-to-country checks for PoP/gateway metadata | Public domain |

Current pipeline use:

- `geoip_invalid_country_city_pair` validation from GeoNames cities.
- region/country and IATA/country reference APIs for validators.
- gateway country reference remains metadata only and is never treated as IP GeoIP.

Sources:

- GeoNames: https://download.geonames.org/export/dump/
- GeoNames license/about: https://www.geonames.org/about.html
- OurAirports data: https://ourairports.com/data/
- OurAirports CSV mirror: https://davidmegginson.github.io/ourairports-data/

