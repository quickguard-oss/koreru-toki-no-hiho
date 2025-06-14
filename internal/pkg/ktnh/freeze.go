package ktnh

import (
	"fmt"
	"log/slog"
	"time"
)

/*
Freeze creates a CloudFormation stack to keep the Aurora cluster or RDS instance stopped.
*/
func (k *ktnh) Freeze(templateBody string, qualifier string, timeout time.Duration) error {
	existingStackName, found, err := k.findMatchingStack()

	if err != nil {
		return fmt.Errorf("error while checking for existing stacks: %w", err)
	}

	if found {
		return fmt.Errorf("stack '%s' for DB identifier '%s' already exists", existingStackName, k.dbIdentifier)
	}

	newStackName := k.generateStackName(&stackNameOption{
		dbIdentifierShort: k.dbIdentifierShort,
		qualifier:         qualifier,
	})

	slog.Info("Creating CloudFormation stack", "stackName", newStackName)

	err = k.cfn.CreateStack(newStackName, templateBody)

	if err != nil {
		return fmt.Errorf("failed to create CloudFormation stack: %w", err)
	}

	if timeout == 0 {
		slog.Info("Skipped wait for stack creation")

		return nil
	}

	slog.Info("Waiting for CloudFormation stack creation to complete", "timeout", timeout.Seconds())

	err = k.cfn.WaitForStackCreation(newStackName, timeout)

	if err != nil {
		return fmt.Errorf("failed while waiting for stack creation: %w", err)
	}

	return nil
}
