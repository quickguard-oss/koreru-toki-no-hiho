package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/ktnh"
)

var defrostCmd = &cobra.Command{
	Use:   "defrost <db-identifier>",
	Short: "Remove indefinite stop configuration for Aurora cluster or RDS instance",
	Long:  "Removes the CloudFormation stack that enforces automatic stopping, returning the database to normal operational state.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbIdentifier := args[0]

		k, err := ktnh.NewKtnh(dbIdentifier, stackPrefixFlag)

		if err != nil {
			return fmt.Errorf("failed to initialize ktnh instance: %w", err)
		}

		slog.Info("Defrosting DB", "dbIdentifier", dbIdentifier)

		err = k.Defrost(timeoutDuration())

		if err != nil {
			return fmt.Errorf("failed to defrost DB: %w", err)
		}

		slog.Info("DB defrosted successfully")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(defrostCmd)
}
