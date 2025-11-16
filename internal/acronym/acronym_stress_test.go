package acronym

import (
	"encoding/json"
	"sort"
	"testing"
)

// reuse the small file helper from the integration test (same package)

func collectNamesFromTestdata(t *testing.T) []string {
	candidates := []string{"../../testdata/filtros.json", "../testdata/filtros.json", "testdata/filtros.json", "filtros.json"}
	data, err := readFirstExistingFile(candidates)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}
	names := make([]string, 0, 512)
	if m, ok := raw["filtros"].(map[string]interface{}); ok {
		for k := range m {
			names = append(names, k)
		}
	}
	if m, ok := raw["tipos-filtros"].(map[string]interface{}); ok {
		for k := range m {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	return names
}

// Test that generating shorts for a stable sequence is deterministic and unique
func TestStress_DeterministicAndUnique(t *testing.T) {
	names := collectNamesFromTestdata(t)
	if len(names) == 0 {
		t.Fatal("no names found in testdata to run stress test")
	}

	run := func() map[string]string {
		used := make(map[string]bool)
		out := make(map[string]string, len(names))
		for _, name := range names {
			s, _, err := TryExtractSignificant(name, used)
			if err != nil {
				t.Fatalf("generator error for %s: %v", name, err)
			}
			if s == "" {
				t.Fatalf("generator returned empty short for %s", name)
			}
			if _, ok := out[name]; ok {
				t.Fatalf("duplicate generation for same name in run: %s", name)
			}
			out[name] = s
		}
		// ensure uniqueness among values
		seen := make(map[string]bool, len(out))
		for _, v := range out {
			if seen[v] {
				t.Fatalf("duplicate short found across names: %s", v)
			}
			seen[v] = true
		}
		return out
	}

	a := run()
	b := run()

	if len(a) != len(b) {
		t.Fatalf("mismatched result sizes: %d vs %d", len(a), len(b))
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || vb != va {
			t.Fatalf("non-deterministic output for %s: %q vs %q", k, va, vb)
		}
	}
}

// Test that calling TryExtractSignificant twice for the same name in the same
// used map produces different shorts (collision resolution behavior).
func TestStress_DuplicateInputsResolve(t *testing.T) {
	name := "Alimentação"
	used := make(map[string]bool)
	s1, _, err := TryExtractSignificant(name, used)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	s2, _, err := TryExtractSignificant(name, used)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if s1 == "" || s2 == "" {
		t.Fatalf("empty short returned for duplicate test: %q / %q", s1, s2)
	}
	if s1 == s2 {
		t.Fatalf("duplicate resolution failed, both equal: %s", s1)
	}
}
