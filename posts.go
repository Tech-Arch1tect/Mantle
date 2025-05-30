package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	ErrNoFrontMatter      = errors.New("no frontmatter found")
	ErrInvalidFrontMatter = errors.New("invalid frontmatter format")
)

type FrontMatter struct {
	Title    string   `yaml:"title"`
	Author   string   `yaml:"author"`
	Date     string   `yaml:"date"`
	Tags     []string `yaml:"tags"`
	Category string   `yaml:"category,omitempty"`
	Excerpt  string   `yaml:"excerpt,omitempty"`
}

func (fm FrontMatter) Validate() []string {
	var warnings []string
	if strings.TrimSpace(fm.Title) == "" {
		warnings = append(warnings, "title is empty")
	}
	if strings.TrimSpace(fm.Author) == "" {
		warnings = append(warnings, "author is empty")
	}
	if strings.TrimSpace(fm.Date) == "" {
		warnings = append(warnings, "date is empty")
	}
	return warnings
}

type Post struct {
	Index       int         `json:"index"`
	Markdown    string      `json:"markdown"`
	FrontMatter FrontMatter `json:"frontmatter"`
	Excerpt     string      `json:"excerpt"`
}

type PostLoaderInterface interface {
	LoadAll() ([]Post, error)
	Count() (int, error)
}

type PostLoader struct {
	contentDir string
	logger     *log.Logger
	fs         fs.FS
}

func NewPostLoader(contentDir string) *PostLoader {
	return &PostLoader{
		contentDir: contentDir,
		logger:     log.New(os.Stdout, "[PostLoader] ", log.LstdFlags),
		fs:         os.DirFS(contentDir),
	}
}

func (pl *PostLoader) LoadAll() ([]Post, error) {
	files, err := pl.listMarkdownFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list markdown files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no markdown files found in %s", pl.contentDir)
	}

	posts := make([]Post, 0, len(files))

	for index, file := range files {
		post, err := pl.loadPost(file, index)
		if err != nil {
			pl.logger.Printf("failed to load post %s: %v", file.Name(), err)
			continue
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (pl *PostLoader) listMarkdownFiles() ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(pl.fs, ".")
	if err != nil {
		return nil, err
	}

	var mdFiles []fs.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry)
		}
	}

	return mdFiles, nil
}

func (pl *PostLoader) loadPost(file fs.DirEntry, index int) (Post, error) {
	content, err := fs.ReadFile(pl.fs, file.Name())
	if err != nil {
		return Post{}, fmt.Errorf("failed to read file %s: %w", file.Name(), err)
	}

	frontMatter, body, err := pl.parseFrontMatter(string(content), file.Name())
	if err != nil {
		return Post{}, fmt.Errorf("failed to parse frontmatter for %s: %w", file.Name(), err)
	}

	excerpt := pl.generateExcerpt(frontMatter, body)

	return Post{
		Index:       index,
		Markdown:    body,
		FrontMatter: frontMatter,
		Excerpt:     excerpt,
	}, nil
}

func (pl *PostLoader) parseFrontMatter(content, filename string) (FrontMatter, string, error) {
	var fm FrontMatter

	if !strings.HasPrefix(content, "---") {
		return fm, content, ErrNoFrontMatter
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return fm, content, ErrInvalidFrontMatter
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return fm, content, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	if warnings := fm.Validate(); len(warnings) > 0 {
		pl.logger.Printf("warnings for %s: %v", filename, warnings)
	}

	return fm, strings.TrimSpace(parts[2]), nil
}

func (pl *PostLoader) generateExcerpt(fm FrontMatter, body string) string {
	if fm.Excerpt != "" {
		return fm.Excerpt
	}

	excerptSeparator := "<!--more-->"
	if idx := strings.Index(body, excerptSeparator); idx != -1 {
		return strings.TrimSpace(body[:idx])
	}

	return pl.autoGenerateExcerpt(body)
}

func (pl *PostLoader) autoGenerateExcerpt(body string) string {
	const maxWords = 200
	const maxChars = 1000

	lines := strings.Split(body, "\n")
	var cleanedLines []string
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "" {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	cleanedBody := strings.Join(cleanedLines, " ")

	words := strings.Fields(cleanedBody)

	var excerpt []string
	charCount := 0

	for i, word := range words {
		if i >= maxWords || charCount+len(word)+1 > maxChars {
			break
		}
		excerpt = append(excerpt, word)
		charCount += len(word) + 1
	}

	result := strings.Join(excerpt, " ")

	if len(words) > len(excerpt) {
		result += "..."
	}

	return result
}

func (pl *PostLoader) Count() (int, error) {
	files, err := pl.listMarkdownFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}
