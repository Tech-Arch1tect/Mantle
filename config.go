package main

import (
	"fmt"

	"github.com/Tech-Arch1tect/config"
)

type Config struct {
	ContentDir string `env:"CONTENT_DIR" validate:"required"`
	OutputDir  string `env:"OUTPUT_DIR" validate:"required"`
}

func NewConfig() *Config {
	return &Config{
		ContentDir: "./content",
		OutputDir:  "./output",
	}
}

func (c *Config) Load() error {
	c.SetDefaults()

	if err := config.Load(c); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

func (c *Config) SetDefaults() {
	if c.ContentDir == "" {
		c.ContentDir = "./content"
	}
	if c.OutputDir == "" {
		c.OutputDir = "./output"
	}
}
