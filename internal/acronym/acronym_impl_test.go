package acronym

import (
	"testing"
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
