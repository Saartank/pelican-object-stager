package pelican

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pelicanplatform/pelicanobjectstager/config"
)

// InvokePelicanBinary executes the Pelican binary with the provided arguments
// and returns stdout and stderr as separate strings.
func InvokePelicanBinary(args []string) (string, string, int, error) {
	binaryPath := config.AppConfig.Pelican.BinaryPath
	if binaryPath == "" {
		return "", "", -1, fmt.Errorf("pelican binary path is not set in configuration")
	}

	// Check if the binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return "", "", -1, fmt.Errorf("pelican binary not found at %s: %v", binaryPath, err)
	}

	cmd := exec.Command(binaryPath, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	var exitCode int
	if err != nil {
		// Check if the error is an *exec.ExitError to extract the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// For non-exit-related errors, return -1 as the exit code
			return stdoutBuf.String(), stderrBuf.String(), -1, fmt.Errorf("failed to execute command: %v", err)
		}
	} else {
		// Set exit code to 0 for successful execution
		exitCode = 0
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, err
}
