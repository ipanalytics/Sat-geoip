package reference

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReferenceData(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "geonames/countryInfo.txt", "US\tUSA\t840\tUS\tUnited States\nPH\tPHL\t608\tRP\tPhilippines\nSA\tSAU\t682\tSA\tSaudi Arabia\nLK\tLKA\t144\tCE\tSri Lanka\n")
	writeFixture(t, root, "geonames/admin1CodesASCII.txt", "US.WA\tWashington\tWashington\t5815135\nPH.00\tMetro Manila\tMetro Manila\t7521306\n")
	writeFixture(t, root, "geonames/cities1000.txt", "108410\tRiyadh\tRiyadh\tRiyadh\t24.68773\t46.72185\tP\tPPLC\tSA\t\t10\t\t\t4205961\t612\t612\tAsia/Riyadh\t2024-01-01\n1701668\tManila\tManila\tManila\t14.6042\t120.9822\tP\tPPLC\tPH\t\t00\t\t\t1600000\t13\t13\tAsia/Manila\t2024-01-01\n")
	writeFixture(t, root, "ourairports/airports.csv", "\"id\",\"ident\",\"type\",\"name\",\"latitude_deg\",\"longitude_deg\",\"elevation_ft\",\"continent\",\"iso_country\",\"iso_region\",\"municipality\",\"scheduled_service\",\"icao_code\",\"iata_code\",\"gps_code\",\"local_code\",\"home_link\",\"wikipedia_link\",\"keywords\"\n1,\"RPLL\",\"large_airport\",\"Ninoy Aquino International Airport\",14.5086,121.0198,75,\"AS\",\"PH\",\"PH-00\",\"Manila\",\"yes\",\"RPLL\",\"MNL\",\"RPLL\",,,,\n2,\"KSEA\",\"large_airport\",\"Seattle Tacoma International Airport\",47.449, -122.309,433,\"NA\",\"US\",\"US-WA\",\"Seattle\",\"yes\",\"KSEA\",\"SEA\",\"KSEA\",,,,\n")

	data, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}

	if known, valid := data.CityCountry("SA", "Riyadh"); !known || !valid {
		t.Fatalf("Riyadh should validate as SA, known=%v valid=%v", known, valid)
	}
	if known, valid := data.CityCountry("LK", "Riyadh"); !known || valid {
		t.Fatalf("Riyadh should not validate as LK, known=%v valid=%v", known, valid)
	}
	if known, valid := data.RegionCountry("US", "US-WA"); !known || !valid {
		t.Fatalf("US-WA should validate as US, known=%v valid=%v", known, valid)
	}
	if country, ok := data.IATACountry("SEA"); !ok || country != "US" {
		t.Fatalf("SEA country = %q, ok=%v", country, ok)
	}
	if country, ok := data.IATACountry("MNL"); !ok || country != "PH" {
		t.Fatalf("MNL country = %q, ok=%v", country, ok)
	}
}

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
