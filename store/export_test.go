package store

import (
	"encoding/json"
	"strings"
	"testing"

	"timesheets/model"
)

func TestExportJSONIsPortableEnvelope(t *testing.T) {
	s := &Store{
		Employees: []model.Employee{{ID: "abc", Name: "Anna", Year: 2026}},
		YearConfigs: []model.YearConfig{
			{Year: 2026, FreePeriods: []model.DateRange{{Label: "Ostern"}}},
		},
	}
	data, err := ExportJSON(s)
	if err != nil {
		t.Fatal(err)
	}
	raw := string(data)
	if strings.Contains(raw, "com.example") {
		t.Fatalf("export must not contain an app ID: %s", raw)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["format"] != setupFormat {
		t.Fatalf("format = %v", doc["format"])
	}
	if int(doc["format_version"].(float64)) != setupFormatVersion {
		t.Fatalf("format_version = %v", doc["format_version"])
	}
}

func TestImportJSONEnvelopeAndLegacy(t *testing.T) {
	envelope := []byte(`{
  "format": "timesheets-setup",
  "format_version": 1,
  "employees": [{"id":"1","name":"Anna","year":2026}],
  "year_configs": [{"year":2026}]
}`)
	s, err := ImportJSON(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Employees) != 1 || s.Employees[0].Name != "Anna" {
		t.Fatalf("employees = %+v", s.Employees)
	}
	if len(s.YearConfigs) != 1 || s.YearConfigs[0].Year != 2026 {
		t.Fatalf("year_configs = %+v", s.YearConfigs)
	}

	legacy := []byte(`{"employees":[{"id":"2","name":"Bert","year":2025}],"year_configs":[]}`)
	s, err = ImportJSON(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Employees) != 1 || s.Employees[0].Name != "Bert" {
		t.Fatalf("legacy employees = %+v", s.Employees)
	}
}

func TestImportJSONRejectsUnknownFormat(t *testing.T) {
	_, err := ImportJSON([]byte(`{"format":"other","format_version":1}`))
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = ImportJSON([]byte(`{"format":"timesheets-setup","format_version":99}`))
	if err == nil {
		t.Fatal("expected version error")
	}
}
