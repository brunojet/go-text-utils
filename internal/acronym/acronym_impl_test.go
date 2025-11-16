package acronym

import (
	"testing"

	"github.com/brunojet/go-text-utils/internal/textutil"
)

func TestTryBuildAcronym_Basic(t *testing.T) {
	// BRUNO -> first B, last O, middle R U N
	isCons := func(r rune) bool { return r != 'U' && r != 'O' && r != 'A' }
	got, ok := tryBuildAcronym("BRUNO", 0, isCons)
	if !ok || got != "BRO" {
		t.Fatalf("tryBuildAcronym BRUNO consonant expected BRO got %q ok=%v", got, ok)
	}

	// Vowel middle for ALUNO -> A L U N O -> vowel U
	isVow := func(r rune) bool { return r == 'U' }
	got2, ok2 := tryBuildAcronym("ALUNO", 0, isVow)
	if !ok2 || got2 != "AUO" {
		t.Fatalf("tryBuildAcronym ALUNO vowel expected AUO got %q ok=%v", got2, ok2)
	}
}

func TestTryExtractSignificantOffset(t *testing.T) {
	s := "ALIMENTACAO"
	if short, ok := tryExtractSignificantOffset(s, 0, Consonant); !ok || len([]rune(short)) != 3 {
		t.Fatalf("expected consonant offset generation for %s to produce 3-rune short, got %q ok=%v", s, short, ok)
	}
	if short, ok := tryExtractSignificantOffset(s, 0, Vowel); !ok || len([]rune(short)) != 3 {
		t.Fatalf("expected vowel offset generation for %s to produce 3-rune short, got %q ok=%v", s, short, ok)
	}
}

func TestFallbackHash36_Unique(t *testing.T) {
	used := make(map[string]bool)
	c1, err1 := fallbackHash36("SOME_CLEAR_TEXT", used)
	if err1 != nil {
		t.Fatalf("fallbackHash36 first err: %v", err1)
	}
	if len([]rune(c1)) != 3 {
		t.Fatalf("fallbackHash36 first returned %q, want 3 chars", c1)
	}
	// call again; since first was recorded in used map, second call should return different
	c2, err2 := fallbackHash36("SOME_CLEAR_TEXT", used)
	if err2 != nil {
		t.Fatalf("fallbackHash36 second err: %v", err2)
	}
	if c1 == c2 {
		t.Fatalf("fallbackHash36 returned same candidate twice: %q", c1)
	}
}

func TestTryExtractSignificant_Empty(t *testing.T) {
	_, _, err := TryExtractSignificant("", make(map[string]bool))
	if err == nil {
		t.Fatalf("expected error for empty input")
	}
}

func TestTryExtractSignificant_Hash36Fallback(t *testing.T) {
	// Prepare a clear text and pre-fill usedKeys with raw and all consonant/vowel candidates
	clear := "ALIMENTACAO"
	used := make(map[string]bool)
	// mark raw
	used[textutil.RemoveAccents(textutil.RemoveConnectors(clear))] = true
	// mark all consonant/vowel generated candidates so the function will fallthrough to hash36
	lenRune := len([]rune(textutil.RemoveAccents(textutil.RemoveConnectors(clear))))
	for offset := 0; offset < lenRune-1; offset++ {
		if s, ok := tryExtractSignificantOffset(textutil.RemoveAccents(textutil.RemoveConnectors(clear)), offset, Consonant); ok {
			used[s] = true
		}
		if s, ok := tryExtractSignificantOffset(textutil.RemoveAccents(textutil.RemoveConnectors(clear)), offset, Vowel); ok {
			used[s] = true
		}
	}

	short, mode, err := TryExtractSignificant(clear, used)
	if err != nil {
		t.Fatalf("unexpected error from TryExtractSignificant: %v", err)
	}
	if mode != string(Hash36) {
		t.Fatalf("expected Hash36 mode when others are exhausted, got mode=%q short=%q", mode, short)
	}
}
