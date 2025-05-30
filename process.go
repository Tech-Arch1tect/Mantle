package main

import "strings"

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

type ProcessedPosts struct {
	Posts      []Post                  `json:"posts"`
	Tags       map[string][]int        `json:"tags"`
	Categories map[string]CategoryInfo `json:"categories"`
}

type DefaultPostProcessor struct{}

func NewPostProcessor() PostProcessor {
	return &DefaultPostProcessor{}
}

func (pp *DefaultPostProcessor) Process(posts []Post) ProcessedPosts {
	processedPosts := ProcessedPosts{
		Posts:      make([]Post, 0, len(posts)),
		Tags:       make(map[string][]int),
		Categories: make(map[string]CategoryInfo),
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

	return processedPosts
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
