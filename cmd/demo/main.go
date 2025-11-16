package main

import (
	"fmt"
	"os"

	"github.com/brunojet/go-text-utils/pkg/acronym"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: demo <name> [name...]")
		os.Exit(2)
	}

	// reuse a single usedKeys map so the demo shows how collisions are
	// avoided when generating multiple shorts in the same run.
	used := make(map[string]bool)
	for _, name := range os.Args[1:] {
		short, mode, err := acronym.TryExtractSignificant(name, used)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating short for %q: %v\n", name, err)
			continue
		}
		fmt.Printf("%s -> %s (%s)\n", name, short, mode)
	}
}
