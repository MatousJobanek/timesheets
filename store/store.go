package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"timesheets/model"
)

const storeFileName = "store.json"

type Store struct {
	Employees   []model.Employee   `json:"employees"`
	YearConfigs []model.YearConfig `json:"year_configs"`
}

func Load(dir string) (*Store, error) {
	path := filepath.Join(dir, storeFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{}, nil
		}
		return nil, fmt.Errorf("store lesen fehlgeschlagen: %w", err)
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("store JSON ungültig: %w", err)
	}
	return &s, nil
}

func Save(dir string, s *Store) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("store serialisieren fehlgeschlagen: %w", err)
	}
	path := filepath.Join(dir, storeFileName)
	tmp, err := os.CreateTemp(dir, "store-*.json.tmp")
	if err != nil {
		return fmt.Errorf("temp-Datei erstellen fehlgeschlagen: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("temp-Datei schreiben fehlgeschlagen: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("temp-Datei schließen fehlgeschlagen: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("store ersetzen fehlgeschlagen: %w", err)
	}
	return nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func AddEmployee(s *Store, emp model.Employee) string {
	emp.ID = generateID()
	s.Employees = append(s.Employees, emp)
	EnsureYearConfig(s, emp.Year)
	return emp.ID
}

func UpdateEmployee(s *Store, emp model.Employee) {
	for i, e := range s.Employees {
		if e.ID == emp.ID {
			s.Employees[i] = emp
			return
		}
	}
}

func DeleteEmployee(s *Store, id string) {
	for i, e := range s.Employees {
		if e.ID == id {
			s.Employees = append(s.Employees[:i], s.Employees[i+1:]...)
			return
		}
	}
}

func FindYearConfig(s *Store, year int) *model.YearConfig {
	for i := range s.YearConfigs {
		if s.YearConfigs[i].Year == year {
			return &s.YearConfigs[i]
		}
	}
	return nil
}

func EnsureYearConfig(s *Store, year int) *model.YearConfig {
	if yc := FindYearConfig(s, year); yc != nil {
		return yc
	}
	s.YearConfigs = append(s.YearConfigs, model.YearConfig{Year: year})
	return &s.YearConfigs[len(s.YearConfigs)-1]
}

func MergeFreePeriods(shared, overrides []model.DateRange) []model.DateRange {
	merged := make([]model.DateRange, 0, len(shared)+len(overrides))
	merged = append(merged, shared...)
	merged = append(merged, overrides...)
	return merged
}
