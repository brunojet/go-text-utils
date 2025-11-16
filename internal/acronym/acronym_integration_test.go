package acronym

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
)

// helper: try a few candidate paths to load the testdata file
func readFirstExistingFile(candidates []string) ([]byte, error) {
	var lastErr error
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return nil, os.ErrNotExist
	}
	return nil, lastErr
}

// extractObjectMap like in main.go
func extractObjectMap(raw map[string]interface{}, key string) map[string]map[string]interface{} {
	res := make(map[string]map[string]interface{})
	if m, ok := raw[key].(map[string]interface{}); ok {
		for k, v := range m {
			if obj, ok := v.(map[string]interface{}); ok {
				res[k] = obj
			}
		}
	}
	return res
}

// determineTypeKey mirrors main.go's heuristic
func determineTypeKey(filters map[string]map[string]interface{}, filterTypes map[string]map[string]interface{}, name string) string {
	if _, ok := filterTypes[name]; ok {
		return name
	}
	if f, ok := filters[name]; ok {
		if t, ok2 := f["tipo"]; ok2 {
			return fmt.Sprintf("%v", t)
		}
	}
	return "_GLOBAL"
}

// getHierarchy mirrors main.go's recursive parent walk
func getHierarchy(filters map[string]map[string]interface{}, name string) []string {
	f, ok := filters[name]
	if !ok {
		return []string{name}
	}
	typ := fmt.Sprintf("%v", f["tipo"])
	parent := fmt.Sprintf("%v", f["pai"])
	if parent != "<nil>" && parent != "" && parent != "null" {
		return append(getHierarchy(filters, parent), name)
	}
	return []string{typ, name}
}

// ordenar helper
func ordenar(lista []string) []string {
	res := make([]string, len(lista))
	copy(res, lista)
	sort.Strings(res)
	return res
}

// ensureShortForPart: similar to main's logic but local to test
func ensureShortForPart(part string, level int, levelUsed map[string]bool, nameToShort map[string]string, nameToMode map[string]string) (string, error) {
	if prev, ok := nameToShort[part]; ok {
		if prev == "" {
			return "", fmt.Errorf("existing short for '%s' is empty", part)
		}
		levelUsed[prev] = true
		return prev, nil
	}
	short, genType, err := TryExtractSignificant(part, levelUsed)
	if err != nil {
		return "", fmt.Errorf("failed to generate short for '%s' at level %d: %w", part, level, err)
	}
	if short == "" {
		return "", fmt.Errorf("generator returned empty short for '%s' at level %d", part, level)
	}
	nameToShort[part] = short
	nameToMode[part] = genType
	levelUsed[short] = true
	return short, nil
}

// NOTE: collision resolution should be the responsibility of the acronym
// generator (TryExtractSignificant). The test will NOT implement its own
// deterministic suffix probing; instead, when a composed short collides we
// ask TryExtractSignificant to produce a unique short for the full name using
// the per-type usedKeys map.

