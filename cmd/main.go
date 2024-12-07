package main

import (
	"github.com/pelicanplatform/pelicanobjectstager/config"
	"github.com/pelicanplatform/pelicanobjectstager/pelican"
	"github.com/pelicanplatform/pelicanobjectstager/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	var rootCmd = &cobra.Command{
		Use:   "github.com/pelicanplatform/pelicanobjectstager",
		Short: "Wrapper for managing Pelican binaries",
	}

	// Subcommand to start the server
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the github.com/pelicanplatform/pelicanobjectstager daemon",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Info("Starting github.com/pelicanplatform/pelicanobjectstager daemon...")
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
				logrus.Errorf("PelicanBinary stderr: %s", stderr)
			}

			// Handle errors
			if err != nil {
				logrus.Fatalf("Failed to invoke PelicanBinary: %v", err)
			}

			// Log stdout if present
			if stdout != "" {
				logrus.Infof("PelicanBinary stdout: %s", stdout)
			} else {
				logrus.Info("PelicanBinary executed successfully with no output")
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
		logrus.Fatal(err)
	}
}
