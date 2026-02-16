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
		Total      int            `json:"total" example:"15"`
		PerPage    int            `json:"perPage" example:"20"`
		TotalPages int            `json:"totalPages" example:"1"`
		Stats      map[string]int `json:"stats"`
	} `json:"tags"`
	Categories struct {
		Total      int            `json:"total" example:"8"`
		PerPage    int            `json:"perPage" example:"20"`
		TotalPages int            `json:"totalPages" example:"1"`
		Stats      map[string]int `json:"stats"`
	} `json:"categories"`
	Related struct {
		PerPage int `json:"perPage" example:"5"`
	} `json:"related"`
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

	if err := op.savePaginatedPosts(sortedPosts); err != nil {
		return fmt.Errorf("failed to save paginated posts: %w", err)
	}

	if err := op.savePostPreviews(sortedPosts); err != nil {
		return fmt.Errorf("failed to save post previews: %w", err)
	}

	if err := op.saveTags(processedPosts.Tags, sortedPosts); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	if err := op.saveCategories(processedPosts.Categories, sortedPosts); err != nil {
		return fmt.Errorf("failed to save categories: %w", err)
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

func paginateAndSave[T any](op *OutputProcessor, items []T, perPage int, baseDir string) (PaginationInfo, error) {
	totalItems := len(items)
	totalPages := (totalItems + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	var overallPagination PaginationInfo

	for page := 0; page < totalPages; page++ {
		start := page * perPage
		end := start + perPage
		if end > totalItems {
			end = totalItems
		}

		var pageItems []T
		if start < totalItems {
			pageItems = items[start:end]
		} else {
			pageItems = []T{}
		}

		pagination := PaginationInfo{
			Page:        page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			HasNext:     page < totalPages-1,
			HasPrevious: page > 0,
		}

		if pagination.HasNext {
			nextPage := page + 1
			pagination.NextPage = &nextPage
		}
		if pagination.HasPrevious {
			prevPage := page - 1
			pagination.PrevPage = &prevPage
		}

		envelope := PaginatedResponse{
			Data:       pageItems,
			Pagination: pagination,
		}

		pagePath := filepath.Join(baseDir, fmt.Sprintf("%d.json", page))
		if err := op.saveJSON(pagePath, envelope); err != nil {
			return PaginationInfo{}, fmt.Errorf("failed to save page %d: %w", page, err)
		}

		if page == 0 {
			overallPagination = pagination
		}
	}

	overallPagination.TotalItems = totalItems
	overallPagination.TotalPages = totalPages
	return overallPagination, nil
}

func (op *OutputProcessor) saveSingleItem(item interface{}, path string) error {
	envelope := SingleResponse{Data: item}
	return op.saveJSON(path, envelope)
}

func (op *OutputProcessor) savePosts(posts []Post) error {
	for _, post := range posts {
		postPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-slug",
			fmt.Sprintf("%s.json", post.FrontMatter.Slug))
		if err := op.saveSingleItem(post, postPath); err != nil {
			return fmt.Errorf("failed to save post %s: %w", post.FrontMatter.Slug, err)
		}
	}
	return nil
}

func (op *OutputProcessor) savePaginatedPosts(posts []Post) error {
	baseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-page")
	_, err := paginateAndSave(op, posts, op.config.PostsPerPage, baseDir)
	if err != nil {
		return err
	}

	totalPages := (len(posts) + op.config.PostsPerPage - 1) / op.config.PostsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	metadata := map[string]interface{}{
		"totalItems":         len(posts),
		"perPage":            op.config.PostsPerPage,
		"totalPages":         totalPages,
		"dateFormat":         op.config.DateFormat,
		"dateFormatReadable": op.convertDateFormatToReadable(op.config.DateFormat),
	}

	metaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "meta.json")
	if err := op.saveJSON(metaPath, metadata); err != nil {
		return fmt.Errorf("failed to save posts metadata: %w", err)
	}

	op.logger.Printf("Created %d post pagination pages with %d posts per page", totalPages, op.config.PostsPerPage)
	return nil
}

