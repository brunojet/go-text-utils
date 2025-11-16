package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	acronym "github.com/brunojet/go-text-utils/pkg/acronym"
)

const unknownMode = "(unknown)"

// modeLabel returns the mode string for a given name using nameToMode map,
// or returns the default unknownMode when not present.
func modeLabel(nameToMode map[string]string, name string) string {
	if m, ok := nameToMode[name]; ok && m != "" {
		return m
	}
	return unknownMode
}

// recordModeCounts increments both modeCounts and canonicalCounts using a
// human-readable modeLabel. Canonical is the first token before a space.
func recordModeCounts(modeCounts map[string]int, canonicalCounts map[string]int, modeLabelStr string) {
	modeCounts[modeLabelStr]++
	canonical := modeLabelStr
	if idx := strings.Index(modeLabelStr, " "); idx != -1 {
		canonical = modeLabelStr[:idx]
	}
	canonicalCounts[canonical]++
}

// ordenar returns a sorted copy of the provided slice.
func ordenar(lista []string) []string {
	res := make([]string, len(lista))
	copy(res, lista)
	sort.Strings(res)
	return res
}

func generateShortName(name string, usedKeys map[string]bool) (string, string, error) {
	// delegate to the acronym package (it does normalization internally)
	return acronym.TryExtractSignificant(name, usedKeys)
}

// loadNomeParaCurto carrega um mapeamento nome->curto de um arquivo JSON.
// Se o arquivo não existir, retorna um mapa vazio e nil error.
type auxEntry struct {
	Short string `json:"curto"`
	Mode  string `json:"modo"`
}

// Monta nome curto final juntando nomes curtos de cada camada
func buildShortName(hierarchy []string, usedKeys map[string]bool, nameToShort map[string]string, nameToMode map[string]string) (string, string) {
	shortParts := []string{}
	usedPerLevel := make([]map[string]bool, len(hierarchy))
	for i := range usedPerLevel {
		usedPerLevel[i] = make(map[string]bool)
	}
	for i, h := range hierarchy {
		// Reuse if mapping for this name already exists (global consistency)
		var short string
		if prev, ok := nameToShort[h]; ok {
			short = prev
			if short != "" {
				usedPerLevel[i][short] = true
			}
		} else {
			short, genType, err := generateShortName(h, usedPerLevel[i])
			if err != nil {
				panic(fmt.Errorf("failed to generate short name for '%s' at level %d: %w", h, i, err))
			}
			if short == "" {
				panic(fmt.Errorf("generator returned empty short for '%s' at level %d", h, i))
			}
			nameToShort[h] = short
			nameToMode[h] = genType
		}
		shortParts = append(shortParts, short)
	}
	shortName := strings.Join(shortParts, "-")

	// Resolve global collision: try the canonical shortName first, otherwise
	// probe deterministic numeric suffixes until we find a free one.
	if !usedKeys[shortName] {
		usedKeys[shortName] = true
		last := hierarchy[len(hierarchy)-1]
		lastMode := ""
		if m, ok := nameToMode[last]; ok {
			lastMode = m
		}
		return shortName, lastMode
	}

	// Collision detected. Generate deterministic suffixed candidates.
	last := hierarchy[len(hierarchy)-1]
	lastMode := ""
	if m, ok := nameToMode[last]; ok {
		lastMode = m
	}
	// Try numeric suffixes up to a reasonable limit.
	for i := 1; i <= 999; i++ {
		candidate := fmt.Sprintf("%s-%03d", shortName, i)
		if !usedKeys[candidate] {
			usedKeys[candidate] = true
			return candidate, lastMode
		}
	}

	// Extremely unlikely: no free candidate found in probe range.
	panic(fmt.Errorf("unable to resolve collision for shortName '%s' (hierarchy %v)", shortName, hierarchy))
}

// readmeContent lê o conteúdo do README.md e retorna como string
func readmeContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Função global: busca hierarquia de pais
func getHierarchy(filters map[string]map[string]interface{}, name string) []string {
	f, ok := filters[name]
	if !ok {
		return []string{name}
	}
	typ := fmt.Sprintf("%v", f["tipo"])   // keep domain field name 'tipo'
	parent := fmt.Sprintf("%v", f["pai"]) // keep domain field name 'pai'
	if parent != "<nil>" && parent != "" && parent != "null" {
		return append(getHierarchy(filters, parent), name)
	}
	return []string{typ, name}
}

// determineTypeKey retorna a chave usada para agrupar chaves curtas por tipo.
// Se o nome for um tipo (ex: "Ramo"), a própria string do tipo é retornada.
// Se o nome for um filtro (tem campo "tipo"), retornamos esse tipo.
// Caso contrário, retornamos um marcador global (_GLOBAL) para nomes desconhecidos.
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

// readFirstExistingFile lê o primeiro caminho existente da lista e retorna o conteúdo.
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

