package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type WebServerGenerator struct {
	config *Config
	logger *log.Logger
}

func NewWebServerGenerator(config *Config) *WebServerGenerator {
	return &WebServerGenerator{
		config: config,
		logger: log.New(os.Stdout, "[WebServerGenerator] ", log.LstdFlags),
	}
}

func (wsg *WebServerGenerator) Generate() error {
	wsg.logger.Println("Generating Docker deployment files...")

	outputDir := wsg.config.OutputDir

	if err := wsg.generateDockerfile(outputDir); err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	if err := wsg.generateDockerCompose(outputDir); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	if err := wsg.generateNginxConfigs(outputDir); err != nil {
		return fmt.Errorf("failed to generate nginx configurations: %w", err)
	}

	wsg.logger.Println("Docker deployment files generated successfully")
	return nil
}

func (wsg *WebServerGenerator) generateDockerfile(outputDir string) error {
	dockerfile := `FROM nginx:alpine

COPY public_html /var/www/html
COPY nginx/* /etc/nginx/
COPY nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
`

	dockerfilePath := filepath.Join(outputDir, "Dockerfile")
	return wsg.writeFile(dockerfilePath, dockerfile)
}

func (wsg *WebServerGenerator) generateDockerCompose(outputDir string) error {
	dockerCompose := `services:
  api:
    build: .
    ports:
      - "8080:80"
    restart: unless-stopped
`

	composePath := filepath.Join(outputDir, "docker-compose.yml")
	return wsg.writeFile(composePath, dockerCompose)
}

func (wsg *WebServerGenerator) generateNginxConfigs(outputDir string) error {
	nginxDir := filepath.Join(outputDir, "nginx")

	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx directory: %w", err)
	}

	if err := wsg.generateMainNginxConfig(nginxDir); err != nil {
		return err
	}

	if err := wsg.generateMapsConfig(nginxDir); err != nil {
		return err
	}

	if err := wsg.generateCorsConfig(nginxDir); err != nil {
		return err
	}

	if err := wsg.generateJsonConfig(nginxDir); err != nil {
		return err
	}

	if err := wsg.generateSecurityConfig(nginxDir); err != nil {
		return err
	}

	if err := wsg.generateCompressionConfig(nginxDir); err != nil {
		return err
	}

	return nil
}

func (wsg *WebServerGenerator) generateMainNginxConfig(nginxDir string) error {
	mainConf := `include maps.conf;

server {
    listen 80;
    server_name localhost;
    root /var/www/html;
        
    include json.conf;
    include security.conf;
    include compression.conf;
    
    location / {
        return 404;
    }
    
    location = /api/posts/by-page {
        include cors.conf;
        
        if ($page_param != "") {
            rewrite ^ /api/posts/by-page/$page_param.json last;
        }
        
        try_files /api/posts/by-page/0.json =404;
    }
    
    location = /api/posts/by-slug {
        include cors.conf;
        
        if ($slug_param = "") {
            return 400;
        }
        
        rewrite ^ /api/posts/by-slug/$slug_param.json last;
    }
    
    location = /api/previews/by-page {
        include cors.conf;
        
        if ($page_param != "") {
            rewrite ^ /api/previews/by-page/$page_param.json last;
        }
        
        try_files /api/previews/by-page/0.json =404;
    }
    
    location = /api/previews/by-slug {
        include cors.conf;
        
        if ($slug_param = "") {
            return 400;
        }
        
        rewrite ^ /api/previews/by-slug/$slug_param.json last;
    }
    
    location = /api/tags {
        include cors.conf;
        
        if ($tag_resource != "") {
            rewrite ^ /api/tags/$tag_resource last;
        }
        
        try_files /api/tags/all.json =404;
    }
    
    location = /api/categories {
        include cors.conf;
        
        if ($category_resource != "") {
            rewrite ^ /api/categories/$category_resource last;
        }
        
        try_files /api/categories/all.json =404;
    }
    
    location = /api/related {
        include cors.conf;
        
        if ($related_resource != "") {
            rewrite ^ /api/related/$related_resource last;
        }
        
        try_files /api/related/all.json =404;
    }
    
    location /api/ {
        include cors.conf;
        try_files $uri $uri/ =404;
    }
}
`

	mainConfPath := filepath.Join(nginxDir, "default.conf")
	return wsg.writeFile(mainConfPath, mainConf)
}

func (wsg *WebServerGenerator) generateMapsConfig(nginxDir string) error {
	mapsConf := `# Map query parameters to resource files

# Page parameter mapping - ?page=2 -> 2
map $arg_page $page_param {
    ~^(\d+)$    $1;
    default     "";
}

# Slug parameter mapping - ?slug=my-post -> my-post
map $arg_slug $slug_param {
    ~^([a-z0-9-]+)$    $1;
    default            "";
}

# Tags mapping - ?tag=golang -> golang.json
map $arg_tag $tag_resource {
    ~^(.+)$     $1.json;
    default     "";
}

# Categories mapping - ?category=tech_tutorials -> tech_tutorials.json
map $arg_category $category_resource {
    ~^(.+)$     $1.json;
    default     "";
}

# Related posts mapping - ?slug=my-post -> my-post.json
map $arg_slug $related_resource {
    ~^([^/]+)$  $1.json;
    default     "";
}
`

	mapsConfPath := filepath.Join(nginxDir, "maps.conf")
	return wsg.writeFile(mapsConfPath, mapsConf)
}

func (wsg *WebServerGenerator) generateCorsConfig(nginxDir string) error {
	corsConf := fmt.Sprintf(`add_header Access-Control-Allow-Origin "%s" always;
add_header Access-Control-Allow-Methods "%s" always;
add_header Access-Control-Allow-Headers "%s" always;

if ($request_method = 'OPTIONS') {
    add_header Access-Control-Allow-Origin "%s";
    add_header Access-Control-Allow-Methods "%s";
    add_header Access-Control-Allow-Headers "%s";
    add_header Access-Control-Max-Age %d;
    add_header Content-Length 0;
    add_header Content-Type text/plain;
    return 204;
}
`,
		wsg.config.CorsAllowOrigin,
		wsg.config.CorsAllowMethods,
		wsg.config.CorsAllowHeaders,
		wsg.config.CorsAllowOrigin,
		wsg.config.CorsAllowMethods,
		wsg.config.CorsAllowHeaders,
		wsg.config.CorsMaxAge,
	)

	corsConfPath := filepath.Join(nginxDir, "cors.conf")
	return wsg.writeFile(corsConfPath, corsConf)
}

func (wsg *WebServerGenerator) generateJsonConfig(nginxDir string) error {
	jsonConf := `location ~ \.json$ {
    include cors.conf;
    add_header Content-Type application/json;
    add_header Cache-Control "public, max-age=300";
}
`

	jsonConfPath := filepath.Join(nginxDir, "json.conf")
	return wsg.writeFile(jsonConfPath, jsonConf)
}

func (wsg *WebServerGenerator) generateSecurityConfig(nginxDir string) error {
	securityConf := `add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;

server_tokens off;

autoindex off;
`

	securityConfPath := filepath.Join(nginxDir, "security.conf")
	return wsg.writeFile(securityConfPath, securityConf)
}

func (wsg *WebServerGenerator) generateCompressionConfig(nginxDir string) error {
	compressionConf := `gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss;
gzip_disable "MSIE [1-6]\.";
`

	compressionConfPath := filepath.Join(nginxDir, "compression.conf")
	return wsg.writeFile(compressionConfPath, compressionConf)
}

func (wsg *WebServerGenerator) writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	wsg.logger.Printf("Generated: %s", filepath.Base(path))
	return nil
}
