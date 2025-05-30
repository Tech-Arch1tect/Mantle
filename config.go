package main

import (
	"fmt"

	"github.com/Tech-Arch1tect/config"
)

type Config struct {
	ContentDir      string `env:"CONTENT_DIR" validate:"required"`
	OutputDir       string `env:"OUTPUT_DIR" validate:"required"`
	PostsPerPage    int    `env:"POSTS_PER_PAGE"`
	PreviewsPerPage int    `env:"PREVIEWS_PER_PAGE"`
	DateFormat      string `env:"DATE_FORMAT"`
}

func NewConfig() *Config {
	return &Config{
		ContentDir:      "./content",
		OutputDir:       "./output",
		PostsPerPage:    10,
		PreviewsPerPage: 10,
		DateFormat:      "2006-01-02",
	}
}

func (c *Config) Load() error {
	c.SetDefaults()

	if err := config.Load(c); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if c.PostsPerPage < 1 {
		return fmt.Errorf("posts per page must be at least 1, got %d", c.PostsPerPage)
	}
	if c.PreviewsPerPage < 1 {
		return fmt.Errorf("previews per page must be at least 1, got %d", c.PreviewsPerPage)
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
	if c.PostsPerPage == 0 {
		c.PostsPerPage = 10
	}
	if c.PreviewsPerPage == 0 {
		c.PreviewsPerPage = 10
	}
	if c.DateFormat == "" {
		c.DateFormat = "2006-01-02"
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{ContentDir: %q, OutputDir: %q, PostsPerPage: %d, PreviewsPerPage: %d, DateFormat: %q}",
		c.ContentDir, c.OutputDir, c.PostsPerPage, c.PreviewsPerPage, c.DateFormat)
}
