package main

import (
	"sort"
	"strings"
)

type PostProcessor interface {
	Process(posts []Post) ProcessedPosts
}

// @Description Category information including hierarchy and post associations
type CategoryInfo struct {
	Name      string   `json:"name" example:"Tutorials"`
	Path      string   `json:"path" example:"tech/tutorials"`
	PostSlugs []string `json:"postSlugs" example:"getting-started-with-go,advanced-go-patterns"`
	PostCount int      `json:"postCount" example:"2"`
	Parent    string   `json:"parent,omitempty" example:"tech"`
	Children  []string `json:"children,omitempty" example:"golang,javascript"`
}

// @Description Related post information with similarity metrics
type RelatedPost struct {
	Slug        string `json:"slug" example:"advanced-go-patterns"`
	Title       string `json:"title" example:"Advanced Go Patterns"`
	Date        string `json:"date" example:"2024-01-20"`
	CommonTags  int    `json:"commonTags" example:"3"`
	ReadingTime int    `json:"readingTime" example:"8"`
}

// @Description Complete processed blog data including posts, tags, categories, and relationships
type ProcessedPosts struct {
	Posts        []Post                   `json:"posts"`
	Tags         map[string][]string      `json:"tags"`
	Categories   map[string]CategoryInfo  `json:"categories"`
	RelatedPosts map[string][]RelatedPost `json:"relatedPosts"`
}

// @Description Mapping of tag names to arrays of post slugs
type TagsMap map[string][]string

// @Description Mapping of category paths to category information
type CategoriesMap map[string]CategoryInfo

// @Description Mapping of post slugs to arrays of related posts
type RelatedPostsMap map[string][]RelatedPost

// @Description Inverted search index mapping terms to post slugs for client-side search
type SearchIndex map[string][]string

type DefaultPostProcessor struct{}

func NewPostProcessor() PostProcessor {
	return &DefaultPostProcessor{}
}

func (pp *DefaultPostProcessor) Process(posts []Post) ProcessedPosts {
	processedPosts := ProcessedPosts{
		Posts:        make([]Post, 0, len(posts)),
		Tags:         make(map[string][]string),
		Categories:   make(map[string]CategoryInfo),
		RelatedPosts: make(map[string][]RelatedPost),
	}

	for _, post := range posts {
		processedPosts.Posts = append(processedPosts.Posts, post)

		for _, tag := range post.FrontMatter.Tags {
			processedPosts.Tags[tag] = append(processedPosts.Tags[tag], post.FrontMatter.Slug)
		}

		if post.FrontMatter.Category != "" {
			pp.processCategory(post.FrontMatter.Category, post.FrontMatter.Slug, processedPosts.Categories)
		}
	}

	pp.buildCategoryHierarchy(processedPosts.Categories)
	pp.buildRelatedPosts(posts, processedPosts.RelatedPosts)

	return processedPosts
}

func (pp *DefaultPostProcessor) buildRelatedPosts(posts []Post, relatedPosts map[string][]RelatedPost) {
	for i, post := range posts {
		var related []RelatedPost

		for j, otherPost := range posts {
			if i == j {
				continue
			}

			commonTags := pp.countCommonTags(post.FrontMatter.Tags, otherPost.FrontMatter.Tags)
			if commonTags > 0 {
				related = append(related, RelatedPost{
					Slug:        otherPost.FrontMatter.Slug,
					Title:       otherPost.FrontMatter.Title,
					Date:        otherPost.FrontMatter.Date,
					CommonTags:  commonTags,
					ReadingTime: otherPost.ReadingTime,
				})
			}
		}

		sort.Slice(related, func(i, j int) bool {
			if related[i].CommonTags != related[j].CommonTags {
				return related[i].CommonTags > related[j].CommonTags
			}
			return related[i].Date > related[j].Date
		})

		if len(related) > 5 {
			related = related[:5]
		}

		relatedPosts[post.FrontMatter.Slug] = related
	}
}

func (pp *DefaultPostProcessor) countCommonTags(tags1, tags2 []string) int {
	tagMap := make(map[string]bool)
	for _, tag := range tags1 {
		tagMap[tag] = true
	}

	count := 0
	for _, tag := range tags2 {
		if tagMap[tag] {
			count++
		}
	}
	return count
}

func (pp *DefaultPostProcessor) processCategory(category string, postSlug string, categories map[string]CategoryInfo) {
	parts := strings.Split(category, "/")
	fullPath := ""

	for i, part := range parts {
		if i > 0 {
			fullPath += "/"
		}
		fullPath += part

		info, exists := categories[fullPath]
		if !exists {
			info = CategoryInfo{
				Name:      part,
				Path:      fullPath,
				PostSlugs: []string{},
				Children:  []string{},
			}
			if i > 0 {
				parentPath := strings.Join(parts[:i], "/")
				info.Parent = parentPath
			}
		}

		if i == len(parts)-1 {
			info.PostSlugs = append(info.PostSlugs, postSlug)
		}
		info.PostCount = len(info.PostSlugs)

		categories[fullPath] = info
	}
}

func (pp *DefaultPostProcessor) buildCategoryHierarchy(categories map[string]CategoryInfo) {
	for path, info := range categories {
		if info.Parent != "" {
			if parent, exists := categories[info.Parent]; exists {
				childExists := false
				for _, child := range parent.Children {
					if child == path {
						childExists = true
						break
					}
				}
				if !childExists {
					parent.Children = append(parent.Children, path)
					categories[info.Parent] = parent
				}
			}
		}
	}
}
