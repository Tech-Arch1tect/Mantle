package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[Mantle] ", log.LstdFlags)

	cfg := NewConfig()
	if err := cfg.Load(); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	logger.Printf("Loaded configuration: %s", cfg)

	loader := NewPostLoader(cfg.ContentDir)
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

	logger.Println("Mantle completed successfully")
}
