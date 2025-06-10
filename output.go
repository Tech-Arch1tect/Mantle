package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var tokenRegex = regexp.MustCompile(`[A-Za-z0-9_]+`)

func tokenize(text string) []string {
	words := tokenRegex.FindAllString(strings.ToLower(text), -1)
	seen := make(map[string]struct{})
	var tokens []string
	for _, w := range words {
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			tokens = append(tokens, w)
		}
	}
	return tokens
}

// @Description Hierarchical category tree node
type CategoryTreeNode struct {
	Name      string             `json:"name" example:"Tutorials"`
	Path      string             `json:"path" example:"tech/tutorials"`
	PostCount int                `json:"postCount" example:"5"`
	Children  []CategoryTreeNode `json:"children,omitempty"`
}

// @Description Hierarchical category tree structure
type CategoryTree []CategoryTreeNode

// @Description Category information with full post details
type CategoryDetail struct {
	Info  CategoryInfo `json:"info"`
	Posts []Post       `json:"posts"`
}

// @Description Unified API metadata including counts, pagination info, and configuration
type MetadataResponse struct {
	Posts struct {
		Total      int                    `json:"total" example:"42"`
		PerPage    int                    `json:"perPage" example:"10"`
		TotalPages int                    `json:"totalPages" example:"5"`
		Newest     map[string]interface{} `json:"newest,omitempty"`
		Oldest     map[string]interface{} `json:"oldest,omitempty"`
	} `json:"posts"`
	Previews struct {
		Total      int `json:"total" example:"42"`
		PerPage    int `json:"perPage" example:"10"`
		TotalPages int `json:"totalPages" example:"5"`
	} `json:"previews"`
	Tags struct {
		Total int            `json:"total" example:"15"`
		Stats map[string]int `json:"stats"`
	} `json:"tags"`
	Categories struct {
		Total int            `json:"total" example:"8"`
		Stats map[string]int `json:"stats"`
	} `json:"categories"`
	Config struct {
		DateFormat         string `json:"dateFormat" example:"2006-01-02"`
		DateFormatReadable string `json:"dateFormatReadable" example:"yyyy-mm-dd"`
	} `json:"config"`
	Site struct {
		Name        string `json:"name" example:"My Site"`
		Description string `json:"description" example:"My Site Description"`
		Tagline     string `json:"tagline" example:"My Site Tagline"`
	} `json:"site"`
}

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

	if err := op.saveCategories(processedPosts.Categories, sortedPosts); err != nil {
		return fmt.Errorf("failed to save categories: %w", err)
	}

	if err := op.savePaginatedPosts(sortedPosts); err != nil {
		return fmt.Errorf("failed to save paginated posts: %w", err)
	}

	if err := op.saveRelatedPosts(processedPosts.RelatedPosts); err != nil {
		return fmt.Errorf("failed to save related posts: %w", err)
	}

	if err := op.saveSearchIndex(sortedPosts); err != nil {
		return fmt.Errorf("failed to save search index: %w", err)
	}

	if err := op.saveUnifiedMetadata(sortedPosts, processedPosts); err != nil {
		return fmt.Errorf("failed to save unified metadata: %w", err)
	}

	op.logger.Println("Output processed successfully")
	return nil
}

