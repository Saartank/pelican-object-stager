package pelican

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/pelicanplatform/pelicanobjectstager/config"
)

// InvokePelicanBinary executes the Pelican binary with the provided arguments
// and returns stdout and stderr as separate strings.
func InvokePelicanBinary(args []string) (string, string, error) {
	binaryPath := config.AppConfig.Pelican.BinaryPath
	if binaryPath == "" {
		return "", "", fmt.Errorf("Pelican binary path is not set in configuration")
	}

	// Check if the binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return "", "", fmt.Errorf("Pelican binary not found at %s: %v", binaryPath, err)
	}

	cmd := exec.Command(binaryPath, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
