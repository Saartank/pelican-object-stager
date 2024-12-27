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
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`                                    // Automatically set the current timestamp
	PelicanURL      string    `gorm:"type:varchar(255);uniqueIndex:idx_pelican_staging"` // Part of unique combination
	StagingStorage  string    `gorm:"type:varchar(255);uniqueIndex:idx_pelican_staging"` // Part of unique combination
	ObjectSize      int64     `gorm:"type:bigint"`                                       // Object size in bytes as an int64
	JobID           string    `gorm:"type:varchar(255)"`                                 // Job ID as a string
	PelicanExitCode int       `gorm:"type:int"`                                          // Pelican client exit code as an integer
	PelicanStdout   string    `gorm:"type:text"`                                         // Pelican client stdout as a string
	PelicanStderr   string    `gorm:"type:text"`                                         // Pelican client stderr as a string
}

var (
	DB  *gorm.DB
	log = logger.With(zap.String("component", "database"))
)

// Initialize sets up the database connection and runs migrations.
func InitializeDB() {
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

func InsertOrUpdateStagingRecord(pelicanURL, stagingStorage, jobID string, objectSize int64, exitCode int, stdout, stderr string) error {
	// Check if the record with the given combination already exists
	var existingRecord StagingRecord
	err := DB.Where("pelican_url = ? AND staging_storage = ?", pelicanURL, stagingStorage).First(&existingRecord).Error

	if err == nil {
		// Record exists, update it
		existingRecord.ObjectSize = objectSize
		existingRecord.JobID = jobID
		existingRecord.PelicanExitCode = exitCode
		existingRecord.PelicanStdout = stdout
		existingRecord.PelicanStderr = stderr

		if updateErr := DB.Save(&existingRecord).Error; updateErr != nil {
			return fmt.Errorf("failed to update record: %v", updateErr)
		}
	} else if err == gorm.ErrRecordNotFound {
		// Record does not exist, create a new one
		newRecord := StagingRecord{
			PelicanURL:      pelicanURL,
			StagingStorage:  stagingStorage,
			ObjectSize:      objectSize,
			JobID:           jobID,
			PelicanExitCode: exitCode,
			PelicanStdout:   stdout,
			PelicanStderr:   stderr,
		}

		if createErr := DB.Create(&newRecord).Error; createErr != nil {
			return fmt.Errorf("failed to create new record: %v", createErr)
		}
	} else {
		// Some other error occurred during the query
		return fmt.Errorf("error checking existing record: %v", err)
	}

	return nil
}
