package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns a simple liveness probe.
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "finbud-backend",
	})
}

