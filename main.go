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

	posts, err := loadPosts(cfg.ContentDir)
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
