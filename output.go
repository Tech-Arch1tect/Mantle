package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

	sortedPosts, err := op.sortPostsByDate(processedPosts.Posts)
	if err != nil {
		return fmt.Errorf("failed to sort posts: %w", err)
	}

	if err := op.savePosts(sortedPosts); err != nil {
		return fmt.Errorf("failed to save posts: %w", err)
	}

	if err := op.savePostPreviews(sortedPosts); err != nil {
		return fmt.Errorf("failed to save post previews: %w", err)
	}

	if err := op.saveTags(processedPosts.Tags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	if err := op.savePaginatedPosts(sortedPosts); err != nil {
		return fmt.Errorf("failed to save paginated posts: %w", err)
	}

	op.logger.Println("Output processed successfully")
	return nil
}

func (op *OutputProcessor) sortPostsByDate(posts []Post) ([]Post, error) {
	sorted := make([]Post, len(posts))
	copy(sorted, posts)

	type postWithDate struct {
		post Post
		date time.Time
	}

	postsWithDates := make([]postWithDate, 0, len(sorted))

	for _, post := range sorted {
		parsedDate, err := time.Parse(op.config.DateFormat, post.FrontMatter.Date)
		if err != nil {
			op.logger.Printf("Warning: Failed to parse date '%s' for post %d using format '%s': %v",
				post.FrontMatter.Date, post.Index, op.config.DateFormat, err)
			parsedDate = time.Unix(0, 0)
		}
		postsWithDates = append(postsWithDates, postWithDate{
			post: post,
			date: parsedDate,
		})
	}

	sort.Slice(postsWithDates, func(i, j int) bool {
		return postsWithDates[i].date.After(postsWithDates[j].date)
	})

	for i, pwd := range postsWithDates {
		sorted[i] = pwd.post
	}

	return sorted, nil
}

func (op *OutputProcessor) savePaginatedPosts(posts []Post) error {
	postsPerPage := op.config.PostsPerPage

	type PaginatedResponse struct {
		Posts       []Post `json:"posts"`
		Page        int    `json:"page"`
		TotalPages  int    `json:"totalPages"`
		TotalPosts  int    `json:"totalPosts"`
		HasNext     bool   `json:"hasNext"`
		HasPrevious bool   `json:"hasPrevious"`
		NextPage    *int   `json:"nextPage,omitempty"`
		PrevPage    *int   `json:"prevPage,omitempty"`
	}

	totalPages := (len(posts) + postsPerPage - 1) / postsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	for page := 0; page < totalPages; page++ {
		start := page * postsPerPage
		end := start + postsPerPage
		if end > len(posts) {
			end = len(posts)
		}

		paginated := PaginatedResponse{
			Posts:       posts[start:end],
			Page:        page,
			TotalPages:  totalPages,
			TotalPosts:  len(posts),
			HasNext:     page < totalPages-1,
			HasPrevious: page > 0,
		}

		if paginated.HasNext {
			nextPage := page + 1
			paginated.NextPage = &nextPage
		}
		if paginated.HasPrevious {
			prevPage := page - 1
			paginated.PrevPage = &prevPage
		}

		pagePath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts",
			fmt.Sprintf("page_%d.json", page))
		if err := op.saveJSON(pagePath, paginated); err != nil {
			return fmt.Errorf("failed to save page %d: %w", page, err)
		}
	}

	metadata := map[string]interface{}{
		"totalPages":         totalPages,
		"totalPosts":         len(posts),
		"postsPerPage":       postsPerPage,
		"dateFormat":         op.config.DateFormat,
		"dateFormatReadable": op.convertDateFormatToReadable(op.config.DateFormat),
	}

	metaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "meta.json")
	if err := op.saveJSON(metaPath, metadata); err != nil {
		return fmt.Errorf("failed to save pagination metadata: %w", err)
	}

	op.logger.Printf("Created %d pagination pages with %d posts per page", totalPages, postsPerPage)

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

