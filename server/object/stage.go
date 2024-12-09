package object

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pelicanplatform/pelicanobjectstager/pelican"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// StageRequest represents the input structure for the /object/stage endpoint
type StageRequest struct {
	Entries     []RequestEntry `json:"entries" binding:"required"`      // List of entries
	TargetCache string         `json:"target_cache" binding:"required"` // Target cache
}

// RequestEntry represents a single request entry
type RequestEntry struct {
	RequestURL string `json:"request_url" binding:"required"` // Object URL
	Parameters string `json:"parameters,omitempty"`           // Optional flags/options
}

func HandleStage(c *gin.Context) {
	var input StageRequest

	// Parse and validate JSON payload
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Worker pool setup
	numWorkers := 5 // Configurable
	entryChan := make(chan RequestEntry, len(input.Entries))
	resultsChan := make(chan map[string]interface{}, len(input.Entries))
	var wg sync.WaitGroup

	// Start staging workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go stagingWorker(entryChan, input.TargetCache, resultsChan, &wg)
	}

	// Send entries to the workers
	for _, entry := range input.Entries {
		entryChan <- entry
	}
	close(entryChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultsChan)

	// Collect results
	results := make(map[string]interface{})
	hasErrors := false

	for result := range resultsChan {
		url := result["request_url"].(string)
		status := result["result"]

		results[url] = status
		if status != "success" {
			hasErrors = true
		}
	}

	// Determine response status
	if hasErrors {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Staging completed with errors",
			"results": results,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Staging completed successfully",
			"results": results,
		})
	}
}

// stagingWorker processes a single entry and sends results to channels
func stagingWorker(entries <-chan RequestEntry, targetCache string, results chan<- map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Get temp destination from configuration
	tempDestination := viper.GetString("staging.temp_destination")

	for entry := range entries {
		args := []string{"object", "get", entry.RequestURL, tempDestination}

		if entry.Parameters != "" {
			parameterArgs := strings.Fields(entry.Parameters) // Split by space
			args = append(args, parameterArgs...)
		}

		args = append(args, "--cache", targetCache)

		// Log debug information
		logrus.WithFields(logrus.Fields{
			"request_url":      entry.RequestURL,
			"parameters":       entry.Parameters,
			"parsed_args":      args,
			"temp_destination": tempDestination,
		}).Debug("Processing entry")

		// Invoke the binary
		stdout, stderr, err := pelican.InvokePelicanBinary(args)

		// Prepare result
		if err != nil {
			errorMessage := stderr
			// If stderr is empty, use the default error message
			if stderr == "" {
				errorMessage = err.Error()
			}

			logrus.WithFields(logrus.Fields{
				"request_url": entry.RequestURL,
				"stdout":      stdout,
				"stderr":      stderr,
				"error":       errorMessage,
			}).Error("Failed to process entry")
			results <- map[string]interface{}{
				"request_url": entry.RequestURL,
				"result":      errorMessage,
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"request_url": entry.RequestURL,
				"stdout":      stdout,
				"stderr":      stderr,
			}).Info("Entry processed successfully")
			results <- map[string]interface{}{
				"request_url": entry.RequestURL,
				"result":      "success",
			}
		}
	}
}
