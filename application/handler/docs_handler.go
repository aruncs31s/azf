package handler

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aruncs31s/azf/application/dto"
	"github.com/aruncs31s/azf/application/templates"
	"github.com/gin-gonic/gin"
)

// DocPage represents a documentation page with content
type DocPage struct {
	Title       string
	Content     string
	Path        string
	NavTree     []dto.DocItem
	BreadCrumbs []BreadCrumb
}

// BreadCrumb represents a navigation breadcrumb
type BreadCrumb struct {
	Name string
	Path string
}

// DocsHandler handles documentation page rendering
type DocsHandler struct {
	docsPath string
}

// NewDocsHandler creates a new docs handler
func NewDocsHandler(docsPath string) *DocsHandler {
	return &DocsHandler{
		docsPath: docsPath,
	}
}

// GetDocsPage retrieves and renders a documentation page
func (h *DocsHandler) GetDocsPage(c *gin.Context) {
	docPath := c.Query("path")
	if docPath == "" {
		docPath = "README"
	}

	// Sanitize path to prevent directory traversal
	docPath = filepath.Clean(docPath)
	if strings.Contains(docPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	// Ensure .md extension
	if !strings.HasSuffix(docPath, ".md") {
		docPath = docPath + ".md"
	}

	fullPath := filepath.Join(h.docsPath, docPath)

	// Verify file is within docs directory
	absDocsPath, _ := filepath.Abs(h.docsPath)
	absFilePath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFilePath, absDocsPath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid path"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "documentation not found"})
		return
	}

	// Read the markdown file
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read documentation"})
		return
	}

	// Build navigation tree
	navTree := h.buildNavTree(h.docsPath, "")

	// Extract title from content (first heading)
	title := h.extractTitle(string(content), docPath)

	// Get current path for active state
	currentPath := docPath

	// Render using Templ component
	component := templates.DocsViewer(
		title,
		string(content),
		navTree,
		currentPath,
	)

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(context.Background(), c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render template"})
		return
	}
}

// GetDocsList returns the full documentation tree structure
func (h *DocsHandler) GetDocsList(c *gin.Context) {
	navTree := h.buildNavTree(h.docsPath, "")
	c.JSON(http.StatusOK, gin.H{"docs": navTree})
}

// buildNavTree recursively builds the navigation tree
func (h *DocsHandler) buildNavTree(basePath string, prefix string) []dto.DocItem {
	var items []dto.DocItem

	entries, err := ioutil.ReadDir(basePath)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		var relPath string
		if prefix == "" {
			relPath = entry.Name()
		} else {
			relPath = filepath.Join(prefix, entry.Name())
		}

		fullPath := filepath.Join(basePath, entry.Name())

		item := dto.DocItem{
			Name:     entry.Name(),
			Path:     relPath,
			IsFolder: entry.IsDir(),
		}

		if entry.IsDir() {
			item.Children = h.buildNavTree(fullPath, relPath)
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension from path for display
			item.Path = strings.TrimSuffix(relPath, ".md")
		} else {
			// Skip non-markdown files
			continue
		}

		items = append(items, item)
	}

	// Sort items - directories first, then files
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsFolder != items[j].IsFolder {
			return items[i].IsFolder
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items
}

// buildBreadcrumbs creates breadcrumb navigation from a path
func (h *DocsHandler) buildBreadcrumbs(docPath string) []BreadCrumb {
	var breadcrumbs []BreadCrumb

	// Add root
	breadcrumbs = append(breadcrumbs, BreadCrumb{
		Name: "Docs",
		Path: "README",
	})

	// Parse path components
	parts := strings.Split(strings.TrimSuffix(docPath, ".md"), "/")
	currentPath := ""

	for i, part := range parts {
		if part == "" {
			continue
		}

		if i == 0 {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		breadcrumbs = append(breadcrumbs, BreadCrumb{
			Name: part,
			Path: currentPath,
		})
	}

	return breadcrumbs
}

// extractTitle extracts the title from markdown content
func (h *DocsHandler) extractTitle(content string, filePath string) string {
	lines := strings.Split(content, "\n")

	// Look for first heading
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
	}

	// Fall back to filename
	return strings.TrimSuffix(filepath.Base(filePath), ".md")
}
