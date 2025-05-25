package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type FrontMatter struct {
	Title  string `yaml:"title"`
	Author string `yaml:"author"`
	Date   string `yaml:"date"`
}

type Post struct {
	Markdown    string      `json:"markdown"`
	FrontMatter FrontMatter `json:"frontmatter"`
}

func loadPosts(contentDir string) ([]Post, error) {
	posts := []Post{}

	files, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(contentDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		fm, body, err := parseFrontMatter(string(content), file.Name())
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %w", file.Name(), err)
		}

		posts = append(posts, Post{
			Markdown:    body,
			FrontMatter: fm,
		})
	}

	return posts, nil
}

func parseFrontMatter(content string, filename string) (FrontMatter, string, error) {
	var fm FrontMatter
	if !strings.HasPrefix(content, "---") {
		return fm, content, errors.New("no frontmatter found")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return fm, content, fmt.Errorf("invalid frontmatter format")
	}

	err := yaml.Unmarshal([]byte(parts[1]), &fm)
	if err != nil {
		log.Printf("error parsing frontmatter for %s: %v", filename, err)
		return fm, content, err
	}

	if strings.TrimSpace(fm.Title) == "" {
		log.Printf("warning: frontmatter 'title' is empty for %s", filename)
	}
	if strings.TrimSpace(fm.Author) == "" {
		log.Printf("warning: frontmatter 'author' is empty for %s", filename)
	}
	if strings.TrimSpace(fm.Date) == "" {
		log.Printf("warning: frontmatter 'date' is empty for %s", filename)
	}

	return fm, strings.TrimSpace(parts[2]), nil
}