// extractObjectMap extrai um mapa[string]map[string]interface{} do mapa genérico
// retornado pelo json.Unmarshal quando o valor em `key` é um objeto de objetos.
// Ex: transforma filtrosRaw["filtros"] em map[nome] = map[string]interface{}.
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

// printStatistics prints mode and canonical aggregated statistics.
func printStatistics(modeCounts map[string]int, canonicalCounts map[string]int) {
	fmt.Println("\n--- Estatísticas de modos ---")
	total := 0
	for _, c := range modeCounts {
		total += c
	}
	if total == 0 {
		fmt.Println("Nenhuma entrada processada.")
		return
	}
	var modosList []string
	for m := range modeCounts {
		modosList = append(modosList, m)
	}
	sort.Strings(modosList)
	for _, m := range modosList {
		c := modeCounts[m]
		pct := (float64(c) / float64(total)) * 100.0
		fmt.Printf("%s: %d (%.1f%%)\n", m, c, pct)
	}
	fmt.Printf("Total: %d\n", total)

	fmt.Println("\n--- Estatísticas canônicas (agregado) ---")
	var canList []string
	for k := range canonicalCounts {
		canList = append(canList, k)
	}
	sort.Strings(canList)
	canTotal := 0
	for _, c := range canonicalCounts {
		canTotal += c
	}
	for _, k := range canList {
		c := canonicalCounts[k]
		pct := (float64(c) / float64(canTotal)) * 100.0
		fmt.Printf("%s: %d (%.1f%%)\n", k, c, pct)
	}

}

// generateShortNames gera os nomes curtos por tipo e popula `filters` com
// o campo `nomeCurto`. Não faz nenhuma impressão; retorna as contagens de modos
// e canônicos para uso posterior.
func generateShortNames(filterTypes map[string]map[string]interface{}, filters map[string]map[string]interface{}, nameToShort map[string]string, nameToMode map[string]string, keysByType map[string]map[string]bool) (map[string]int, map[string]int) {
	// split responsibilities: iterate tipos and delegate per-tipo processing
	modeCounts := make(map[string]int)
	canonicalCounts := make(map[string]int)
	for tipoNome := range filterTypes {
		// collect names for this tipo
		filterNames := getFilterNamesOfType(filters, tipoNome)
		filterNames = ordenar(filterNames)
		generateShortNamesForType(tipoNome, filterNames, filterTypes, filters, nameToShort, nameToMode, keysByType, modeCounts, canonicalCounts)
	}
	return modeCounts, canonicalCounts
}

// getFilterNamesOfType returns the names of filters that belong to the given tipo.
func getFilterNamesOfType(filters map[string]map[string]interface{}, tipo string) []string {
	var res []string
	for name, filter := range filters {
		if fmt.Sprintf("%v", filter["tipo"]) == tipo {
			res = append(res, name)
		}
	}
	return res
}

// precomputeNameToShort builds initial nameToShort and nameToMode maps and
// per-type used keys map. It returns (nameToShort, nameToMode, keysByType, error).
func precomputeNameToShort(filters map[string]map[string]interface{}, filterTypes map[string]map[string]interface{}) (map[string]string, map[string]string, map[string]map[string]bool, error) {
	nameToShort := make(map[string]string)
	nameToMode := make(map[string]string)
	keysByType := make(map[string]map[string]bool)

	// include both filter names and tipo names
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
			short, genType, err := generateShortName(name, keysByType[typeKey])
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to generate short name during precompute for '%s': %w", name, err)
			}
			if short == "" {
				return nil, nil, nil, fmt.Errorf("generator returned empty short for '%s' during precompute", name)
			}
			nameToShort[name] = short
			nameToMode[name] = genType
			keysByType[typeKey][short] = true
		}
	}

	return nameToShort, nameToMode, keysByType, nil
}

// generateShortNamesForType handles generation for a single filter type.
func generateShortNamesForType(tipoNome string, filterNames []string, filterTypes map[string]map[string]interface{}, filters map[string]map[string]interface{}, nameToShort map[string]string, nameToMode map[string]string, keysByType map[string]map[string]bool, modeCounts map[string]int, canonicalCounts map[string]int) {
	// ensure per-type usedKeys map exists
	if _, ok := keysByType[tipoNome]; !ok {
		keysByType[tipoNome] = make(map[string]bool)
	}
	usedKeys := keysByType[tipoNome]

	// If the 'tipo' itself has a short mapping (we precomputed types too), record it
	if tShort, ok := nameToShort[tipoNome]; ok {
		if tf, tok := filterTypes[tipoNome]; tok {
			tf["nomeCurto"] = tShort
		}
		usedKeys[tShort] = true
		modeLabelStr := modeLabel(nameToMode, tipoNome)
		recordModeCounts(modeCounts, canonicalCounts, modeLabelStr)
	}

	for _, name := range filterNames {
		processFilterName(name, filters, nameToShort, nameToMode, usedKeys, modeCounts, canonicalCounts)
	}
}

