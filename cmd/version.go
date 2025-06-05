package cmd

import (
	"log/slog"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "(unknown)" // version number
	commit  = "(unknown)" // commit hash
	built   = "(unknown)" // build date

	versionOverridden = "false" // flag to detect if version was overridden by ldflags
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Displays version number, commit hash, and build date of the ktnh command.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionOverridden != "true" {
			info, ok := debug.ReadBuildInfo()

			if ok {
				version = info.Main.Version
			}
		}

		slog.Info("ktnh", "version", version, "commit", commit, "built", built)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
