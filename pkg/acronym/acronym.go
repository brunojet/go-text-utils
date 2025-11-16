package acronym

import (
	internal "github.com/brunojet/go-text-utils/internal/acronym"
)

// Este arquivo expõe a API mínima do pacote `pkg/acronym` e delega
// para a implementação em internal/acronym.

// Wrapper that delegates to internal/acronym.
// TryExtractSignificant is the public wrapper that delegates to
// the implementation in internal/acronym.
func TryExtractSignificant(s string, usedKeys map[string]bool) (string, string, error) {
	return internal.TryExtractSignificant(s, usedKeys)
}