func (op *OutputProcessor) savePostPreviews(posts []Post) error {
	previews := make([]PostPreview, 0, len(posts))
	for _, post := range posts {
		previews = append(previews, PostPreview{
			FrontMatter: post.FrontMatter,
			Excerpt:     post.Excerpt,
			ReadingTime: post.ReadingTime,
		})
	}

	for _, preview := range previews {
		previewPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-slug",
			fmt.Sprintf("%s.json", preview.FrontMatter.Slug))
		if err := op.saveSingleItem(preview, previewPath); err != nil {
			return fmt.Errorf("failed to save preview %s: %w", preview.FrontMatter.Slug, err)
		}
	}

	baseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-page")
	_, err := paginateAndSave(op, previews, op.config.PreviewsPerPage, baseDir)
	if err != nil {
		return err
	}

	totalPages := (len(previews) + op.config.PreviewsPerPage - 1) / op.config.PreviewsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	previewMeta := map[string]interface{}{
		"totalItems": len(previews),
		"perPage":    op.config.PreviewsPerPage,
		"totalPages": totalPages,
	}

	previewMetaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "meta.json")
	if err := op.saveJSON(previewMetaPath, previewMeta); err != nil {
		return fmt.Errorf("failed to save preview metadata: %w", err)
	}

	return nil
}

func (op *OutputProcessor) saveTags(tags map[string][]string, allPosts []Post) error {
	tagInfos := make([]TagInfo, 0, len(tags))
	for tag, postSlugs := range tags {
		tagInfos = append(tagInfos, TagInfo{
			Name:      tag,
			PostCount: len(postSlugs),
		})
	}
	sort.Slice(tagInfos, func(i, j int) bool {
		return tagInfos[i].Name < tagInfos[j].Name
	})

	baseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "by-page")
	_, err := paginateAndSave(op, tagInfos, op.config.TagsPerPage, baseDir)
	if err != nil {
		return fmt.Errorf("failed to paginate tags: %w", err)
	}

	postBySlug := make(map[string]Post, len(allPosts))
	for _, post := range allPosts {
		postBySlug[post.FrontMatter.Slug] = post
	}

	for tag, postSlugs := range tags {
		var previews []PostPreview
		for _, slug := range postSlugs {
			if post, ok := postBySlug[slug]; ok {
				previews = append(previews, PostPreview{
					FrontMatter: post.FrontMatter,
					Excerpt:     post.Excerpt,
					ReadingTime: post.ReadingTime,
				})
			}
		}
		sort.Slice(previews, func(i, j int) bool {
			return previews[i].FrontMatter.Date > previews[j].FrontMatter.Date
		})

		tagBaseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "by-name", tag, "by-page")
		if _, err := paginateAndSave(op, previews, op.config.TagsPerPage, tagBaseDir); err != nil {
			return fmt.Errorf("failed to paginate tag %s: %w", tag, err)
		}
	}

	totalTagPages := (len(tagInfos) + op.config.TagsPerPage - 1) / op.config.TagsPerPage
	if totalTagPages == 0 {
		totalTagPages = 1
	}

	tagStats := make(map[string]int, len(tags))
	for tag, slugs := range tags {
		tagStats[tag] = len(slugs)
	}

	tagMeta := map[string]interface{}{
		"totalItems": len(tagInfos),
		"perPage":    op.config.TagsPerPage,
		"totalPages": totalTagPages,
		"stats":      tagStats,
	}

	tagMetaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "meta.json")
	if err := op.saveJSON(tagMetaPath, tagMeta); err != nil {
		return fmt.Errorf("failed to save tags metadata: %w", err)
	}

	op.logger.Printf("Saved %d tags", len(tags))
	return nil
}

func (op *OutputProcessor) saveCategories(categories map[string]CategoryInfo, allPosts []Post) error {
	catInfos := make([]CategoryInfo, 0, len(categories))
	for _, info := range categories {
		catInfos = append(catInfos, info)
	}
	sort.Slice(catInfos, func(i, j int) bool {
		return catInfos[i].Path < catInfos[j].Path
	})

	baseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "by-page")
	if _, err := paginateAndSave(op, catInfos, op.config.CategoriesPerPage, baseDir); err != nil {
		return fmt.Errorf("failed to paginate categories: %w", err)
	}

	postBySlug := make(map[string]Post, len(allPosts))
	for _, post := range allPosts {
		postBySlug[post.FrontMatter.Slug] = post
	}

	for catPath, info := range categories {
		var previews []PostPreview
		for _, slug := range info.PostSlugs {
			if post, ok := postBySlug[slug]; ok {
				previews = append(previews, PostPreview{
					FrontMatter: post.FrontMatter,
					Excerpt:     post.Excerpt,
					ReadingTime: post.ReadingTime,
				})
			}
		}
		sort.Slice(previews, func(i, j int) bool {
			return previews[i].FrontMatter.Date > previews[j].FrontMatter.Date
		})

		catBaseDir := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "by-path", catPath, "by-page")
		if _, err := paginateAndSave(op, previews, op.config.CategoriesPerPage, catBaseDir); err != nil {
			return fmt.Errorf("failed to paginate category %s: %w", catPath, err)
		}
	}

	tree := op.buildCategoryTree(categories)
	treePath := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "tree.json")
	if err := op.saveSingleItem(tree, treePath); err != nil {
		return fmt.Errorf("failed to save category tree: %w", err)
	}

	totalCatPages := (len(catInfos) + op.config.CategoriesPerPage - 1) / op.config.CategoriesPerPage
	if totalCatPages == 0 {
		totalCatPages = 1
	}

	categoryStats := make(map[string]int, len(categories))
	for _, info := range categories {
		categoryStats[info.Path] = info.PostCount
	}

	catMeta := map[string]interface{}{
		"totalItems": len(catInfos),
		"perPage":    op.config.CategoriesPerPage,
		"totalPages": totalCatPages,
		"stats":      categoryStats,
	}

	catMetaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "meta.json")
	if err := op.saveJSON(catMetaPath, catMeta); err != nil {
		return fmt.Errorf("failed to save categories metadata: %w", err)
	}

	op.logger.Printf("Saved %d categories", len(categories))
	return nil
}

