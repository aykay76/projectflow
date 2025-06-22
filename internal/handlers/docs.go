package handlers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aykay76/projectflow/internal/logger"
)

// DocumentInfo represents metadata about a documentation file
type DocumentInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	ModifiedAt  time.Time `json:"modified_at"`
	Size        int64     `json:"size"`
}

// DocumentContent represents the full content of a documentation file
type DocumentContent struct {
	DocumentInfo
	Content string `json:"content"`
}

// DocsHandler handles documentation API endpoints
type DocsHandler struct {
	docsPath string
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(docsPath string) *DocsHandler {
	return &DocsHandler{
		docsPath: docsPath,
	}
}

// HandleDocsList handles GET /api/docs/list
func (d *DocsHandler) HandleDocsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logger.Info("Listing documentation files", "path", d.docsPath)

	docs, err := d.listDocuments()
	if err != nil {
		logger.Error("Failed to list documentation files", "error", err)
		http.Error(w, fmt.Sprintf("Failed to list documentation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if err := json.NewEncoder(w).Encode(docs); err != nil {
		logger.Error("Failed to encode documentation list", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully listed documentation files", "count", len(docs))
}

// HandleDocsGet handles GET /api/docs/{filename}
func (d *DocsHandler) HandleDocsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from path
	path := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	if path == "" || path == "/" {
		http.Error(w, "Document name required", http.StatusBadRequest)
		return
	}

	// Clean and validate the path
	path = strings.TrimSuffix(path, "/")
	if !isValidDocPath(path) {
		logger.Warn("Invalid document path requested", "path", path)
		http.Error(w, "Invalid document path", http.StatusBadRequest)
		return
	}

	logger.Info("Fetching documentation file", "path", path)

	// Get document content
	doc, err := d.getDocument(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Documentation file not found", "path", path)
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to read documentation file", "path", path, "error", err)
		http.Error(w, fmt.Sprintf("Failed to read document: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if err := json.NewEncoder(w).Encode(doc); err != nil {
		logger.Error("Failed to encode documentation content", "path", path, "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully served documentation file", "path", path, "size", len(doc.Content))
}

// listDocuments scans the docs directory and returns a list of available documents
func (d *DocsHandler) listDocuments() ([]DocumentInfo, error) {
	var docs []DocumentInfo

	err := filepath.WalkDir(d.docsPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		// Get relative path from docs directory
		relPath, err := filepath.Rel(d.docsPath, path)
		if err != nil {
			return err
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			return err
		}

		// Extract metadata from file
		title, description, category := d.extractMetadata(path)

		// Use filename without extension as name if no title found
		name := strings.TrimSuffix(entry.Name(), ".md")
		if title == "" {
			title = formatTitle(name)
		}

		doc := DocumentInfo{
			Name:        name,
			Path:        relPath,
			Title:       title,
			Description: description,
			Category:    category,
			ModifiedAt:  info.ModTime(),
			Size:        info.Size(),
		}

		docs = append(docs, doc)
		return nil
	})

	return docs, err
}

// getDocument reads and returns the content of a specific document
func (d *DocsHandler) getDocument(docPath string) (*DocumentContent, error) {
	// Construct full file path
	fullPath := filepath.Join(d.docsPath, docPath)

	// Ensure .md extension
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath += ".md"
	}

	// Check if file exists and is within docs directory
	if !d.isValidFilePath(fullPath) {
		return nil, os.ErrNotExist
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// Extract metadata
	title, description, category := d.extractMetadata(fullPath)

	// Use filename without extension as name if no title found
	name := strings.TrimSuffix(filepath.Base(fullPath), ".md")
	if title == "" {
		title = formatTitle(name)
	}

	// Get relative path
	relPath, err := filepath.Rel(d.docsPath, fullPath)
	if err != nil {
		relPath = docPath
	}

	doc := &DocumentContent{
		DocumentInfo: DocumentInfo{
			Name:        name,
			Path:        relPath,
			Title:       title,
			Description: description,
			Category:    category,
			ModifiedAt:  info.ModTime(),
			Size:        info.Size(),
		},
		Content: string(content),
	}

	return doc, nil
}

// extractMetadata extracts title, description, and category from markdown frontmatter or content
func (d *DocsHandler) extractMetadata(filePath string) (title, description, category string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", ""
	}

	contentStr := string(content)

	// Try to extract title from markdown headers
	titleRegex := regexp.MustCompile(`^#\s+(.+)$`)
	if matches := titleRegex.FindStringSubmatch(contentStr); len(matches) > 1 {
		title = strings.TrimSpace(matches[1])
	}

	// Try to extract description from first paragraph
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
			// Take first non-header, non-empty line as description
			if len(line) > 100 {
				description = line[:97] + "..."
			} else {
				description = line
			}
			break
		}
	}

	// Determine category based on filename
	filename := strings.ToLower(filepath.Base(filePath))
	switch {
	case strings.Contains(filename, "user"):
		category = "User Guide"
	case strings.Contains(filename, "developer"):
		category = "Developer Guide"
	case strings.Contains(filename, "api") || strings.Contains(filename, "mcp"):
		category = "API Reference"
	case strings.Contains(filename, "config") || strings.Contains(filename, "deploy"):
		category = "Configuration"
	case strings.Contains(filename, "troubleshoot") || strings.Contains(filename, "faq"):
		category = "Support"
	default:
		category = "General"
	}

	return title, description, category
}

// isValidFilePath checks if the file path is valid and within the docs directory
func (d *DocsHandler) isValidFilePath(filePath string) bool {
	// Get absolute paths for comparison
	absDocsPath, err := filepath.Abs(d.docsPath)
	if err != nil {
		return false
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	// Check if file is within docs directory
	relPath, err := filepath.Rel(absDocsPath, absFilePath)
	if err != nil {
		return false
	}

	// Ensure the path doesn't escape the docs directory
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	return true
}

// isValidDocPath validates document path for API access
func isValidDocPath(path string) bool {
	// Check for invalid characters and patterns
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return false
	}

	// Must be alphanumeric, hyphens, underscores, and forward slashes only
	validPath := regexp.MustCompile(`^[a-zA-Z0-9_/-]+$`)
	return validPath.MatchString(path)
}

// formatTitle converts a filename to a human-readable title
func formatTitle(name string) string {
	// Replace hyphens and underscores with spaces
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize each word
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
