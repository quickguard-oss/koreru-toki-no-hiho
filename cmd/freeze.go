package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/ktnh"
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/logger"
)

var (
	templateFlag bool
)

var freezeCmd = &cobra.Command{
	Use:   "freeze <db-identifier>",
	Short: "Keep specified Aurora cluster or RDS instance permanently stopped",
	Long:  "Creates the CloudFormation stack to keep the specified Aurora cluster or RDS instance in a permanently stopped state.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbIdentifier := args[0]

		k, err := ktnh.NewKtnh(dbIdentifier, stackPrefixFlag)

		if err != nil {
			return fmt.Errorf("failed to initialize ktnh instance: %w", err)
		}

		templateBody, qualifier, err := k.Template()

		if err != nil {
			return fmt.Errorf("failed to generate CloudFormation template: %w", err)
		}

		if templateFlag {
			var output string

			if jsonLogFlag {
				output, err = logger.FormatAsJSON([]string{"content"}, [][]string{{templateBody}})

				if err != nil {
					return fmt.Errorf("failed to format template as JSON: %w", err)
				}
			} else {
				output = templateBody
			}

			cmd.Println(output)

			return nil
		}

		slog.Info("Freezing DB", "dbIdentifier", dbIdentifier)

		err = k.Freeze(templateBody, qualifier, timeoutDuration())

		if err != nil {
			return fmt.Errorf("failed to freeze DB: %w", err)
		}

		slog.Info("DB frozen successfully")

		return nil
	},
}

func init() {
	freezeCmd.Flags().BoolVarP(&templateFlag, "template", "t", false, "display CloudFormation template without creating stack")

	rootCmd.AddCommand(freezeCmd)
}
