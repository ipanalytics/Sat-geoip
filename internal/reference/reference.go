package reference

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Data struct {
	Countries map[string]struct{}
	Regions   map[string]string
	Cities    map[string]map[string]struct{}
	IATA      map[string]string
}

func Load(root string) (*Data, error) {
	data := &Data{
		Countries: map[string]struct{}{},
		Regions:   map[string]string{},
		Cities:    map[string]map[string]struct{}{},
		IATA:      map[string]string{},
	}
	if err := data.loadCountries(filepath.Join(root, "geonames", "countryInfo.txt")); err != nil {
		return nil, err
	}
	if err := data.loadRegions(filepath.Join(root, "geonames", "admin1CodesASCII.txt")); err != nil {
		return nil, err
	}
	if err := data.loadCities(filepath.Join(root, "geonames", "cities1000.txt")); err != nil {
		return nil, err
	}
	if err := data.loadAirports(filepath.Join(root, "ourairports", "airports.csv")); err != nil {
		return nil, err
	}
	return data, nil
}

func LoadIfAvailable(root string) (*Data, error) {
	data, err := Load(root)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, err
}

func (d *Data) CityCountry(country, city string) (bool, bool) {
	city = normalize(city)
	country = strings.ToUpper(strings.TrimSpace(country))
	if city == "" || country == "" {
		return false, false
	}
	countries, ok := d.Cities[city]
	if !ok {
		return false, false
	}
	_, valid := countries[country]
	return true, valid
}

func (d *Data) RegionCountry(country, region string) (bool, bool) {
	country = strings.ToUpper(strings.TrimSpace(country))
	region = strings.ToUpper(strings.TrimSpace(region))
	if country == "" || region == "" {
		return false, false
	}
	regionCountry, ok := d.Regions[region]
	if !ok {
		return false, false
	}
	return true, regionCountry == country
}

func (d *Data) IATACountry(iata string) (string, bool) {
	country, ok := d.IATA[strings.ToUpper(strings.TrimSpace(iata))]
	return country, ok
}

func (d *Data) loadCountries(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return readLines(f, func(fields []string) error {
		if len(fields) < 1 || strings.HasPrefix(fields[0], "#") {
			return nil
		}
		code := strings.ToUpper(strings.TrimSpace(fields[0]))
		if len(code) == 2 {
			d.Countries[code] = struct{}{}
		}
		return nil
	})
}

func (d *Data) loadRegions(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return readLines(f, func(fields []string) error {
		if len(fields) < 1 || strings.HasPrefix(fields[0], "#") {
			return nil
		}
		parts := strings.SplitN(strings.ToUpper(strings.TrimSpace(fields[0])), ".", 2)
		if len(parts) != 2 {
			return nil
		}
		d.Regions[parts[0]+"-"+parts[1]] = parts[0]
		return nil
	})
}

func (d *Data) loadCities(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return readLines(f, func(fields []string) error {
		if len(fields) < 9 {
			return nil
		}
		country := strings.ToUpper(strings.TrimSpace(fields[8]))
		if len(country) != 2 {
			return nil
		}
		for _, name := range []string{fields[1], fields[2]} {
			d.addCity(name, country)
		}
		return nil
	})
}

func (d *Data) loadAirports(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return err
	}
	index := map[string]int{}
	for i, field := range header {
		index[field] = i
	}
	iataIdx, ok := index["iata_code"]
	if !ok {
		return fmt.Errorf("airports.csv missing iata_code")
	}
	countryIdx, ok := index["iso_country"]
	if !ok {
		return fmt.Errorf("airports.csv missing iso_country")
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		iata := strings.ToUpper(strings.TrimSpace(record[iataIdx]))
		country := strings.ToUpper(strings.TrimSpace(record[countryIdx]))
		if len(iata) == 3 && len(country) == 2 {
			d.IATA[iata] = country
		}
	}
	return nil
}

func (d *Data) addCity(city, country string) {
	city = normalize(city)
	if city == "" {
		return
	}
	if d.Cities[city] == nil {
		d.Cities[city] = map[string]struct{}{}
	}
	d.Cities[city][country] = struct{}{}
}

func readLines(r io.Reader, fn func([]string) error) error {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := fn(record); err != nil {
			return err
		}
	}
	return nil
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
