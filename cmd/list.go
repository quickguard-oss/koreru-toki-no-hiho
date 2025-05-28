package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/ktnh"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases managed by ktnh",
	Long:  "Lists all Aurora clusters or RDS instances that are being kept in a permanently stopped state by ktnh.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		k, err := ktnh.NewKtnh("", stackPrefixFlag)

		if err != nil {
			return fmt.Errorf("failed to initialize ktnh instance: %w", err)
		}

		lines, err := k.List()

		if err != nil {
			return fmt.Errorf("failed to list managed databases: %w", err)
		}

		if len(lines) == 0 {
			slog.Info("No databases are currently being managed by ktnh")

			return nil
		}

		for _, line := range lines {
			cmd.Println(line)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
