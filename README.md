# Mantle

**Mantle** is a static API generator for blogs that transforms markdown files with frontmatter into a JSON API, complete with pagination, search (client side), categorisation, and related content features.

## Features

- **Markdown Processing**: Converts markdown files with YAML frontmatter into structured JSON data
- **Automatic API Generation**: Creates RESTful JSON endpoints for posts, previews, tags, categories, and search
- **Pagination Support**: Built-in pagination for posts and previews with configurable page sizes
- **Hierarchical Categories**: Support for nested categories (e.g., `tech/tutorials/golang`)
- **Related Posts**: Automatically generates related post suggestions based on common tags
- **Search Index**: Creates an inverted index for fast content searching (client side)
- **Excerpt Generation**: Automatic excerpt generation with support for custom excerpts and `<!--more-->` tags
- **Docker Ready**: Generates complete Docker deployment with nginx configuration

## Installation

TODO

## Configuration

Configure Mantle using environment variables:

| Environment Variable | Default Value | Description                            |
| -------------------- | ------------- | -------------------------------------- |
| `CONTENT_DIR`        | `./content`   | Directory containing markdown files    |
| `OUTPUT_DIR`         | `./output`    | Directory for generated files          |
| `POSTS_PER_PAGE`     | `10`          | Number of posts per pagination page    |
| `PREVIEWS_PER_PAGE`  | `10`          | Number of previews per pagination page |
| `DATE_FORMAT`        | `2006-01-02`  | Go date format for parsing dates       |
| `CORS_ALLOW_ORIGIN`  | `*`           | CORS allowed origins                   |

## Usage

### 1. Prepare Content

Create markdown files in your content directory with YAML frontmatter:

```markdown
---
title: "Getting Started with Go"
author: "John Doe"
date: "2024-01-15"
tags: ["golang", "tutorial", "beginner"]
category: "tech/tutorials"
excerpt: "Learn the basics of Go programming language"
---

# Getting Started with Go

This is the content of your post...

<!--more-->

Additional content that appears after the excerpt...
```

### 2. Generate API

```bash
# Using default configuration
./mantle

# Using custom configuration
CONTENT_DIR=/path/to/markdown OUTPUT_DIR=/path/to/output ./mantle
```

### 3. Deploy

The generated output includes Docker deployment files:

```bash
cd output
docker-compose up -d --build
```

Your API will be available at `http://localhost:8080/api/`

## API Endpoints

### Posts

- `GET /api/posts/by-page` - Paginated posts (default: page 0)
- `GET /api/posts/by-page?page=1` - Specific page of posts
- `GET /api/posts/by-slug?slug=my-post` - Individual post by slug

### Previews

- `GET /api/previews/by-page` - Paginated previews (default: page 0)
- `GET /api/previews/by-page?page=1` - Specific page of previews
- `GET /api/previews/by-slug?slug=my-post` - Individual preview by slug

### Tags

- `GET /api/tags` - All tags with post indices
- `GET /api/tags?tag=golang` - Posts for specific tag

### Categories

- `GET /api/categories` - All categories
- `GET /api/categories?category=tech_tutorials` - Specific category
- `GET /api/categories/tree.json` - Hierarchical category tree

### Related Posts

- `GET /api/related?id=1` - Related posts for specific post

### Search

- `GET /api/search/inverted.json` - Search index for client-side search

### Metadata

- `GET /api/meta.json` - Unified API metadata

## Frontmatter Schema

| Field      | Type   | Required | Description                                     |
| ---------- | ------ | -------- | ----------------------------------------------- |
| `title`    | string | Yes      | Post title                                      |
| `author`   | string | Yes      | Post author                                     |
| `date`     | string | Yes      | Publication date (must match `DATE_FORMAT`)     |
| `tags`     | array  | No       | Array of tags                                   |
| `category` | string | No       | Hierarchical category (e.g., "tech/tutorials")  |
| `excerpt`  | string | No       | Custom excerpt (auto-generated if not provided) |
