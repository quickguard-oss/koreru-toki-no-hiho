package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/logger"
)

var (
	jsonLogFlag     bool
	noWaitFlag      bool
	stackPrefixFlag string
	verboseFlag     bool
	waitTimeoutFlag time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "ktnh",
	Short: "Keep Aurora clusters or RDS instances stopped permanently",
	Long: `ktnh prevents Amazon Aurora clusters or RDS instances from automatically restarting after 7 days,
keeping them in a stopped state indefinitely.
It uses CloudFormation to create and manage the necessary AWS resources.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validateStackPrefix(); err != nil {
			return fmt.Errorf("invalid --prefix '%s': %w", stackPrefixFlag, err)
		}

		if err := validateWaitTimeout(); err != nil {
			return fmt.Errorf("invalid --wait-timeout '%s': %w", waitTimeoutFlag, err)
		}

		logger.SetLogger(verboseFlag, jsonLogFlag)

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonLogFlag, "json-log", "j", false, "output logs in JSON format instead of plain text")
	rootCmd.PersistentFlags().BoolVar(&noWaitFlag, "no-wait", false, "don't wait for CloudFormation stack operation to complete")
	rootCmd.PersistentFlags().StringVarP(&stackPrefixFlag, "prefix", "p", "ktnh", "prefix for CloudFormation stack name (1-10 alphanumeric characters)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().DurationVar(&waitTimeoutFlag, "wait-timeout", 15*time.Minute, "timeout duration for waiting on stack operation")
}

/*
Execute starts the application and handles any errors.
It will exit with status code 1 if the command execution fails.
*/
func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	err := rootCmd.Execute()

	if err != nil {
		slog.Error("Failed to execute command", "err", err)

		os.Exit(1)
	}
}

/*
validateStackPrefix validates whether the --prefix value is valid.
*/
func validateStackPrefix() error {
	if len(stackPrefixFlag) <= 0 || 11 <= len(stackPrefixFlag) {
		return fmt.Errorf("--prefix must be between 1 and 10 characters long")
	}

	match, err := regexp.MatchString("^[A-Za-z0-9]+$", stackPrefixFlag)

	if err != nil {
		return fmt.Errorf("failed to validate --prefix: %w", err)
	}

	if !match {
		return fmt.Errorf("--prefix must only contain alphanumeric characters (A-Z, a-z, 0-9)")
	}

	return nil
}

/*
validateWaitTimeout validates whether the --wait-timeout value is valid.
*/
func validateWaitTimeout() error {
	if waitTimeoutFlag <= 0 {
		return fmt.Errorf("--wait-timeout must be greater than 0")
	}

	return nil
}

/*
timeoutDuration returns the timeout duration for waiting on stack operations.
*/
func timeoutDuration() time.Duration {
	if noWaitFlag {
		return 0
	}

	return waitTimeoutFlag
}
