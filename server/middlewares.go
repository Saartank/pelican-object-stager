package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pelicanplatform/pelicanobjectstager/logger"
	"go.uber.org/zap"
)

var middlewares = []gin.HandlerFunc{
	JobIDMiddleware(),
	GinLoggerMiddleware(),
	GinRecoveryLoggerMiddleware(),
}

// JobIDMiddleware generates a unique Job ID for each request
func JobIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a UUID
		jobID := uuid.New().String()

		// Store the Job ID in the context
		c.Set("job_id", jobID)

		// Add the Job ID to the response header for debugging
		c.Writer.Header().Set("X-Job-ID", jobID)

		// Proceed with the request
		c.Next()
	}
}

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		jobID, _ := c.Get("job_id")

		logger.Base().Info("Request handled",
			zap.String("job_id", jobID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", statusCode),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func GinRecoveryLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, _ := c.Get("job_id")
		defer func() {
			if err := recover(); err != nil {
				logger.Base().Error("Panic recovered",
					zap.String("job_id", jobID.(string)),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
