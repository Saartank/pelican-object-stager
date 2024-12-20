package db

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/logger"
)

type StagingRecord struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"` // Auto-generated primary key
	CreatedAt       time.Time `gorm:"autoCreateTime"`           // Automatically set the current timestamp
	PelicanURL      string    `gorm:"type:varchar(255)"`        // Pelican URL as a string
	StagingStorage  string    `gorm:"type:varchar(255)"`        // Staging storage as a string
	ObjectSize      int64     `gorm:"type:bigint"`              // Object size in bytes as an int64
	JobID           string    `gorm:"type:varchar(255)"`        // Job ID as a string
	PelicanExitCode int       `gorm:"type:int"`                 // Pelican client exit code as an integer
	PelicanStdout   string    `gorm:"type:text"`                // Pelican client stdout as a string
	PelicanStderr   string    `gorm:"type:text"`                // Pelican client stderr as a string
}

var (
	db  *gorm.DB
	log = logger.With(zap.String("component", "database"))
)

// Initialize sets up the database connection and runs migrations.
func InitializeDB() {
	// Attempt to connect to the database
	var err error
	databaseLocation := config.AppConfig.Database.Location
	db, err = gorm.Open(sqlite.Open(databaseLocation), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database", zap.Error(err))
		return
	}
	log.Info("Database connection established", zap.String("location", databaseLocation))

	// Run migrations
	err = db.AutoMigrate(&StagingRecord{})
	if err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
		return
	}
	log.Info("Database migration completed")
}

func InsertStagingRecord(pelicanURL, stagingStorage, jobID string, objectSize int64, exitCode int, stdout, stderr string) error {
	// Create a new record instance
	record := StagingRecord{
		PelicanURL:      pelicanURL,
		StagingStorage:  stagingStorage,
		ObjectSize:      objectSize,
		JobID:           jobID,
		PelicanExitCode: exitCode,
		PelicanStdout:   stdout,
		PelicanStderr:   stderr,
	}

	// Insert the record into the database
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to add record: %v", err)
	}

	return nil
}
