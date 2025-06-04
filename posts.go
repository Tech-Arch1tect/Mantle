package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	ErrNoFrontMatter      = errors.New("no frontmatter found")
	ErrInvalidFrontMatter = errors.New("invalid frontmatter format")
)

// @Description Post frontmatter containing metadata
type FrontMatter struct {
	Title    string   `yaml:"title" json:"title" example:"Getting Started with Go"`
	Author   string   `yaml:"author" json:"author" example:"John Doe"`
	Date     string   `yaml:"date" json:"date" example:"2024-01-15"`
	Tags     []string `yaml:"tags" json:"tags" example:"golang,tutorial,beginner"`
	Category string   `yaml:"category,omitempty" json:"category,omitempty" example:"tech/tutorials"`
	Excerpt  string   `yaml:"excerpt,omitempty" json:"excerpt,omitempty" example:"Learn the basics of Go programming language"`
	Slug     string   `yaml:"slug,omitempty" json:"slug,omitempty" example:"getting-started-with-go"`
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

// @Description Complete blog post including markdown content and frontmatter
type Post struct {
	Markdown    string      `json:"markdown" example:"# Getting Started with Go\n\nThis is the content..."`
	FrontMatter FrontMatter `json:"frontmatter"`
	Excerpt     string      `json:"excerpt" example:"This is a brief excerpt of the post..."`
	ReadingTime int         `json:"readingTime" example:"5"`
}

// @Description Post preview containing frontmatter, excerpt, and reading time
type PostPreview struct {
	FrontMatter FrontMatter `json:"frontmatter"`
	Excerpt     string      `json:"excerpt" example:"This is a brief excerpt of the post..."`
	ReadingTime int         `json:"readingTime" example:"5"`
}

// @Description Error response format
type ErrorResponse struct {
	Error   string `json:"error" example:"Not found"`
	Message string `json:"message,omitempty" example:"The requested resource was not found"`
}

// @Description Pagination information for responses
type PaginationInfo struct {
	Page        int  `json:"page" example:"0"`
	TotalPages  int  `json:"totalPages" example:"5"`
	TotalItems  int  `json:"totalItems" example:"42"`
	HasNext     bool `json:"hasNext" example:"true"`
	HasPrevious bool `json:"hasPrevious" example:"false"`
	NextPage    *int `json:"nextPage,omitempty" example:"1"`
	PrevPage    *int `json:"prevPage,omitempty" example:"0"`
}

// @Description Paginated response containing posts and pagination metadata
type PostsResponse struct {
	Posts []Post `json:"posts"`
	PaginationInfo
}

// @Description Paginated response containing post previews and pagination metadata
type PreviewsResponse struct {
	Previews []PostPreview `json:"previews"`
	PaginationInfo
}

type PostLoaderInterface interface {
	LoadAll() ([]Post, error)
	Count() (int, error)
}

type PostLoader struct {
	contentDir            string
	logger                *log.Logger
	fs                    fs.FS
	averageWordsPerMinute int
}

func NewPostLoader(contentDir string, averageWordsPerMinute int) *PostLoader {
	return &PostLoader{
		contentDir:            contentDir,
		logger:                log.New(os.Stdout, "[PostLoader] ", log.LstdFlags),
		fs:                    os.DirFS(contentDir),
		averageWordsPerMinute: averageWordsPerMinute,
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
	usedSlugs := make(map[string]bool)

	for _, file := range files {
		post, err := pl.loadPost(file, usedSlugs)
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

func (pl *PostLoader) loadPost(file fs.DirEntry, usedSlugs map[string]bool) (Post, error) {
	content, err := fs.ReadFile(pl.fs, file.Name())
	if err != nil {
		return Post{}, fmt.Errorf("failed to read file %s: %w", file.Name(), err)
	}

	frontMatter, body, err := pl.parseFrontMatter(string(content), file.Name())
	if err != nil {
		return Post{}, fmt.Errorf("failed to parse frontmatter for %s: %w", file.Name(), err)
	}

	if frontMatter.Slug == "" {
		frontMatter.Slug = pl.generateSlug(frontMatter.Title)
	}

	frontMatter.Slug = pl.ensureUniqueSlug(frontMatter.Slug, usedSlugs)
	usedSlugs[frontMatter.Slug] = true

	excerpt := pl.generateExcerpt(frontMatter, body)
	readingTime := pl.calculateReadingTime(body)

	return Post{
		Markdown:    body,
		FrontMatter: frontMatter,
		Excerpt:     excerpt,
		ReadingTime: readingTime,
	}, nil
}

func (pl *PostLoader) ensureUniqueSlug(baseSlug string, usedSlugs map[string]bool) string {
	slug := baseSlug
	counter := 1

	for usedSlugs[slug] {
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	return slug
}

func (pl *PostLoader) generateSlug(title string) string {
	slug := strings.ToLower(title)

	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = reg.ReplaceAllString(slug, "")

	spaceReg := regexp.MustCompile(`\s+`)
	slug = spaceReg.ReplaceAllString(slug, "-")

	dashReg := regexp.MustCompile(`-+`)
	slug = dashReg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "untitled"
	}

	return slug
}

func (pl *PostLoader) calculateReadingTime(markdown string) int {
	words := strings.Fields(strings.ReplaceAll(markdown, "\n", " "))
	wordCount := len(words)

	averageWordsPerMinute := pl.averageWordsPerMinute
	readingTimeMinutes := (wordCount + averageWordsPerMinute - 1) / averageWordsPerMinute

	if readingTimeMinutes < 1 {
		return 1
	}

	return readingTimeMinutes
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
