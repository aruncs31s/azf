package routes

import (
	"path/filepath"

	"github.com/aruncs31s/azf/application/handler"
	"github.com/gin-gonic/gin"
)

// SetupDocsRoutes registers all documentation routes with the Gin router.
// It provides access to markdown documentation files stored in the specified directory.
//
// Usage:
//
//	routes.SetupDocsRoutes(router, "docs")
//	// Now available at GET /docs and GET /api/docs/list
//
// Parameters:
//   - router: The Gin engine instance
//   - docsPath: Path to the documentation directory (e.g., "docs")
//
// Routes registered:
//   - GET /docs - Display documentation viewer with sidebar navigation
//   - GET /api/docs/list - Get JSON list of all documentation files
//
// Query parameters for /docs:
//   - path: Optional path to specific document (e.g., ?path=guides/authentication)
//
// Example:
//
//	func main() {
//	    router := gin.Default()
//	    router.LoadHTMLGlob("application/templates/*.html")
//
//	    // Setup documentation routes
//	    routes.SetupDocsRoutes(router, "docs")
//
//	    router.Run(":8080")
//	}
func SetupDocsRoutes(router *gin.Engine, docsPath string) {
	if docsPath == "" {
		docsPath = filepath.Join("docs")
	}

	docsHandler := handler.NewDocsHandler(docsPath)

	// GET /docs - Display the documentation page
	// Shows a specific document or README by default
	// Query parameters:
	//   - path: Document path without .md extension (e.g., ?path=api/endpoints)
	router.GET("/docs", func(c *gin.Context) {
		docsHandler.GetDocsPage(c)
	})

	// GET /api/docs/list - Get documentation structure as JSON
	// Returns the folder structure and all available documents
	router.GET("/api/docs/list", func(c *gin.Context) {
		docsHandler.GetDocsList(c)
	})
}

// SetupDocsRoutesWithMiddleware registers documentation routes with custom middleware.
// Use this if you want to add authentication, logging, or other middleware to docs routes.
//
// Usage:
//
//	routes.SetupDocsRoutesWithMiddleware(
//	    router,
//	    "docs",
//	    authMiddleware,
//	    loggingMiddleware,
//	)
//
// Parameters:
//   - router: The Gin engine instance
//   - docsPath: Path to the documentation directory
//   - middleware: Variable number of Gin middleware handlers to apply
//
// Example with authentication:
//
//	func authMiddleware() gin.HandlerFunc {
//	    return func(c *gin.Context) {
//	        token := c.GetHeader("Authorization")
//	        if token == "" {
//	            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
//	            c.Abort()
//	            return
//	        }
//	        c.Next()
//	    }
//	}
//
//	routes.SetupDocsRoutesWithMiddleware(router, "docs", authMiddleware())
func SetupDocsRoutesWithMiddleware(router *gin.Engine, docsPath string, middleware ...gin.HandlerFunc) {
	if docsPath == "" {
		docsPath = filepath.Join("docs")
	}

	docsHandler := handler.NewDocsHandler(docsPath)

	// Apply middleware to docs routes
	docsGroup := router.Group("/docs")
	docsGroup.Use(middleware...)
	{
		docsGroup.GET("", func(c *gin.Context) {
			docsHandler.GetDocsPage(c)
		})
	}

	// Apply middleware to API docs routes
	apiDocsGroup := router.Group("/api/docs")
	apiDocsGroup.Use(middleware...)
	{
		apiDocsGroup.GET("/list", func(c *gin.Context) {
			docsHandler.GetDocsList(c)
		})
	}
}

// GetDefaultDocsPath returns the default path to the documentation directory.
// This is typically "docs" in the project root.
//
// Usage:
//
//	docsPath := routes.GetDefaultDocsPath()
//	routes.SetupDocsRoutes(router, docsPath)
func GetDefaultDocsPath() string {
	return filepath.Join("docs")
}

// SetupDocsRoutesWithPrefix registers documentation routes under a custom URL prefix.
// Useful if you want docs under a different path like /help or /knowledge-base.
//
// Usage:
//
//	routes.SetupDocsRoutesWithPrefix(router, "docs", "/help")
//	// Routes will be: GET /help and GET /api/help/list
//
// Parameters:
//   - router: The Gin engine instance
//   - docsPath: Path to the documentation directory on disk
//   - prefix: URL prefix for the routes (e.g., "/help", "/kb", "/documentation")
//
// Example:
//
//	func main() {
//	    router := gin.Default()
//
//	    // Docs available at /help instead of /docs
//	    routes.SetupDocsRoutesWithPrefix(router, "docs", "/help")
//
//	    // Routes are now:
//	    // GET /help - View docs
//	    // GET /api/help/list - Get docs list
//
//	    router.Run(":8080")
//	}
func SetupDocsRoutesWithPrefix(router *gin.Engine, docsPath string, prefix string) {
	if docsPath == "" {
		docsPath = filepath.Join("docs")
	}

	if prefix == "" {
		prefix = "/docs"
	}

	docsHandler := handler.NewDocsHandler(docsPath)

	// Main docs page
	router.GET(prefix, func(c *gin.Context) {
		docsHandler.GetDocsPage(c)
	})

	// API endpoint
	apiPrefix := prefix
	if prefix == "/docs" {
		apiPrefix = "/api/docs/list"
	} else {
		apiPrefix = "/api" + prefix + "/list"
	}

	router.GET(apiPrefix, func(c *gin.Context) {
		docsHandler.GetDocsList(c)
	})
}
