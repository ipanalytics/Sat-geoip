package reference

import "testing"

func TestLoadReferenceData(t *testing.T) {
	data, err := Load("../../data/reference")
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
