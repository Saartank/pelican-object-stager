package db

import (
	"time"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/logger"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type StagingRecord struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"` // Auto-generated primary key
	CreatedAt      time.Time `gorm:"autoCreateTime"`           // Automatically set the current timestamp
	PelicanURL     string    `gorm:"type:varchar(255)"`        // Pelican URL as a string
	StagingStorage string    `gorm:"type:varchar(255)"`        // Staging storage as a string
	VolumeOccupied float64   `gorm:"type:decimal(10,2)"`       // Volume occupied as a decimal value
	JobID          string    `gorm:"type:varchar(255)"`        // Job ID as a string
}

var (
	DB  *gorm.DB
	log = logger.With(zap.String("component", "database"))
)

// Initialize sets up the database connection and runs migrations.
func Initialize() {
	// Attempt to connect to the database
	var err error
	databaseLocation := config.AppConfig.Database.Location
	DB, err = gorm.Open(sqlite.Open(databaseLocation), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database", zap.Error(err))
		return
	}
	log.Info("Database connection established", zap.String("location", databaseLocation))

	// Run migrations
	err = DB.AutoMigrate(&StagingRecord{})
	if err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
		return
	}
	log.Info("Database migration completed")
}
