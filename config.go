package main

import (
	"fmt"

	"github.com/Tech-Arch1tect/config"
)

type Config struct {
	ContentDir            string `env:"CONTENT_DIR" validate:"required"`
	OutputDir             string `env:"OUTPUT_DIR" validate:"required"`
	PostsPerPage          int    `env:"POSTS_PER_PAGE"`
	PreviewsPerPage       int    `env:"PREVIEWS_PER_PAGE"`
	DateFormat            string `env:"DATE_FORMAT"`
	CorsAllowOrigin       string `env:"CORS_ALLOW_ORIGIN"`
	CorsAllowMethods      string `env:"CORS_ALLOW_METHODS"`
	CorsAllowHeaders      string `env:"CORS_ALLOW_HEADERS"`
	CorsMaxAge            int    `env:"CORS_MAX_AGE"`
	AverageWordsPerMinute int    `env:"AVERAGE_WORDS_PER_MINUTE"`
}

func NewConfig() *Config {
	return &Config{
		ContentDir:            "./content",
		OutputDir:             "./output",
		PostsPerPage:          10,
		PreviewsPerPage:       10,
		DateFormat:            "2006-01-02",
		CorsAllowOrigin:       "*",
		CorsAllowMethods:      "GET, OPTIONS",
		CorsAllowHeaders:      "Origin, X-Requested-With, Content-Type, Accept",
		CorsMaxAge:            86400,
		AverageWordsPerMinute: 200,
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
	if c.CorsMaxAge < 0 {
		return fmt.Errorf("CORS max age must be non-negative, got %d", c.CorsMaxAge)
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
	if c.CorsAllowOrigin == "" {
		c.CorsAllowOrigin = "*"
	}
	if c.CorsAllowMethods == "" {
		c.CorsAllowMethods = "GET, OPTIONS"
	}
	if c.CorsAllowHeaders == "" {
		c.CorsAllowHeaders = "Origin, X-Requested-With, Content-Type, Accept"
	}
	if c.CorsMaxAge == 0 {
		c.CorsMaxAge = 86400
	}
	if c.AverageWordsPerMinute == 0 {
		c.AverageWordsPerMinute = 200
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{ContentDir: %q, OutputDir: %q, PostsPerPage: %d, PreviewsPerPage: %d, DateFormat: %q, CorsAllowOrigin: %q, CorsAllowMethods: %q, CorsAllowHeaders: %q, CorsMaxAge: %d}",
		c.ContentDir, c.OutputDir, c.PostsPerPage, c.PreviewsPerPage, c.DateFormat, c.CorsAllowOrigin, c.CorsAllowMethods, c.CorsAllowHeaders, c.CorsMaxAge)
}
