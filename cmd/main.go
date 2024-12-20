package main

import (
	"go.uber.org/zap"

	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/db"
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
			stdout, stderr, exitCode, err := pelican.InvokePelicanBinary(args)

			// Consolidated logging
			logger.Base().Info("PelicanBinary execution details",
				zap.String("stdout", stdout),
				zap.String("stderr", stderr),
				zap.Int("pelican_client_exit_code", exitCode),
				zap.Error(err),
			)

			// Handle errors after logging
			if err != nil {
				logger.Base().Fatal("Failed to invoke PelicanBinary", zap.Error(err))
			}
		},
		DisableFlagParsing: true, // Forward unparsed flags directly to the binary
	}

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(pelicanCmd)

	cobra.OnInitialize(func() {
		config.LoadConfig("/etc/pelican/config.yaml")
		db.InitializeDB()
	})

	if err := rootCmd.Execute(); err != nil {
		logger.Base().Error("Failed to initialize", zap.Error(err))
	}
}
