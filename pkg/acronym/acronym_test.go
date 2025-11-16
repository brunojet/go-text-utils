package acronym

import (
	"strings"
	"testing"
)

func TestTryExtractSignificant_Basic(t *testing.T) {
	used := make(map[string]bool)
	short, mode, err := TryExtractSignificant("Alimentação", used)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len([]rune(short)) != 3 {
		t.Fatalf("expected 3-rune short, got %q", short)
	}
	if strings.ToUpper(short) != short {
		t.Fatalf("expected uppercase short, got %q", short)
	}
	if !used[short] {
		t.Fatalf("usedKeys must contain the returned short")
	}
	if mode == "" {
		t.Fatalf("mode should be non-empty")
	}
}

func TestTryExtractSignificant_Raw(t *testing.T) {
	used := make(map[string]bool)
	short, mode, err := TryExtractSignificant("EAD", used)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if short != "EAD" {
		t.Fatalf("expected raw 'EAD' to be preserved, got %q (mode=%s)", short, mode)
	}
	if mode != "raw" {
		t.Fatalf("expected mode 'raw' for EAD, got %q", mode)
	}
}

func TestTryExtractSignificant_Collision(t *testing.T) {
	used := make(map[string]bool)
	s1, _, err := TryExtractSignificant("Alimentação", used)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// call again with same used map; should return a different short (or at least not the same)
	s2, _, err := TryExtractSignificant("Alimentação", used)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if s1 == s2 {
		t.Fatalf("expected a different short on second call when first is reserved, got same %q", s1)
	}
}
