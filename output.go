package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type OutputProcessor struct {
	config *Config
	logger *log.Logger
}

func NewOutputProcessor(config *Config) *OutputProcessor {
	return &OutputProcessor{
		config: config,
		logger: log.New(os.Stdout, "[OutputProcessor] ", log.LstdFlags),
	}
}

func (op *OutputProcessor) Process(processedPosts ProcessedPosts) error {
	op.logger.Println("Processing output...")

	if err := op.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := op.savePosts(processedPosts.Posts); err != nil {
		return fmt.Errorf("failed to save posts: %w", err)
	}

	if err := op.saveTags(processedPosts.Tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	op.logger.Println("Output processed successfully")
	return nil
}

func (op *OutputProcessor) savePosts(posts []Post) error {
	allPostsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "all.json")
	if err := op.saveJSON(allPostsPath, posts); err != nil {
		return fmt.Errorf("failed to save all posts: %w", err)
	}

	for _, post := range posts {
		postPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts",
			fmt.Sprintf("post_%d.json", post.Index))
		if err := op.saveJSON(postPath, post); err != nil {
			return fmt.Errorf("failed to save post %d: %w", post.Index, err)
		}
	}

	return nil
}

func (op *OutputProcessor) saveTags(tags map[string][]int) error {
	allTagsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "all.json")
	if err := op.saveJSON(allTagsPath, tags); err != nil {
		return fmt.Errorf("failed to save all tags: %w", err)
	}

	for tag, postIndices := range tags {
		tagPath := filepath.Join(op.config.OutputDir, "public_html", "api", "tags",
			fmt.Sprintf("%s.json", tag))
		if err := op.saveJSON(tagPath, postIndices); err != nil {
			return fmt.Errorf("failed to save tag %s: %w", tag, err)
		}
	}

	return nil
}

func (op *OutputProcessor) saveJSON(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func (op *OutputProcessor) createDirectories() error {
	directories := []string{
		filepath.Join(op.config.OutputDir, "public_html", "api", "tags"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "posts"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
