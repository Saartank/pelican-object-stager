package server

import (
	"net/http"
	"strings"

	"github.com/pelicanplatform/pelicanobjectstager/pelican"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// handleStartBinary - Starts the binary with arguments from the request
func handleStartBinary(c *gin.Context) {
	// Define a struct to bind the JSON request body
	type RequestBody struct {
		Args []string `json:"args"` // Arguments to pass to the binary
	}

	// Parse the JSON request body
	var requestBody RequestBody
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logrus.Errorf("Failed to parse request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
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
	stdout, stderr, err := pelican.InvokePelicanBinary(args)
	if stderr != "" {
		logrus.Errorf("PelicanBinary stderr: %s", stderr)
	}

	// Handle execution errors
	if err != nil {
		logrus.Errorf("Failed to execute PelicanBinary: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute PelicanBinary",
			"details": err.Error(),
			"stderr":  stderr,
		})
		return
	}

	// Log and return the successful output
	logrus.Infof("PelicanBinary executed successfully: %s", stdout)
	c.JSON(http.StatusOK, gin.H{
		"message": "PelicanBinary executed successfully",
		"stdout":  stdout,
	})
}

// handleHealthCheck - Health check handler
func handleHealthCheck(c *gin.Context) {
	// Run the `--version` command on the binary
	stdout, stderr, err := pelican.InvokePelicanBinary([]string{"--version"})

	// Handle stderr if present
	if stderr != "" {
		logrus.Errorf("PelicanBinary stderr: %s", stderr)
	}

	// Handle errors
	if err != nil {
		logrus.Errorf("Failed to execute PelicanBinary --version: %v", err)
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "PelicanBinary failed to execute",
			"error":   err.Error(),
			"stderr":  stderr,
		})
		return
	}

	// Process stdout for version information
	version := strings.TrimSpace(stdout)
	logrus.Infof("Health check successful: PelicanBinary version: %s", version)
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "PelicanBinary is working",
		"version": version,
	})
}
