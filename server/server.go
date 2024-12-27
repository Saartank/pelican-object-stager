package server

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/dbrefresh"
	"github.com/pelicanplatform/pelicanobjectstager/logger"
	"github.com/pelicanplatform/pelicanobjectstager/server/object"
)

var log = logger.With(zap.String("component", "server"))

func StartServer() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := gin.New()

	for _, mw := range middlewares {
		r.Use(mw)
	}
	r.POST("/pelican", handleStartBinary)
	r.GET("/health", handleHealthCheck)
	object.RegisterObjectRoutes(r)

	address := config.AppConfig.Server.Port

	log.Debug("Starting LaunchPeriodicRefreshRecords...")
	go dbrefresh.LaunchPeriodicRefreshRecords(ctx)

	log.Info("Starting server",
		zap.Int("port", address),
	)

	port := strconv.Itoa(address)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server",
			zap.Error(err),
		)
	}
}
