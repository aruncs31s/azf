package helper

import (
	"net/http"

	"github.com/aruncs31s/responsehelper"
	"github.com/gin-gonic/gin"
)

func NewResponseHelper() responsehelper.ResponseHelper {
	return responsehelper.NewResponseHelper()
}

// SendBadRequest sends a standardized bad request response
func SendBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}
