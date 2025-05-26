package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

func processOutput(cfg AppConfig, processedPosts ProcessedPosts) {
	log.Println("Processing output...")

	err := makeDirectories(cfg)
	if err != nil {
		log.Fatalf("Failed to make directories: %v", err)
	}

	err = saveJSON(cfg.OutputDir+"/public_html/api/posts/all.json", processedPosts.Posts)
	if err != nil {
		log.Fatalf("Failed to save posts: %v", err)
	}

	for _, post := range processedPosts.Posts {
		err = saveJSON(cfg.OutputDir+"/public_html/api/posts/post_"+strconv.Itoa(post.Index)+".json", post)
		if err != nil {
			log.Fatalf("Failed to save post: %v", err)
		}
	}

	saveJSON(cfg.OutputDir+"/public_html/api/tags/all.json", processedPosts.Tags)
	for tag, posts := range processedPosts.Tags {
		err = saveJSON(cfg.OutputDir+"/public_html/api/tags/"+tag+".json", posts)
		if err != nil {
			log.Fatalf("Failed to save tag: %v", err)
		}
	}

	log.Println("Output processed")
}

func saveJSON(path string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

func makeDirectories(cfg AppConfig) error {
	directories := []string{
		cfg.OutputDir + "/public_html/api/tags",
		cfg.OutputDir + "/public_html/api/posts",
	}
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
