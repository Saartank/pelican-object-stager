package server

import (
	"strconv"

	"github.com/pelicanplatform/pelicanobjectstager/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	r := gin.Default()

	// POST /pelican - Invokes the pelican binary
	r.POST("/pelican", handleStartBinary)

	// GET /health - Performs a health check
	r.GET("/health", handleHealthCheck)

	address := config.AppConfig.Server.Port
	logrus.Infof("Starting server on port %d", address)

	port := strconv.Itoa(address)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
