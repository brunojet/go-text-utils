package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	acronym "github.com/brunojet/go-text-utils/pkg/acronym"
)

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

// loadNomeParaCurto carrega o mapa nome->(curto,modo). Suporta o formato
// antigo (map[string]string) para compatibilidade.
func loadNameToShort(path string) (map[string]string, map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), make(map[string]string), nil
		}
		return nil, nil, err
	}
	// Primeiro tenta o novo formato: map[name] = {curto, modo}
	var obj map[string]auxEntry
	if err := json.Unmarshal(data, &obj); err == nil {
		shortMap := make(map[string]string)
		modeMap := make(map[string]string)
		for k, v := range obj {
			shortMap[k] = v.Short
			if v.Mode != "" {
				modeMap[k] = v.Mode
			}
		}
		return shortMap, modeMap, nil
	}
	// Fallback: formato antigo simples map[string]string
	var old map[string]string
	if err := json.Unmarshal(data, &old); err == nil {
		modeMap := make(map[string]string)
		for k := range old {
			modeMap[k] = ""
		}
		return old, modeMap, nil
	}
	return nil, nil, fmt.Errorf("formato desconhecido em %s", path)
}

// saveNomeParaCurto persiste o mapa em JSON (cria diretório se necessário).
// saveNomeParaCurto salva o mapa nome->(curto,modo) em JSON.
func saveNameToShort(path string, m map[string]string, modes map[string]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	obj := make(map[string]auxEntry)
	for k, v := range m {
		mode := ""
		if mm, ok := modes[k]; ok {
			mode = mm
		}
		obj[k] = auxEntry{Short: v, Mode: mode}
	}
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
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
			usedPerLevel[i][short] = true
		} else {
			short, genType, err := generateShortName(h, usedPerLevel[i])
			if err != nil {
				panic(fmt.Errorf("failed to generate short name for '%s' at level %d: %w", h, i, err))
			}
			nameToShort[h] = short
			nameToMode[h] = genType
		}
		shortParts = append(shortParts, short)
	}
	shortName := strings.Join(shortParts, "-")

	// Resolve global collision
	if !usedKeys[shortName] {
		usedKeys[shortName] = true
		last := hierarchy[len(hierarchy)-1]
		lastMode := ""
		if m, ok := nameToMode[last]; ok {
			lastMode = m
		}
		return shortName, lastMode
	}

	return "", ""
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

