package main

import (
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/calforge/internal/cmd"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
