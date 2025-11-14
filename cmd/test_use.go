package main

import (
	AZFauthzframework "github.com/aruncs31s/azf"
	"github.com/aruncs31s/azf/application/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// var DB *gorm.DB

// func init() {
// 	db := db.InitDB()
// 	DB = db
// 	user1 := &model.User{
// 		Username: "aruncs31s",
// 		Password: "12345678",
// 		Email:    "aruncs31ss@gmail.com",
// 	}
// 	user2 := &model.User{
// 		Username: "simple",
// 		Password: "12345678",
// 		Email:    "simple@example.com",
// 	}
// 	db.Create(user1)
// 	db.Create(user2)
// }

func main() {
	r := gin.Default()
	godotenv.Load()
	// Public route (no auth required)
	r.GET("/hi", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hi",
		})
	})
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "aruncs")
		c.Set("role", "admin")
		c.Next()
	})

	AZFauthzframework.InitAuthZModule(nil, nil)
	// Usage Tracking
	AZFauthzframework.InitUsageTracking()
	AZFauthzframework.SetApiTrackingMiddleware(r)

	// Documents
	routes.SetupDocsRoutes(r, "docs")

	// Ui Routes

	AZFauthzframework.SetupUI(r)
	// AuthZMiddleware
	AZFauthzframework.SetAuthZMiddleware(r)

	// Create sample audit logs for testing
	// createSampleAuditLogs()

	// Protected routes that will trigger authorization checks and generate audit logs
	r.GET("/api/admin/users", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Admin users endpoint",
			"users":   []string{"alice", "bob"},
		})
	})

	r.Run()
}
