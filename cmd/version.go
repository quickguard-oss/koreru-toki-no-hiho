package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	version = "dev"     // version number
	commit  = "HEAD"    // commit hash
	built   = "unknown" // build date
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Displays version number, commit hash, and build date of the ktnh command.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("ktnh", "version", version, "commit", commit, "built", built)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
