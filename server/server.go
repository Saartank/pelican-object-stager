package server

import (
	"strconv"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/server/object"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	r := gin.Default()

	r.POST("/pelican", handleStartBinary)

	r.GET("/health", handleHealthCheck)

	object.RegisterObjectRoutes(r)

	address := config.AppConfig.Server.Port
	logrus.Infof("Starting server on port %d", address)

	port := strconv.Itoa(address)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