func (op *OutputProcessor) saveUnifiedMetadata(sortedPosts []Post, processedPosts ProcessedPosts) error {
	totalPosts := len(sortedPosts)
	postsPerPage := op.config.PostsPerPage
	previewsPerPage := op.config.PreviewsPerPage
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage
	totalPreviewPages := (totalPosts + previewsPerPage - 1) / previewsPerPage

	if totalPages == 0 {
		totalPages = 1
	}
	if totalPreviewPages == 0 {
		totalPreviewPages = 1
	}

	var oldestPost, newestPost map[string]interface{}
	if len(sortedPosts) > 0 {
		newest := sortedPosts[0]
		newestPost = map[string]interface{}{
			"slug":  newest.FrontMatter.Slug,
			"title": newest.FrontMatter.Title,
			"date":  newest.FrontMatter.Date,
		}
		oldest := sortedPosts[len(sortedPosts)-1]
		oldestPost = map[string]interface{}{
			"slug":  oldest.FrontMatter.Slug,
			"title": oldest.FrontMatter.Title,
			"date":  oldest.FrontMatter.Date,
		}
	}

	tagStats := make(map[string]int)
	for tag, postSlugs := range processedPosts.Tags {
		tagStats[tag] = len(postSlugs)
	}

	categoryStats := make(map[string]int)
	for _, categoryInfo := range processedPosts.Categories {
		categoryStats[categoryInfo.Path] = categoryInfo.PostCount
	}

	metadata := map[string]interface{}{
		"posts": map[string]interface{}{
			"total":      totalPosts,
			"perPage":    postsPerPage,
			"totalPages": totalPages,
			"newest":     newestPost,
			"oldest":     oldestPost,
		},
		"previews": map[string]interface{}{
			"total":      totalPosts,
			"perPage":    previewsPerPage,
			"totalPages": totalPreviewPages,
		},
		"tags": map[string]interface{}{
			"total": len(processedPosts.Tags),
			"stats": tagStats,
		},
		"categories": map[string]interface{}{
			"total": len(processedPosts.Categories),
			"stats": categoryStats,
		},
		"config": map[string]interface{}{
			"dateFormat":         op.config.DateFormat,
			"dateFormatReadable": op.convertDateFormatToReadable(op.config.DateFormat),
		},
		"site": map[string]interface{}{
			"name":        op.config.SiteName,
			"description": op.config.SiteDescription,
			"tagline":     op.config.SiteTagline,
		},
	}

	metaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "meta.json")
	if err := op.saveJSON(metaPath, metadata); err != nil {
		return fmt.Errorf("failed to save unified metadata: %w", err)
	}

	op.logger.Println("Saved unified metadata")
	return nil
}

func (op *OutputProcessor) saveRelatedPosts(relatedPosts map[string][]RelatedPost) error {
	allRelatedPath := filepath.Join(op.config.OutputDir, "public_html", "api", "related", "all.json")
	if err := op.saveJSON(allRelatedPath, relatedPosts); err != nil {
		return fmt.Errorf("failed to save all related posts: %w", err)
	}

	for postSlug, related := range relatedPosts {
		relatedPath := filepath.Join(op.config.OutputDir, "public_html", "api", "related",
			fmt.Sprintf("%s.json", postSlug))
		if err := op.saveJSON(relatedPath, related); err != nil {
			return fmt.Errorf("failed to save related posts for post %s: %w", postSlug, err)
		}
	}

	op.logger.Printf("Saved related posts for %d posts", len(relatedPosts))
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
			op.logger.Printf("Warning: Failed to parse date '%s' for post %s using format '%s': %v",
				post.FrontMatter.Date, post.FrontMatter.Slug, op.config.DateFormat, err)
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

func (op *OutputProcessor) saveCategories(categories map[string]CategoryInfo, allPosts []Post) error {
	allCategoriesPath := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "all.json")
	if err := op.saveJSON(allCategoriesPath, categories); err != nil {
		return fmt.Errorf("failed to save all categories: %w", err)
	}

	for categoryPath, info := range categories {
		safeFilename := strings.ReplaceAll(categoryPath, "/", "_")

		categoryPosts := make([]Post, 0, len(info.PostSlugs))
		for _, slug := range info.PostSlugs {
			for _, post := range allPosts {
				if post.FrontMatter.Slug == slug {
					categoryPosts = append(categoryPosts, post)
					break
				}
			}
		}

		sortedCategoryPosts, _ := op.sortPostsByDate(categoryPosts)

		categoryDetail := CategoryDetail{
			Info:  info,
			Posts: sortedCategoryPosts,
		}

		categoryPath := filepath.Join(op.config.OutputDir, "public_html", "api", "categories",
			fmt.Sprintf("%s.json", safeFilename))
		if err := op.saveJSON(categoryPath, categoryDetail); err != nil {
			return fmt.Errorf("failed to save category %s: %w", categoryPath, err)
		}
	}

	tree := op.buildCategoryTree(categories)
	treePath := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "tree.json")
	if err := op.saveJSON(treePath, tree); err != nil {
		return fmt.Errorf("failed to save category tree: %w", err)
	}

	op.logger.Printf("Saved %d categories", len(categories))

	return nil
}

func (op *OutputProcessor) buildCategoryTree(categories map[string]CategoryInfo) CategoryTree {
	var roots CategoryTree

	for path, info := range categories {
		if info.Parent == "" {
			roots = append(roots, op.buildTreeNode(path, categories))
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	return roots
}

func (op *OutputProcessor) buildTreeNode(path string, categories map[string]CategoryInfo) CategoryTreeNode {
	info := categories[path]
	node := CategoryTreeNode{
		Name:      info.Name,
		Path:      path,
		PostCount: info.PostCount,
	}

	for _, childPath := range info.Children {
		node.Children = append(node.Children, op.buildTreeNode(childPath, categories))
	}

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Name < node.Children[j].Name
	})

	return node
}