func (op *OutputProcessor) saveRelatedPosts(relatedPosts map[string][]RelatedPost) error {
	for postSlug, related := range relatedPosts {
		pagination := PaginationInfo{
			Page:        0,
			TotalPages:  1,
			TotalItems:  len(related),
			HasNext:     false,
			HasPrevious: false,
		}

		envelope := PaginatedResponse{
			Data:       related,
			Pagination: pagination,
		}

		relatedPath := filepath.Join(op.config.OutputDir, "public_html", "api", "related", "by-slug",
			fmt.Sprintf("%s.json", postSlug))
		if err := op.saveJSON(relatedPath, envelope); err != nil {
			return fmt.Errorf("failed to save related posts for post %s: %w", postSlug, err)
		}
	}

	relatedMeta := map[string]interface{}{
		"totalItems": len(relatedPosts),
		"perPage":    op.config.RelatedPerPage,
		"totalPages": 1,
	}

	relatedMetaPath := filepath.Join(op.config.OutputDir, "public_html", "api", "related", "meta.json")
	if err := op.saveJSON(relatedMetaPath, relatedMeta); err != nil {
		return fmt.Errorf("failed to save related metadata: %w", err)
	}

	op.logger.Printf("Saved related posts for %d posts", len(relatedPosts))
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

	path := filepath.Join(op.config.OutputDir, "public_html", "api", "search", "index.json")
	return op.saveSingleItem(inverted, path)
}

func (op *OutputProcessor) saveUnifiedMetadata(sortedPosts []Post, processedPosts ProcessedPosts) error {
	totalPosts := len(sortedPosts)
	postsPerPage := op.config.PostsPerPage
	previewsPerPage := op.config.PreviewsPerPage
	tagsPerPage := op.config.TagsPerPage
	categoriesPerPage := op.config.CategoriesPerPage

	totalPostPages := (totalPosts + postsPerPage - 1) / postsPerPage
	totalPreviewPages := (totalPosts + previewsPerPage - 1) / previewsPerPage
	totalTagPages := (len(processedPosts.Tags) + tagsPerPage - 1) / tagsPerPage
	totalCatPages := (len(processedPosts.Categories) + categoriesPerPage - 1) / categoriesPerPage

	if totalPostPages == 0 {
		totalPostPages = 1
	}
	if totalPreviewPages == 0 {
		totalPreviewPages = 1
	}
	if totalTagPages == 0 {
		totalTagPages = 1
	}
	if totalCatPages == 0 {
		totalCatPages = 1
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
			"totalPages": totalPostPages,
			"newest":     newestPost,
			"oldest":     oldestPost,
		},
		"previews": map[string]interface{}{
			"total":      totalPosts,
			"perPage":    previewsPerPage,
			"totalPages": totalPreviewPages,
		},
		"tags": map[string]interface{}{
			"total":      len(processedPosts.Tags),
			"perPage":    tagsPerPage,
			"totalPages": totalTagPages,
			"stats":      tagStats,
		},
		"categories": map[string]interface{}{
			"total":      len(processedPosts.Categories),
			"perPage":    categoriesPerPage,
			"totalPages": totalCatPages,
			"stats":      categoryStats,
		},
		"related": map[string]interface{}{
			"perPage": op.config.RelatedPerPage,
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
		filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-slug"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "posts", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-slug"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "previews", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "tags", "by-name"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "by-page"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "categories", "by-path"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "related", "by-slug"),
		filepath.Join(op.config.OutputDir, "public_html", "api", "search"),
	}
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
