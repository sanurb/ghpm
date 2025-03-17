package main

import (
	"log"

	"github.com/sanurb/ghpm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