// processFilterName generates the short name for a single filter and updates
// the provided counts and maps.
func processFilterName(name string, filters map[string]map[string]interface{}, nameToShort map[string]string, nameToMode map[string]string, usedKeys map[string]bool, modeCounts map[string]int, canonicalCounts map[string]int) {
	filter := filters[name]
	hierarchy := getHierarchy(filters, name)
	shortName, mode := buildShortName(hierarchy, usedKeys, nameToShort, nameToMode)
	usedKeys[shortName] = true
	filter["nomeCurto"] = shortName
	modeLabelStr := mode
	if modeLabelStr == "" {
		modeLabelStr = unknownMode
	}
	recordModeCounts(modeCounts, canonicalCounts, modeLabelStr)
}

// logShortNames imprime o log de nomes curtos baseado nos dados já gerados
// em `filters`, `nameToShort` e `nameToMode`.
func logShortNames(filterTypes map[string]map[string]interface{}, filters map[string]map[string]interface{}, nameToShort map[string]string, nameToMode map[string]string, keysByType map[string]map[string]bool) {
	fmt.Println("\n--- LOG DE nomesCurtos ---")
	for tipoNome := range filterTypes {
		fmt.Printf("\nTipo-filtro: %s\n", tipoNome)
		filterNames := getFilterNamesOfType(filters, tipoNome)
		filterNames = ordenar(filterNames)
		if _, ok := keysByType[tipoNome]; !ok {
			keysByType[tipoNome] = make(map[string]bool)
		}
		usedKeys := keysByType[tipoNome]

		if tShort, ok := nameToShort[tipoNome]; ok {
			modeLabelStr := "(unknown)"
			if m, mok := nameToMode[tipoNome]; mok && m != "" {
				modeLabelStr = m
			}
			fmt.Printf("%s --> %s (modo=%s)\n", tShort, tipoNome, modeLabelStr)
			if tf, tok := filterTypes[tipoNome]; tok {
				tf["nomeCurto"] = tShort
			}
			usedKeys[tShort] = true
		}
		for _, name := range filterNames {
			filter := filters[name]
			if v, ok := filter["nomeCurto"]; ok {
				longHierarchy := strings.Join(getHierarchy(filters, name), "-")
				modeLabelStr := "(unknown)"
				// infer mode from nameToMode if available
				if m, mok := nameToMode[name]; mok && m != "" {
					modeLabelStr = m
				}
				fmt.Printf("%s --> %s (modo=%s)\n", v, longHierarchy, modeLabelStr)
			} else {
				// nomeCurto missing: log a warning instead of recomputing here.
				longHierarchy := strings.Join(getHierarchy(filters, name), "-")
				fmt.Printf("[WARN] nomeCurto ausente para %s (hierarchy=%s)\n", name, longHierarchy)
			}
		}
	}
}

func main() {
	// Lê e exibe conteúdo do README.md
	content, err := readmeContent("../README.md")
	if err == nil {
		fmt.Println("Conteúdo do README.md:")
		fmt.Println(content)
	}

	// Lê testdata/filtros.json (try a few likely relative paths so the program
	// works whether executed from the repo root or from src/)
	candidates := []string{"testdata/filtros.json", "../testdata/filtros.json", "filtros.json"}
	filtrosData, readErr := readFirstExistingFile(candidates)
	if readErr != nil {
		fmt.Println("Erro ao ler filtros.json:", readErr)
		return
	}

	// Decodifica para estrutura genérica
	var filtrosRaw map[string]interface{}
	err = json.Unmarshal(filtrosData, &filtrosRaw)
	if err != nil {
		fmt.Println("Erro ao decodificar filtros.json:", err)
		return
	}

	filterTypes := extractObjectMap(filtrosRaw, "tipos-filtros")
	filters := extractObjectMap(filtrosRaw, "filtros")

	// nota: usamos a função top-level `ordenar` definida acima

	// ...existing code...

	// Pré-calcula (sem cache em disco): inicializa mapas em memória.
	// Removemos o uso de `data/nomeParaCurto.json` para evitar que um cache
	// em disco influencie os resultados gerados nesta execução.
	nameToShort, nameToMode, keysByType, err := precomputeNameToShort(filters, filterTypes)
	if err != nil {
		panic(err)
	}

	// Nota: não salvamos mapeamento auxiliar em disco (cache desabilitado).

	modeCounts, canonicalCounts := generateShortNames(filterTypes, filters, nameToShort, nameToMode, keysByType)
	logShortNames(filterTypes, filters, nameToShort, nameToMode, keysByType)

	// imprime estatísticas resumidas por modo e canônicas (delegado)
	printStatistics(modeCounts, canonicalCounts)
}

// modoLabel removido: agora usamos a string do modo diretamente
