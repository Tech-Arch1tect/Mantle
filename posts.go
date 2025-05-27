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
	Title  string   `yaml:"title"`
	Author string   `yaml:"author"`
	Date   string   `yaml:"date"`
	Tags   []string `yaml:"tags"`
}

type Post struct {
	Index       int         `json:"index"`
	Markdown    string      `json:"markdown"`
	FrontMatter FrontMatter `json:"frontmatter"`
}

type PostLoader struct {
	contentDir string
	logger     *log.Logger
}

func NewPostLoader(contentDir string) *PostLoader {
	return &PostLoader{
		contentDir: contentDir,
		logger:     log.New(os.Stdout, "[PostLoader] ", log.LstdFlags),
	}
}

func (pl *PostLoader) LoadAll() ([]Post, error) {
	files, err := pl.listMarkdownFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list markdown files: %w", err)
	}

	posts := make([]Post, 0, len(files))

	for index, file := range files {
		post, err := pl.loadPost(file, index)
		if err != nil {
			return nil, fmt.Errorf("failed to load post %s: %w", file.Name(), err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (pl *PostLoader) listMarkdownFiles() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(pl.contentDir)
	if err != nil {
		return nil, err
	}

	var mdFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry)
		}
	}

	return mdFiles, nil
}

func (pl *PostLoader) loadPost(file os.DirEntry, index int) (Post, error) {
	filePath := filepath.Join(pl.contentDir, file.Name())
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Post{}, err
	}

	frontMatter, body, err := pl.parseFrontMatter(string(content), file.Name())
	if err != nil {
		return Post{}, err
	}

	return Post{
		Index:       index,
		Markdown:    body,
		FrontMatter: frontMatter,
	}, nil
}

func (pl *PostLoader) parseFrontMatter(content string, filename string) (FrontMatter, string, error) {
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
		pl.logger.Printf("error parsing frontmatter for %s: %v", filename, err)
		return fm, content, err
	}

	pl.validateFrontMatter(&fm, filename)

	return fm, strings.TrimSpace(parts[2]), nil
}

func (pl *PostLoader) validateFrontMatter(fm *FrontMatter, filename string) {
	if strings.TrimSpace(fm.Title) == "" {
		pl.logger.Printf("warning: frontmatter 'title' is empty for %s", filename)
	}
	if strings.TrimSpace(fm.Author) == "" {
		pl.logger.Printf("warning: frontmatter 'author' is empty for %s", filename)
	}
	if strings.TrimSpace(fm.Date) == "" {
		pl.logger.Printf("warning: frontmatter 'date' is empty for %s", filename)
	}
}

func (pl *PostLoader) Count() (int, error) {
	files, err := pl.listMarkdownFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}
