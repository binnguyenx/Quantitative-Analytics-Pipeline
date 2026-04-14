// Package middleware contains reusable Gin middleware.
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/finbud/finbud-backend/internal/metrics"
)

// RequestLogger logs every incoming request with its latency.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		route := c.FullPath()
		if route == "" {
			route = path
		}
		metrics.ObserveAPIRequest(c.Request.Method, route, status, float64(latency.Milliseconds()))

		log.Printf("[%s] %d %s %s (%v)",
			c.Request.Method, status, path, c.ClientIP(), latency,
		)
	}
}

// CORS sets permissive CORS headers for development.
// In production you should restrict AllowOrigins.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
