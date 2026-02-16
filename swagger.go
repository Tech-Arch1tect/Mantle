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
// @Produce json
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]Post,pagination=PaginationInfo} "Paginated posts"
// @Failure 404 {object} ErrorResponse "Page not found"
// @Router /posts/by-page [get]
func GetPostsByPage() {}

// @Summary Get post by slug
// @Description Get a specific post by its slug
// @Tags posts
// @Produce json
// @Param slug query string true "Post slug"
// @Success 200 {object} SingleResponse{data=Post} "Single post"
// @Failure 400 {object} ErrorResponse "Missing slug parameter"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Router /posts/by-slug [get]
func GetPostBySlug() {}

// @Summary Get paginated previews
// @Description Get paginated post previews with optional page parameter
// @Tags previews
// @Produce json
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]PostPreview,pagination=PaginationInfo} "Paginated previews"
// @Failure 404 {object} ErrorResponse "Page not found"
// @Router /previews/by-page [get]
func GetPreviewsByPage() {}

// @Summary Get preview by slug
// @Description Get a specific post preview by its slug
// @Tags previews
// @Produce json
// @Param slug query string true "Post slug"
// @Success 200 {object} SingleResponse{data=PostPreview} "Single preview"
// @Failure 400 {object} ErrorResponse "Missing slug parameter"
// @Failure 404 {object} ErrorResponse "Preview not found"
// @Router /previews/by-slug [get]
func GetPreviewBySlug() {}

// @Summary Get paginated tag list
// @Description Get all tags with post counts, paginated
// @Tags tags
// @Produce json
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]TagInfo,pagination=PaginationInfo} "Paginated tag list"
// @Router /tags [get]
func GetTagList() {}

// @Summary Get posts by tag
// @Description Get paginated post previews for a specific tag
// @Tags tags
// @Produce json
// @Param tag query string true "Tag name"
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]PostPreview,pagination=PaginationInfo} "Paginated previews for tag"
// @Failure 404 {object} ErrorResponse "Tag not found"
// @Router /tags [get]
func GetPostsByTag() {}

// @Summary Get paginated category list
// @Description Get all categories with post counts, paginated
// @Tags categories
// @Produce json
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]CategoryInfo,pagination=PaginationInfo} "Paginated category list"
// @Router /categories [get]
func GetCategoryList() {}

// @Summary Get posts by category
// @Description Get paginated post previews for a specific category
// @Tags categories
// @Produce json
// @Param category query string true "Category path (e.g., tutorials/go)"
// @Param page query int false "Page number (0-indexed)" default(0)
// @Success 200 {object} PaginatedResponse{data=[]PostPreview,pagination=PaginationInfo} "Paginated previews for category"
// @Failure 404 {object} ErrorResponse "Category not found"
// @Router /categories [get]
func GetPostsByCategory() {}

// @Summary Get category tree
// @Description Get hierarchical category tree structure
// @Tags categories
// @Produce json
// @Success 200 {object} SingleResponse{data=CategoryTree} "Hierarchical category tree"
// @Router /categories/tree.json [get]
func GetCategoryTree() {}

// @Summary Get related posts
// @Description Get related posts for a specific post by slug
// @Tags related
// @Produce json
// @Param slug query string true "Post slug"
// @Success 200 {object} PaginatedResponse{data=[]RelatedPost,pagination=PaginationInfo} "Related posts"
// @Failure 400 {object} ErrorResponse "Missing slug parameter"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Router /related [get]
func GetRelated() {}

// @Summary Get search index
// @Description Get inverted search index for client-side search
// @Tags search
// @Produce json
// @Success 200 {object} SingleResponse{data=SearchIndex} "Search index"
// @Router /search/index.json [get]
func GetSearchIndex() {}

// @Summary Get API metadata
// @Description Get unified API metadata including counts, pagination info, and configuration
// @Tags metadata
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
