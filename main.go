package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[Mantle] ", log.LstdFlags)

	var generateOpenAPIOnly bool
	flag.BoolVar(&generateOpenAPIOnly, "openapi-only", false, "Generate only the OpenAPI specification and exit")
	flag.BoolVar(&generateOpenAPIOnly, "generate-openapi-only", false, "Generate only the OpenAPI specification and exit (alias)")
	flag.Parse()

	cfg := NewConfig()
	if err := cfg.Load(); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	logger.Printf("Loaded configuration: %s", cfg)

	if generateOpenAPIOnly {
		logger.Println("Generating OpenAPI specification only...")

		swaggerGenerator := NewSwaggerGenerator(cfg)
		if err := swaggerGenerator.Generate(); err != nil {
			logger.Fatalf("Failed to generate OpenAPI specification: %v", err)
		}

		logger.Println("OpenAPI specification generated successfully")
		return
	}

	loader := NewPostLoader(cfg.ContentDir, cfg.AverageWordsPerMinute)
	posts, err := loader.LoadAll()
	if err != nil {
		logger.Fatalf("Failed to load posts: %v", err)
	}

	if len(posts) == 0 {
		logger.Fatalf("No posts found in %s", cfg.ContentDir)
	}
	logger.Printf("Loaded %d post(s)", len(posts))

	processor := NewPostProcessor()
	processedPosts := processor.Process(posts)

	outputProcessor := NewOutputProcessor(cfg)
	if err := outputProcessor.Process(processedPosts); err != nil {
		logger.Fatalf("Failed to process output: %v", err)
	}

	webServerGenerator := NewWebServerGenerator(cfg)
	if err := webServerGenerator.Generate(); err != nil {
		logger.Fatalf("Failed to generate webserver files: %v", err)
	}

	if cfg.GenerateSwagger {
		swaggerGenerator := NewSwaggerGenerator(cfg)
		if err := swaggerGenerator.Generate(); err != nil {
			logger.Fatalf("Failed to generate OpenAPI specification: %v", err)
		}
	}

	logger.Println("Mantle completed successfully")
}