func TestIntegration_ReplicateMainLogicWithTestdata(t *testing.T) {
	candidates := []string{"../../testdata/filtros.json", "../testdata/filtros.json", "testdata/filtros.json", "filtros.json"}
	data, err := readFirstExistingFile(candidates)
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal testdata: %v", err)
	}

	filterTypes := extractObjectMap(raw, "tipos-filtros")
	filters := extractObjectMap(raw, "filtros")

	// precompute nameToShort / nameToMode and keysByType
	nameToShort := make(map[string]string)
	nameToMode := make(map[string]string)
	keysByType := make(map[string]map[string]bool)

	// gather all names (filters + tipos)
	nameSet := make(map[string]bool)
	for name := range filters {
		nameSet[name] = true
	}
	for t := range filterTypes {
		nameSet[t] = true
	}
	var allNames []string
	for name := range nameSet {
		allNames = append(allNames, name)
	}
	allNames = ordenar(allNames)

	for _, name := range allNames {
		if _, ok := nameToShort[name]; !ok {
			typeKey := determineTypeKey(filters, filterTypes, name)
			if _, ok := keysByType[typeKey]; !ok {
				keysByType[typeKey] = make(map[string]bool)
			}
			short, genType, err := TryExtractSignificant(name, keysByType[typeKey])
			if err != nil {
				t.Fatalf("precompute failed for %s: %v", name, err)
			}
			if short == "" {
				t.Fatalf("precompute produced empty short for %s", name)
			}
			nameToShort[name] = short
			nameToMode[name] = genType
			keysByType[typeKey][short] = true
		}
	}

	// replicate generation per tipo and per filter, populating nomeCurto on filters
	modeCounts := make(map[string]int)
	canonicalCounts := make(map[string]int)

	for tipoNome := range filterTypes {
		filterNames := []string{}
		for name, filter := range filters {
			if fmt.Sprintf("%v", filter["tipo"]) == tipoNome {
				filterNames = append(filterNames, name)
			}
		}
		filterNames = ordenar(filterNames)
		if _, ok := keysByType[tipoNome]; !ok {
			keysByType[tipoNome] = make(map[string]bool)
		}
		usedKeys := keysByType[tipoNome]

		// tipo mapping
		if tShort, ok := nameToShort[tipoNome]; ok {
			if tf, tok := filterTypes[tipoNome]; tok {
				tf["nomeCurto"] = tShort
			}
			usedKeys[tShort] = true
			mode := nameToMode[tipoNome]
			if mode == "" {
				mode = "(unknown)"
			}
			modeCounts[mode]++
			canon := mode
			if idx := indexOfSpace(mode); idx != -1 {
				canon = mode[:idx]
			}
			canonicalCounts[canon]++
		}

		for _, name := range filterNames {
			// build hierarchy and per-level generation
			hierarchy := getHierarchy(filters, name)
			// per-level used
			usedPerLevel := make([]map[string]bool, len(hierarchy))
			for i := range usedPerLevel {
				usedPerLevel[i] = make(map[string]bool)
			}
			shortParts := []string{}
			for i, h := range hierarchy {
				if prev, ok := nameToShort[h]; ok {
					shortParts = append(shortParts, prev)
					usedPerLevel[i][prev] = true
				} else {
					sp, err := ensureShortForPart(h, i, usedPerLevel[i], nameToShort, nameToMode)
					if err != nil {
						t.Fatalf("failed to ensure short for %s: %v", h, err)
					}
					shortParts = append(shortParts, sp)
				}
			}
			shortName := joinWithDash(shortParts)
			// resolve global collision within tipo
			if !usedKeys[shortName] {
				usedKeys[shortName] = true
				last := hierarchy[len(hierarchy)-1]
				mode := nameToMode[last]
				if mode == "" {
					mode = "(unknown)"
				}
				modeCounts[mode]++
				canon := mode
				if idx := indexOfSpace(mode); idx != -1 {
					canon = mode[:idx]
				}
				canonicalCounts[canon]++
				filters[name]["nomeCurto"] = shortName
				continue
			}
			// collision: delegate collision mitigation to the acronym generator
			// by asking it to produce a unique short for the full filter name
			newShort, newMode, err := TryExtractSignificant(name, usedKeys)
			if err != nil {
				t.Fatalf("unable to resolve collision for %s: %v", name, err)
			}
			if newShort == "" {
				t.Fatalf("generator produced empty short while resolving collision for %s", name)
			}
			filters[name]["nomeCurto"] = newShort
			modeCounts[newMode]++
			canon := newMode
			if idx := indexOfSpace(newMode); idx != -1 {
				canon = newMode[:idx]
			}
			canonicalCounts[canon]++
		}
	}

	// validations: no empty nomeCurto and per-type uniqueness
	for tipoNome := range filterTypes {
		seen := make(map[string]bool)
		for name, filter := range filters {
			if fmt.Sprintf("%v", filter["tipo"]) != tipoNome {
				continue
			}
			v, ok := filter["nomeCurto"]
			if !ok {
				t.Fatalf("missing nomeCurto for %s", name)
			}
			s := fmt.Sprintf("%v", v)
			if s == "" {
				t.Fatalf("empty nomeCurto for %s", name)
			}
			if seen[s] {
				t.Fatalf("duplicate short within tipo %s: %s", tipoNome, s)
			}
			seen[s] = true
		}
	}

	t.Logf("modeCounts: %v", modeCounts)
	t.Logf("canonicalCounts: %v", canonicalCounts)
}

func joinWithDash(parts []string) string {
	res := ""
	for i, p := range parts {
		if i > 0 {
			res += "-"
		}
		res += p
	}
	return res
}

func indexOfSpace(s string) int {
	for i, r := range s {
		if r == ' ' {
			return i
		}
	}
	return -1
}
