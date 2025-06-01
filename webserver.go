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
	mainConf := `server {
    listen 80;
    server_name localhost;
    root /var/www/html;
        
    include json.conf;
    include security.conf;
    include compression.conf;
    
    location / {
        return 404;
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

func (wsg *WebServerGenerator) generateCorsConfig(nginxDir string) error {
	corsConf := `add_header Access-Control-Allow-Origin "*" always;
add_header Access-Control-Allow-Methods "GET, OPTIONS" always;
add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept" always;

if ($request_method = 'OPTIONS') {
    add_header Access-Control-Allow-Origin "*";
    add_header Access-Control-Allow-Methods "GET, OPTIONS";
    add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept";
    add_header Access-Control-Max-Age 86400;
    add_header Content-Length 0;
    add_header Content-Type text/plain;
    return 204;
}
`

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
