package main

import (
	"log"

	"github.com/Tech-Arch1tect/config"
)

type AppConfig struct {
	ContentDir string `env:"CONTENT_DIR" validate:"required"`
}

func (c *AppConfig) SetDefaults() {
	c.ContentDir = "./content"
}

func loadConfig() (AppConfig, error) {
	var cfg AppConfig
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg, nil
}
