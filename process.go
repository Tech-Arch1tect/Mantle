package main

import (
	"sort"
	"strings"
)

type PostProcessor interface {
	Process(posts []Post) ProcessedPosts
}

type CategoryInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	PostIndices []int    `json:"postIndices"`
	PostCount   int      `json:"postCount"`
	Parent      string   `json:"parent,omitempty"`
	Children    []string `json:"children,omitempty"`
}

type RelatedPost struct {
	Index      int    `json:"index"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	CommonTags int    `json:"commonTags"`
}

type ProcessedPosts struct {
	Posts        []Post                  `json:"posts"`
	Tags         map[string][]int        `json:"tags"`
	Categories   map[string]CategoryInfo `json:"categories"`
	RelatedPosts map[int][]RelatedPost   `json:"relatedPosts"`
}

type DefaultPostProcessor struct{}

func NewPostProcessor() PostProcessor {
	return &DefaultPostProcessor{}
}

func (pp *DefaultPostProcessor) Process(posts []Post) ProcessedPosts {
	processedPosts := ProcessedPosts{
		Posts:        make([]Post, 0, len(posts)),
		Tags:         make(map[string][]int),
		Categories:   make(map[string]CategoryInfo),
		RelatedPosts: make(map[int][]RelatedPost),
	}

	for _, post := range posts {
		processedPosts.Posts = append(processedPosts.Posts, post)

		for _, tag := range post.FrontMatter.Tags {
			processedPosts.Tags[tag] = append(processedPosts.Tags[tag], post.Index)
		}

		if post.FrontMatter.Category != "" {
			pp.processCategory(post.FrontMatter.Category, post.Index, processedPosts.Categories)
		}
	}

	pp.buildCategoryHierarchy(processedPosts.Categories)
	pp.buildRelatedPosts(posts, processedPosts.RelatedPosts)

	return processedPosts
}

func (pp *DefaultPostProcessor) buildRelatedPosts(posts []Post, relatedPosts map[int][]RelatedPost) {
	for i, post := range posts {
		var related []RelatedPost

		for j, otherPost := range posts {
			if i == j {
				continue
			}

			commonTags := pp.countCommonTags(post.FrontMatter.Tags, otherPost.FrontMatter.Tags)
			if commonTags > 0 {
				related = append(related, RelatedPost{
					Index:      otherPost.Index,
					Title:      otherPost.FrontMatter.Title,
					Date:       otherPost.FrontMatter.Date,
					CommonTags: commonTags,
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

		relatedPosts[post.Index] = related
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

func (pp *DefaultPostProcessor) processCategory(category string, postIndex int, categories map[string]CategoryInfo) {
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
				Name:        part,
				Path:        fullPath,
				PostIndices: []int{},
				Children:    []string{},
			}
			if i > 0 {
				parentPath := strings.Join(parts[:i], "/")
				info.Parent = parentPath
			}
		}

		if i == len(parts)-1 {
			info.PostIndices = append(info.PostIndices, postIndex)
		}
		info.PostCount = len(info.PostIndices)

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
