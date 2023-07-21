package cmd

import (
	"log"
	"os"

	"github.com/1989michael/tinyurl/internal/cmd/server"
	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

// ExitFailure status code.
const ExitFailure = 1

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	//nolint: exhaustruct
	root := &cobra.Command{
		Use:     "tinyurl",
		Short:   "Shorten your URLs to make them more memorable",
		Version: versioninfo.Short(),
	}

	server.Register(root)

	if err := root.Execute(); err != nil {
		log.Printf("failed to execute root command %s", err)
		os.Exit(ExitFailure)
	}
}
