package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pelicanplatform/pelicanobjectstager/db"
	"github.com/pelicanplatform/pelicanobjectstager/pelican"
)

func handleStartBinary(c *gin.Context) {
	// Retrieve the Job ID from the context
	jobID, _ := c.Get("job_id")

	// Define a struct to bind the JSON request body
	type RequestBody struct {
		Args []string `json:"args"` // Arguments to pass to the binary
	}

	// Parse the JSON request body
	var requestBody RequestBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		log.Error("Failed to parse request body",
			zap.String("job_id", jobID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"job_id":  jobID, // Include Job ID
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Use provided arguments or default to ["default-command"]
	args := requestBody.Args
	if len(args) == 0 {
		args = []string{"default-command"}
	}

	// Invoke the binary with the arguments
	stdout, stderr, exitCode, err := pelican.InvokePelicanBinary(args)
	if stderr != "" {
		log.Error("PelicanBinary stderr",
			zap.String("job_id", jobID.(string)),
			zap.Int("pelican_client_exit_code", exitCode),
			zap.String("stderr", stderr),
		)
	}

	// Handle execution errors
	if err != nil {
		log.Error("Failed to execute PelicanBinary",
			zap.String("job_id", jobID.(string)),
			zap.Int("pelican_client_exit_code", exitCode),
			zap.Error(err),
			zap.String("stderr", stderr),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"job_id":                   jobID, // Include Job ID
			"pelican_client_exit_code": exitCode,
			"error":                    "Failed to execute PelicanBinary",
			"details":                  err.Error(),
			"stderr":                   stderr,
		})
		return
	}

	// Log and return the successful output
	log.Info("PelicanBinary executed successfully",
		zap.String("job_id", jobID.(string)),
		zap.Int("pelican_client_exit_code", exitCode),
		zap.String("stdout", stdout),
	)
	c.JSON(http.StatusOK, gin.H{
		"job_id":                   jobID, // Include Job ID
		"pelican_client_exit_code": exitCode,
		"message":                  "PelicanBinary executed successfully",
		"stdout":                   stdout,
	})
}

func handleHealthCheck(c *gin.Context) {
	// Retrieve the Job ID from the context
	jobID, _ := c.Get("job_id")

	// Run the `--version` command on the binary
	stdout, stderr, exitCode, err := pelican.InvokePelicanBinary([]string{"--version"})

	// Handle stderr if present
	if stderr != "" {
		log.Error("PelicanBinary stderr",
			zap.String("job_id", jobID.(string)),
			zap.Int("pelican_client_exit_code", exitCode),
			zap.String("stderr", stderr),
		)
	}

	// Handle errors
	if err != nil {
		log.Error("Failed to execute PelicanBinary --version",
			zap.String("job_id", jobID.(string)),
			zap.Int("pelican_client_exit_code", exitCode),
			zap.Error(err),
			zap.String("stderr", stderr),
		)
		c.JSON(500, gin.H{
			"job_id":                   jobID, // Include Job ID
			"pelican_client_exit_code": exitCode,
			"status":                   "error",
			"message":                  "PelicanBinary failed to execute",
			"error":                    err.Error(),
			"stderr":                   stderr,
		})
		return
	}

	// Process stdout for version information
	version := strings.TrimSpace(stdout)
	log.Info("Health check successful",
		zap.String("job_id", jobID.(string)),
		zap.String("version", version),
		zap.Int("pelican_client_exit_code", exitCode),
	)
	c.JSON(200, gin.H{
		"job_id":  jobID, // Include Job ID
		"status":  "ok",
		"message": "PelicanBinary is working",
		"version": version,
	})
}

func handleRecordsAll(c *gin.Context) {
	// Retrieve all records
	records, err := db.GetStagingRecordLites()
	if err != nil {
		log.Error("Failed to retrieve all staging records",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
	})
}

func handleStagingStoragesAll(c *gin.Context) {
	storageSizeMap, err := db.GetStagingStorageSizeMap()
	if err != nil {
		log.Error("Failed to retrieve staging storage size map",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve storage size map",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"storage_sizes": storageSizeMap,
	})
}

func handleGetRecordByID(c *gin.Context) {
	// Parse the ID from the URL parameter
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Error("Invalid ID format",
			zap.String("id", idParam),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	// Fetch the record from the database
	record, err := db.GetStagingRecordByID(uint(id))
	if err != nil {
		log.Error("Failed to retrieve record",
			zap.Int("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve the record",
		})
		return
	}

	if record == nil {
		log.Info("Record not found",
			zap.Int("id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Record not found",
		})
		return
	}

	c.JSON(http.StatusOK, record)
}
