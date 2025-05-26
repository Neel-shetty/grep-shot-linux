package main

import (
	"log"

	"github.com/neel/grepShot/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
