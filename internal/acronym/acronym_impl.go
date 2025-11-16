package acronym

import (
	"crypto/md5"
	"fmt"
	"math/big"
	"strings"

	enc "github.com/brunojet/go-text-utils/internal/encoding"
	textutil "github.com/brunojet/go-text-utils/internal/textutil"
)

// IMPLEMENTAÇÃO interna do gerador de acrônimos.

type GenerationType string

const (
	Raw       GenerationType = "raw"
	Consonant GenerationType = "consoante"
	Vowel     GenerationType = "vogal"
	Hash36    GenerationType = "hash36"
)

// text normalization helpers moved to internal/textutil

func isVowel(r rune) bool {
	vowels := "AEIOU"
	return strings.ContainsRune(vowels, r)
}

func collectMiddleRunes(runes []rune, pred func(rune) bool) []rune {
	res := []rune{}
	for i := 1; i < len(runes)-1; i++ {
		if pred(runes[i]) {
			res = append(res, runes[i])
		}
	}
	return res
}

func tryBuildAcronym(s string, offset int, pred func(rune) bool) (string, bool) {
	runes := []rune(s)
	if len(runes) == 0 {
		return "", false
	}
	first := runes[0]
	last := runes[len(runes)-1]

	validRunes := collectMiddleRunes(runes, pred)
	if offset < len(validRunes) {
		return string([]rune{first, validRunes[offset], last}), true
	}
	return "", false
}

func tryExtractSignificantOffset(s string, offset int, mode GenerationType) (string, bool) {
	isCons := func(r rune) bool { return !isVowel(r) }
	isVog := func(r rune) bool { return isVowel(r) }

	switch mode {
	case Consonant:
		return tryBuildAcronym(s, offset, isCons)
	case Vowel:
		return tryBuildAcronym(s, offset, isVog)
	default:
		return "", false
	}
}

// base36 encoding helper moved to internal/encoding

func fallbackHash36(clear string, usedKeys map[string]bool) (string, error) {
	sum := md5.Sum([]byte(clear))
	n := new(big.Int).SetBytes(sum[:])
	base := big.NewInt(36)
	power := big.NewInt(1)
	power.Exp(base, big.NewInt(3), nil) // 36^3

	start := new(big.Int).Mod(n, power)
	cur := new(big.Int).Set(start)
	one := big.NewInt(1)

	for i := big.NewInt(0); i.Cmp(power) == -1; i.Add(i, one) {
		candidate := enc.EncodeBase36Fixed(cur, 3)
		if !usedKeys[candidate] {
			usedKeys[candidate] = true
			return candidate, nil
		}
		cur.Add(cur, one)
		if cur.Cmp(power) >= 0 {
			cur.Sub(cur, power)
		}
	}
	return "", fmt.Errorf("não foi possível gerar nome curto (hash36) para '%s'", clear)
}
func TryExtractSignificant(s string, usedKeys map[string]bool) (string, string, error) {
	clear := textutil.RemoveAccents(textutil.RemoveConnectors(strings.ToUpper(s)))
	generationTypes := []GenerationType{Raw, Consonant, Vowel, Hash36}
	lenRune := len([]rune(clear))

	if lenRune == 0 {
		return "", "", fmt.Errorf("não foi possível extrair significativas de '%s'", clear)
	}

	for _, genType := range generationTypes {
		switch genType {
		case Raw:
			if lenRune <= 3 {
				if !usedKeys[clear] {
					usedKeys[clear] = true
					return clear, string(Raw), nil
				}
			}
		case Consonant, Vowel:
			for offset := 0; offset < lenRune-1; offset++ {
				if shortName, ok := tryExtractSignificantOffset(clear, offset, genType); ok {
					if !usedKeys[shortName] {
						usedKeys[shortName] = true
						return shortName, fmt.Sprintf("%s (%d)", string(genType), offset), nil
					}
				}
			}
		case Hash36:
			if candidate, err := fallbackHash36(clear, usedKeys); err == nil {
				return candidate, string(Hash36), nil
			}
		default:
		}
	}

	return "", "", fmt.Errorf("não foi possível extrair significativas de '%s'", clear)
}
