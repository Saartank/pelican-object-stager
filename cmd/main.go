package main

import (
	"go.uber.org/zap"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/logger"
	"github.com/pelicanplatform/pelicanobjectstager/pelican"
	"github.com/pelicanplatform/pelicanobjectstager/server"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "github.com/pelicanplatform/pelicanobjectstager",
		Short: "Wrapper for managing Pelican binaries",
	}

	// Subcommand to start the server
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the github.com/pelicanplatform/pelicanobjectstager daemon",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Base().Info("Starting pelicanobjectstager daemon...")
			server.StartServer()
		},
	}

	// Subcommand to invoke PelicanBinary
	var pelicanCmd = &cobra.Command{
		Use:   "pelican [args...]",
		Short: "Invoke the PelicanBinary with the given arguments",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			stdout, stderr, err := pelican.InvokePelicanBinary(args)

			// Log stderr if present
			if stderr != "" {
				logger.Base().Error("PelicanBinary stderr", zap.String("stderr", stderr))
			}

			// Handle errors
			if err != nil {
				logger.Base().Fatal("Failed to invoke PelicanBinary", zap.Error(err))
			}

			// Log stdout if present
			if stdout != "" {
				logger.Base().Info("PelicanBinary stdout", zap.String("stdout", stdout))
			} else {
				logger.Base().Info("PelicanBinary executed successfully with no output")
			}
		},
		DisableFlagParsing: true, // Forward unparsed flags directly to the binary
	}

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(pelicanCmd)

	cobra.OnInitialize(func() {
		config.LoadConfig("/etc/pelican/config.yaml") // Default config location
	})

	if err := rootCmd.Execute(); err != nil {
		logger.Base().Error("Failed to initialize", zap.Error(err))
	}
}
