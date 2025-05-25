package main

import (
	"fmt"
	"log"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded configuration: %+v\n", cfg)
}
