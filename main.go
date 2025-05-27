package main

import (
	"fmt"
	"log"
)

func main() {
	cfg := NewConfig()
	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded configuration: %s\n", cfg)

	loader := NewPostLoader(cfg.ContentDir)

	posts, err := loader.LoadAll()
	if err != nil {
		log.Fatalf("Failed to load posts: %v", err)
	}

	if len(posts) == 0 {
		log.Fatalf("No posts found in %s", cfg.ContentDir)
	}

	fmt.Printf("Loaded %d post(s)\n", len(posts))

	processedPosts := processPosts(posts)

	processOutput(cfg, processedPosts)
}