func main() {
	// Lê e exibe conteúdo do README.md
	content, err := readmeContent("../README.md")
	if err == nil {
		fmt.Println("Conteúdo do README.md:")
		fmt.Println(content)
	}

	// Lê filtros.json
	filtrosData, err := os.ReadFile("filtros.json")
	if err != nil {
		fmt.Println("Erro ao ler filtros.json:", err)
		return
	}

	// Decodifica para estrutura genérica
	var filtrosRaw map[string]interface{}
	err = json.Unmarshal(filtrosData, &filtrosRaw)
	if err != nil {
		fmt.Println("Erro ao decodificar filtros.json:", err)
		return
	}

	filterTypes := make(map[string]map[string]interface{})
	if tf, ok := filtrosRaw["tipos-filtros"].(map[string]interface{}); ok {
		for k, v := range tf {
			if obj, ok := v.(map[string]interface{}); ok {
				filterTypes[k] = obj
			}
		}
	}

	filters := make(map[string]map[string]interface{})
	if fs, ok := filtrosRaw["filtros"].(map[string]interface{}); ok {
		for k, v := range fs {
			if obj, ok := v.(map[string]interface{}); ok {
				filters[k] = obj
			}
		}
	}

	ordenar := func(lista []string) []string {
		res := make([]string, len(lista))
		copy(res, lista)
		for i := 0; i < len(res)-1; i++ {
			for j := 0; j < len(res)-i-1; j++ {
				if res[j] > res[j+1] {
					res[j], res[j+1] = res[j+1], res[j]
				}
			}
		}
		return res
	}

	// ...existing code...

	// Pré-calcula (ou carrega) um mapeamento canônico nome -> curto para todos os filtros
	// usando um conjunto global de chaves por nome, para garantir que o mesmo
	// filtro gere sempre o mesmo acrônimo independentemente da ordem.
	auxPath := "../data/nomeParaCurto.json"
	nameToShort, nameToMode, err := loadNameToShort(auxPath)
	if err != nil {
		fmt.Println("Aviso: falha ao carregar mapeamento auxiliar, prosseguindo com mapa vazio:", err)
		nameToShort = make(map[string]string)
		nameToMode = make(map[string]string)
	}
	keysByName := make(map[string]bool)
	// mark keys already used loaded from file
	for name, v := range nameToShort {
		keysByName[v] = true
		if _, ok := nameToMode[name]; ok {
			// keep existing mode
		} else {
			nameToMode[name] = ""
		}
	}

	// incluir tanto os nomes dos filtros quanto os nomes de tipo (ex: "Ramo", "Funcionalidade")
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
		// only generate a new short if there isn't one in the loaded map
		if _, ok := nameToShort[name]; !ok {
			short, genType, err := generateShortName(name, keysByName)
			if err != nil {
				panic(fmt.Errorf("failed to generate short name during precompute for '%s': %w", name, err))
			}
			nameToShort[name] = short
			nameToMode[name] = genType
			keysByName[short] = true
		}
	}

	// salvar (ou atualizar) o auxiliar no disco
	if err := saveNameToShort(auxPath, nameToShort, nameToMode); err != nil {
		fmt.Println("Aviso: falha ao gravar mapeamento auxiliar:", err)
	}

	fmt.Println("\n--- LOG DE nomesCurtos ---")
	// estatísticas de modos usadas durante a geração/impressão
	modeCounts := make(map[string]int)
	// agregação canônica (sem offsets), ex: "consoante", "vogal", "raw", "hash36", "(unknown)"
	canonicalCounts := make(map[string]int)
	for tipoNome := range filterTypes {
		fmt.Printf("\nTipo-filtro: %s\n", tipoNome)
		var filterNames []string
		for name, filter := range filters {
			if fmt.Sprintf("%v", filter["tipo"]) == tipoNome {
				filterNames = append(filterNames, name)
			}
		}
		filterNames = ordenar(filterNames)
		usedKeys := make(map[string]bool)
		for _, name := range filterNames {
			filter := filters[name]
			hierarchy := getHierarchy(filters, name)
			shortName, mode := buildShortName(hierarchy, usedKeys, nameToShort, nameToMode)
			usedKeys[shortName] = true
			filter["nomeCurto"] = shortName
			longHierarchy := strings.Join(hierarchy, "-")
			modeLabelStr := "(unknown)"
			if mode != "" {
				modeLabelStr = mode
			}
			fmt.Printf("%s --> %s (modo=%s)\n", shortName, longHierarchy, modeLabelStr)
			// count mode
			modeCounts[modeLabelStr]++
			// update canonical aggregation (remove offset/description after space)
			canonical := modeLabelStr
			if idx := strings.Index(modeLabelStr, " "); idx != -1 {
				canonical = modeLabelStr[:idx]
			}
			canonicalCounts[canonical]++
		}
	}

	// imprime estatísticas resumidas por modo
	fmt.Println("\n--- Estatísticas de modos ---")
	total := 0
	for _, c := range modeCounts {
		total += c
	}
	if total == 0 {
		fmt.Println("Nenhuma entrada processada.")
		return
	}
	// ordenar chaves para saída estável
	var modosList []string
	for m := range modeCounts {
		modosList = append(modosList, m)
	}
	// simples ordenação lexical
	for i := 0; i < len(modosList)-1; i++ {
		for j := 0; j < len(modosList)-i-1; j++ {
			if modosList[j] > modosList[j+1] {
				modosList[j], modosList[j+1] = modosList[j+1], modosList[j]
			}
		}
	}
	for _, m := range modosList {
		c := modeCounts[m]
		pct := (float64(c) / float64(total)) * 100.0
		fmt.Printf("%s: %d (%.1f%%)\n", m, c, pct)
	}
	fmt.Printf("Total: %d\n", total)

	// Imprime estatísticas canônicas agregadas (por tipo, sem offsets)
	fmt.Println("\n--- Estatísticas canônicas (agregado) ---")
	// ordenar chaves para saída estável
	var canList []string
	for k := range canonicalCounts {
		canList = append(canList, k)
	}
	for i := 0; i < len(canList)-1; i++ {
		for j := 0; j < len(canList)-i-1; j++ {
			if canList[j] > canList[j+1] {
				canList[j], canList[j+1] = canList[j+1], canList[j]
			}
		}
	}
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

// modoLabel removido: agora usamos a string do modo diretamente