func (op *OutputProcessor) savePaginatedPosts(posts []Post) error {
	postsPerPage := op.config.PostsPerPage

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

		paginated := PostsResponse{
			Posts: posts[start:end],
			PaginationInfo: PaginationInfo{
				Page:        page,
				TotalPages:  totalPages,
				TotalItems:  len(posts),
				HasNext:     page < totalPages-1,
				HasPrevious: page > 0,
			},
		}

		if paginated.HasNext {
			nextPage := page + 1
			paginated.NextPage = &nextPage
		}
		if paginated.HasPrevious {
			prevPage := page - 1
			paginated.PrevPage = &prevPage
		}

		pagePath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-page",
			fmt.Sprintf("%d.json", page))
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
	for _, post := range posts {
		postPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-slug",
			fmt.Sprintf("%s.json", post.FrontMatter.Slug))
		if err := op.saveJSON(postPath, post); err != nil {
			return fmt.Errorf("failed to save post %s: %w", post.FrontMatter.Slug, err)
		}
	}

	allPostsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "all.json")
	if err := op.saveJSON(allPostsPath, posts); err != nil {
		return fmt.Errorf("failed to save all posts: %w", err)
	}

	return nil
}

func (op *OutputProcessor) savePostPreviews(posts []Post) error {
	previews := make([]PostPreview, 0, len(posts))
	for _, post := range posts {
		preview := PostPreview{
			FrontMatter: post.FrontMatter,
			Excerpt:     post.Excerpt,
			ReadingTime: post.ReadingTime,
		}
		previews = append(previews, preview)
	}

	for _, preview := range previews {
		previewPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-slug",
			fmt.Sprintf("%s.json", preview.FrontMatter.Slug))
		if err := op.saveJSON(previewPath, preview); err != nil {
			return fmt.Errorf("failed to save preview %s: %w", preview.FrontMatter.Slug, err)
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

		paginated := PreviewsResponse{
			Previews: previews[start:end],
			PaginationInfo: PaginationInfo{
				Page:        page,
				TotalPages:  totalPages,
				TotalItems:  len(previews),
				HasNext:     page < totalPages-1,
				HasPrevious: page > 0,
			},
		}

		if paginated.HasNext {
			nextPage := page + 1
			paginated.NextPage = &nextPage
		}
		if paginated.HasPrevious {
			prevPage := page - 1
			paginated.PrevPage = &prevPage
		}

		pagePath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-page",
			fmt.Sprintf("%d.json", page))
		if err := op.saveJSON(pagePath, paginated); err != nil {
			return fmt.Errorf("failed to save preview page %d: %w", page, err)
		}
	}

	allPreviewsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "all.json")
	if err := op.saveJSON(allPreviewsPath, previews); err != nil {
		return fmt.Errorf("failed to save all previews: %w", err)
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

func (op *OutputProcessor) saveTags(tags map[string][]string) error {
	allTagsPath := filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "all.json")
	if err := op.saveJSON(allTagsPath, tags); err != nil {
		return fmt.Errorf("failed to save all tags: %w", err)
	}

	for tag, postSlugs := range tags {
		tagPath := filepath.Join(op.config.OutputDir, "public_html", "api", "tags",
			fmt.Sprintf("%s.json", tag))
		if err := op.saveJSON(tagPath, postSlugs); err != nil {
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
		filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-slug"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-slug"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "categories"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "related"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "search"),
	}
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (op *OutputProcessor) saveSearchIndex(posts []Post) error {
	inverted := make(map[string][]string)
	for _, p := range posts {
		text := p.FrontMatter.Title + " " + strings.Join(p.FrontMatter.Tags, " ") + " " + p.Excerpt
		toks := tokenize(text)
		for _, t := range toks {
			inverted[t] = append(inverted[t], p.FrontMatter.Slug)
		}
	}
	for term, list := range inverted {
		seen := make(map[string]struct{})
		var unique []string
		for _, slug := range list {
			if _, ok := seen[slug]; !ok {
				seen[slug] = struct{}{}
				unique = append(unique, slug)
			}
		}
		sort.Strings(unique)
		inverted[term] = unique
	}
	path := filepath.Join(op.config.OutputDir, "public_html", "api", "search", "inverted.json")
	return op.saveJSON(path, inverted)
}