func (op *OutputProcessor) savePostPreviews(posts []Post) error {
	type PostPreview struct {
		Index       int         `json:"index"`
		FrontMatter FrontMatter `json:"frontmatter"`
		Excerpt     string      `json:"excerpt"`
	}

	previews := make([]PostPreview, 0, len(posts))
	for _, post := range posts {
		preview := PostPreview{
			Index:       post.Index,
			FrontMatter: post.FrontMatter,
			Excerpt:     post.Excerpt,
		}
		previews = append(previews, preview)
	}

	previewsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "all.json")
	if err := op.saveJSON(previewsPath, previews); err != nil {
		return fmt.Errorf("failed to save post previews: %w", err)
	}

	for _, preview := range previews {
		previewPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews",
			fmt.Sprintf("preview_%d.json", preview.Index))
		if err := op.saveJSON(previewPath, preview); err != nil {
			return fmt.Errorf("failed to save preview %d: %w", preview.Index, err)
		}
	}

	previewsPerPage := op.config.PreviewsPerPage
	totalPages := (len(previews) + previewsPerPage - 1) / previewsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	for page := 0; page < totalPages; page++ {
		start := page * previewsPerPage
		end := start + previewsPerPage
		if end > len(previews) {
			end = len(previews)
		}

		type PaginatedPreviews struct {
			Previews    []PostPreview `json:"previews"`
			Page        int           `json:"page"`
			TotalPages  int           `json:"totalPages"`
			TotalItems  int           `json:"totalItems"`
			HasNext     bool          `json:"hasNext"`
			HasPrevious bool          `json:"hasPrevious"`
			NextPage    *int          `json:"nextPage,omitempty"`
			PrevPage    *int          `json:"prevPage,omitempty"`
		}

		paginated := PaginatedPreviews{
			Previews:    previews[start:end],
			Page:        page,
			TotalPages:  totalPages,
			TotalItems:  len(previews),
			HasNext:     page < totalPages-1,
			HasPrevious: page > 0,
		}

		if paginated.HasNext {
			nextPage := page + 1
			paginated.NextPage = &nextPage
		}
		if paginated.HasPrevious {
			prevPage := page - 1
			paginated.PrevPage = &prevPage
		}

		pagePath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews",
			fmt.Sprintf("page_%d.json", page))
		if err := op.saveJSON(pagePath, paginated); err != nil {
			return fmt.Errorf("failed to save preview page %d: %w", page, err)
		}
	}

	previewMeta := map[string]interface{}{
		"totalPages":      totalPages,
		"totalPreviews":   len(previews),
		"previewsPerPage": previewsPerPage,
	}

	previewMetaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "meta.json")
	if err := op.saveJSON(previewMetaPath, previewMeta); err != nil {
		return fmt.Errorf("failed to save preview metadata: %w", err)
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

func (op *OutputProcessor) convertDateFormatToReadable(goFormat string) string {
	replacements := map[string]string{
		"2006":    "yyyy",
		"06":      "yy",
		"01":      "mm",
		"1":       "m",
		"Jan":     "mmm",
		"January": "mmmm",
		"02":      "dd",
		"2":       "d",
		"_2":      "d",
		"15":      "HH",
		"3":       "h",
		"03":      "hh",
		"04":      "MM",
		"4":       "M",
		"05":      "SS",
		"5":       "S",
		"PM":      "AM/PM",
		"pm":      "am/pm",
		"MST":     "tz",
		"Z07:00":  "±hh:mm",
		"Z0700":   "±hhmm",
		"Z07":     "±hh",
	}

	readable := goFormat

	keys := make([]string, 0, len(replacements))
	for k := range replacements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, goComponent := range keys {
		readable = strings.ReplaceAll(readable, goComponent, replacements[goComponent])
	}

	return readable
}

func (op *OutputProcessor) saveJSON(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
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
		filepath.Join(op.config.OutputDir, "public_html", "api", "previews"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
