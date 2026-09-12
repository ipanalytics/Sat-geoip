_English version: [README.md](README.md)_

# Справочные наборы данных

Эти наборы данных используются только для валидации. Они не заменяют геофиды
оператора (geofeeds) и не используются для определения местоположения клиентов,
терминалов (dish) или шлюзов.

| Набор данных | Файл | Назначение | Лицензия |
|---|---|---|---|
| GeoNames countryInfo | `geonames/countryInfo.txt` | справочник стран ISO | Creative Commons Attribution |
| GeoNames admin1CodesASCII | `geonames/admin1CodesASCII.txt` | валидация единиц административного деления первого уровня в формате, подобном ISO | Creative Commons Attribution |
| GeoNames cities1000 | `geonames/cities1000.txt` | проверки корректности соответствия «город — страна» | Creative Commons Attribution |
| OurAirports airports | `ourairports/airports.csv` | проверка соответствия «IATA — страна» для метаданных PoP/шлюзов | Общественное достояние |

Использование в текущем конвейере (pipeline):

- валидация `geoip_invalid_country_city_pair` на основе городов GeoNames.
- справочные API регион/страна и IATA/страна для валидаторов.
- справочник стран шлюзов остаётся только метаданными и никогда не трактуется как IP GeoIP.

Источники:

- GeoNames: https://download.geonames.org/export/dump/
- Лицензия/о проекте GeoNames: https://www.geonames.org/about.html
- Данные OurAirports: https://ourairports.com/data/
- Зеркало CSV OurAirports: https://davidmegginson.github.io/ourairports-data/
