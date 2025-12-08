package main

import (
	"log"

	"goservice/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("service failed: %v", err)
	}
}
