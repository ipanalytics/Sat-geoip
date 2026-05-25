package site

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ipanalytics/Sat-geoip/internal/release"
	"github.com/ipanalytics/Sat-geoip/internal/resolver"
)

type Dashboard struct {
	GeneratedAt string         `json:"generated_at"`
	Stats       release.Stats  `json:"stats"`
	Countries   []CountryPoint `json:"countries"`
	PoPs        []PoPPoint     `json:"pops"`
	Gateways    []GatewayPoint `json:"gateways"`
	Operators   []OperatorInfo `json:"operators"`
	Orbits      []OrbitInfo    `json:"orbits"`
}

type CountryPoint struct {
	Country   string         `json:"country"`
	Lat       float64        `json:"lat"`
	Lon       float64        `json:"lon"`
	Prefixes  int            `json:"prefixes"`
	Announced int            `json:"announced"`
	Operators map[string]int `json:"operators"`
}

type PoPPoint struct {
	Code      string         `json:"code"`
	IATA      string         `json:"iata"`
	Lat       float64        `json:"lat"`
	Lon       float64        `json:"lon"`
	Country   string         `json:"country"`
	Prefixes  int            `json:"prefixes"`
	Operators map[string]int `json:"operators"`
}

type GatewayPoint struct {
	Operator  string  `json:"operator"`
	Country   string  `json:"country"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Semantics string  `json:"semantics"`
	Notes     string  `json:"notes"`
}

type OperatorInfo struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	OrbitClass string `json:"orbit_class"`
}

type OrbitInfo struct {
	Operator    string  `json:"operator"`
	OrbitClass  string  `json:"orbit_class"`
	Inclination float64 `json:"inclination"`
	AltitudeKM  float64 `json:"altitude_km"`
	Phase       float64 `json:"phase"`
	Color       string  `json:"color"`
}

type GenerateOptions struct {
	RecordsPath   string
	StatsPath     string
	GatewaysPath  string
	ReferenceRoot string
	OutDir        string
}

type coord struct {
	lat float64
	lon float64
}

type airport struct {
	coord
	country string
}

func GenerateDashboard(opts GenerateOptions) error {
	records, err := readRecords(opts.RecordsPath)
	if err != nil {
		return err
	}
	stats, err := readStats(opts.StatsPath)
	if err != nil {
		return err
	}
	countries, err := readCountryCentroids(filepath.Join(opts.ReferenceRoot, "geonames", "cities1000.txt"))
	if err != nil {
		return err
	}
	airports, err := readAirports(filepath.Join(opts.ReferenceRoot, "ourairports", "airports.csv"))
	if err != nil {
		return err
	}
	gateways, err := readGateways(opts.GatewaysPath, countries)
	if err != nil {
		return err
	}

	dashboard := Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Stats:       stats,
		Countries:   buildCountryPoints(records, countries),
		PoPs:        buildPoPPoints(records, airports),
		Gateways:    gateways,
		Operators:   buildOperators(stats, records),
		Orbits:      defaultOrbits(),
	}
	sort.Slice(dashboard.Countries, func(i, j int) bool {
		return dashboard.Countries[i].Prefixes > dashboard.Countries[j].Prefixes
	})
	sort.Slice(dashboard.PoPs, func(i, j int) bool {
		return dashboard.PoPs[i].Prefixes > dashboard.PoPs[j].Prefixes
	})

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(opts.OutDir, "dashboard.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(dashboard)
}

func readRecords(path string) ([]resolver.ResolvedPrefix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []resolver.ResolvedPrefix
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 32*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record resolver.ResolvedPrefix
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, scanner.Err()
}

func readStats(path string) (release.Stats, error) {
	f, err := os.Open(path)
	if err != nil {
		return release.Stats{}, err
	}
	defer f.Close()
	var stats release.Stats
	return stats, json.NewDecoder(f).Decode(&stats)
}

func readCountryCentroids(path string) (map[string]coord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	type acc struct {
		lat float64
		lon float64
		n   float64
	}
	points := map[string]acc{}
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(row) < 9 {
			continue
		}
		country := strings.ToUpper(strings.TrimSpace(row[8]))
		lat, latErr := strconv.ParseFloat(row[4], 64)
		lon, lonErr := strconv.ParseFloat(row[5], 64)
		if len(country) != 2 || latErr != nil || lonErr != nil {
			continue
		}
		got := points[country]
		got.lat += lat
		got.lon += lon
		got.n++
		points[country] = got
	}
	out := map[string]coord{}
	for country, got := range points {
		if got.n > 0 {
			out[country] = coord{lat: got.lat / got.n, lon: got.lon / got.n}
		}
	}
	return out, nil
}

func readAirports(path string) (map[string]airport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, field := range header {
		idx[field] = i
	}
	required := []string{"iata_code", "iso_country", "latitude_deg", "longitude_deg"}
	for _, field := range required {
		if _, ok := idx[field]; !ok {
			return nil, fmt.Errorf("airports.csv missing %s", field)
		}
	}
	out := map[string]airport{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		iata := strings.ToUpper(strings.TrimSpace(row[idx["iata_code"]]))
		country := strings.ToUpper(strings.TrimSpace(row[idx["iso_country"]]))
		lat, latErr := strconv.ParseFloat(row[idx["latitude_deg"]], 64)
		lon, lonErr := strconv.ParseFloat(row[idx["longitude_deg"]], 64)
		if len(iata) == 3 && len(country) == 2 && latErr == nil && lonErr == nil {
			out[iata] = airport{coord: coord{lat: lat, lon: lon}, country: country}
		}
	}
	return out, nil
}

func readGateways(path string, countries map[string]coord) ([]GatewayPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, field := range header {
		idx[field] = i
	}
	var out []GatewayPoint
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		country := strings.ToUpper(strings.TrimSpace(row[idx["country"]]))
		c, ok := countries[country]
		if !ok {
			continue
		}
		out = append(out, GatewayPoint{
			Operator:  row[idx["operator"]],
			Country:   country,
			Lat:       round(c.lat),
			Lon:       round(c.lon),
			Semantics: row[idx["semantics"]],
			Notes:     row[idx["notes"]],
		})
	}
	return out, nil
}

func buildCountryPoints(records []resolver.ResolvedPrefix, centroids map[string]coord) []CountryPoint {
	type acc struct {
		prefixes  int
		announced int
		operators map[string]int
	}
	counts := map[string]acc{}
	for _, record := range records {
		country := strings.ToUpper(strings.TrimSpace(record.GeoIPCountry))
		if country == "" {
			continue
		}
		got := counts[country]
		got.prefixes++
		if record.ActiveUserClaim {
			got.announced++
		}
		if got.operators == nil {
			got.operators = map[string]int{}
		}
		got.operators[record.Operator]++
		counts[country] = got
	}
	out := make([]CountryPoint, 0, len(counts))
	for country, got := range counts {
		c, ok := centroids[country]
		if !ok {
			continue
		}
		out = append(out, CountryPoint{
			Country:   country,
			Lat:       round(c.lat),
			Lon:       round(c.lon),
			Prefixes:  got.prefixes,
			Announced: got.announced,
			Operators: got.operators,
		})
	}
	return out
}

func buildPoPPoints(records []resolver.ResolvedPrefix, airports map[string]airport) []PoPPoint {
	type acc struct {
		iata      string
		prefixes  int
		operators map[string]int
	}
	counts := map[string]acc{}
	for _, record := range records {
		if record.PoPCode == "" || record.PoPIATA == "" {
			continue
		}
		iata := strings.ToUpper(record.PoPIATA)
		if _, ok := airports[iata]; !ok {
			continue
		}
		got := counts[record.PoPCode]
		got.iata = iata
		got.prefixes++
		if got.operators == nil {
			got.operators = map[string]int{}
		}
		got.operators[record.Operator]++
		counts[record.PoPCode] = got
	}
	out := make([]PoPPoint, 0, len(counts))
	for code, got := range counts {
		a := airports[got.iata]
		out = append(out, PoPPoint{
			Code:      code,
			IATA:      got.iata,
			Lat:       round(a.lat),
			Lon:       round(a.lon),
			Country:   a.country,
			Prefixes:  got.prefixes,
			Operators: got.operators,
		})
	}
	return out
}

func buildOperators(stats release.Stats, records []resolver.ResolvedPrefix) []OperatorInfo {
	orbits := map[string]string{}
	for _, record := range records {
		if record.Operator != "" && record.OrbitClass != "" {
			orbits[record.Operator] = record.OrbitClass
		}
	}
	out := make([]OperatorInfo, 0, len(stats.Operators))
	for op, count := range stats.Operators {
		out = append(out, OperatorInfo{Name: op, Count: count, OrbitClass: orbits[op]})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

func defaultOrbits() []OrbitInfo {
	return []OrbitInfo{
		{Operator: "starlink", OrbitClass: "leo", Inclination: 53, AltitudeKM: 550, Phase: 0, Color: "#38bdf8"},
		{Operator: "oneweb", OrbitClass: "leo", Inclination: 87.9, AltitudeKM: 1200, Phase: 115, Color: "#8b5cf6"},
		{Operator: "ses_o3b", OrbitClass: "meo", Inclination: 0, AltitudeKM: 8063, Phase: 210, Color: "#14b8a6"},
		{Operator: "viasat", OrbitClass: "geo_or_hybrid_satellite", Inclination: 0, AltitudeKM: 35786, Phase: 300, Color: "#f59e0b"},
		{Operator: "hughes", OrbitClass: "geo", Inclination: 0, AltitudeKM: 35786, Phase: 30, Color: "#ef4444"},
	}
}

func round(v float64) float64 {
	return math.Round(v*10000) / 10000
}
