package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
)

// @title           Mantle API
// @version         1.0
// @description     A static API for blogs that transforms markdown files with frontmatter into JSON endpoints

// @host      localhost:8080
// @BasePath  /api

// @Summary Get paginated posts
// @Description Get paginated posts with optional page parameter
// @Tags posts
// @Accept json
// @Produce json
// @Param page query int false "Page number (0-indexed)"
// @Success 200 {object} PostsResponse "Paginated posts"
// @Failure 404 {object} ErrorResponse "Page not found"
// @Router /posts/by-page [get]
func GetPostsByPage() {}

// @Summary Get post by slug
// @Description Get a specific post by its slug
// @Tags posts
// @Accept json
// @Produce json
// @Param slug query string true "Post slug"
// @Success 200 {object} Post "Single post"
// @Failure 400 {object} ErrorResponse "Missing slug parameter"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Router /posts/by-slug [get]
func GetPostBySlug() {}

// @Summary Get paginated previews
// @Description Get paginated post previews with optional page parameter
// @Tags previews
// @Accept json
// @Produce json
// @Param page query int false "Page number (0-indexed)"
// @Success 200 {object} PreviewsResponse "Paginated previews"
// @Failure 404 {object} ErrorResponse "Page not found"
// @Router /previews/by-page [get]
func GetPreviewsByPage() {}

// @Summary Get preview by slug
// @Description Get a specific post preview by its slug
// @Tags previews
// @Accept json
// @Produce json
// @Param slug query string true "Post slug"
// @Success 200 {object} PostPreview "Single preview"
// @Failure 400 {object} ErrorResponse "Missing slug parameter"
// @Failure 404 {object} ErrorResponse "Preview not found"
// @Router /previews/by-slug [get]
func GetPreviewBySlug() {}

// @Summary Get all tags
// @Description Get all tags or filter posts by specific tag
// @Tags tags
// @Accept json
// @Produce json
// @Param tag query string false "Tag name"
// @Success 200 {object} TagsMap "All tags with associated post slugs"
// @Success 200 {array} string "Post slugs for specific tag when tag provided"
// @Failure 404 {object} ErrorResponse "Tag not found"
// @Router /tags [get]
func GetTags() {}

// @Summary Get all categories
// @Description Get all categories or filter by specific category
// @Tags categories
// @Accept json
// @Produce json
// @Param category query string false "Category path (e.g., tech_tutorials)"
// @Success 200 {object} CategoriesMap "All categories"
// @Success 200 {object} CategoryDetail "Category with posts when category provided"
// @Failure 404 {object} ErrorResponse "Category not found"
// @Router /categories [get]
func GetCategories() {}

// @Summary Get category tree
// @Description Get hierarchical category tree structure
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} CategoryTree "Hierarchical category tree"
// @Router /categories/tree.json [get]
func GetCategoryTree() {}

// @Summary Get related posts
// @Description Get related posts for all posts or for a specific post
// @Tags related
// @Accept json
// @Produce json
// @Param slug query string false "Post slug"
// @Success 200 {object} RelatedPostsMap "All related posts mapping"
// @Success 200 {array} RelatedPost "Related posts for specific post when slug provided"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Router /related [get]
func GetRelated() {}

// @Summary Get search index
// @Description Get inverted search index for client-side search
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} SearchIndex "Inverted search index mapping terms to post slugs"
// @Router /search/inverted.json [get]
func GetSearchIndex() {}

// @Summary Get API metadata
// @Description Get unified API metadata including counts, pagination info, and configuration
// @Tags metadata
// @Accept json
// @Produce json
// @Success 200 {object} MetadataResponse "API metadata"
// @Router /meta.json [get]
func GetMetadata() {}

type SwaggerGenerator struct {
	config *Config
	logger *log.Logger
}

func NewSwaggerGenerator(config *Config) *SwaggerGenerator {
	return &SwaggerGenerator{
		config: config,
		logger: log.New(os.Stdout, "[SwaggerGenerator] ", log.LstdFlags),
	}
}

func (sg *SwaggerGenerator) Generate() error {
	sg.logger.Println("Generating OpenAPI specification...")

	config := &gen.Config{
		SearchDir:          ".",
		Excludes:           "",
		MainAPIFile:        "swagger.go",
		PropNamingStrategy: swag.CamelCase,
		OutputDir:          filepath.Join(sg.config.OutputDir, "public_html", "api"),
		OutputTypes:        []string{"json", "yaml"},
		ParseVendor:        false,
		ParseDependency:    0,
		MarkdownFilesDir:   "",
		ParseInternal:      false,
		GeneratedTime:      true,
		RequiredByDefault:  false,
		ParseDepth:         100,
		InstanceName:       "",
	}

	if err := gen.New().Build(config); err != nil {
		return err
	}

	sg.logger.Println("OpenAPI specification generated successfully")
	return nil
}
