package ktnh

import (
	"fmt"
	"log/slog"
	"time"
)

/*
Defrost deletes the CloudFormation stack associated with the DB identifier.
*/
func (k *ktnh) Defrost(timeout time.Duration) error {
	stackName, found, err := k.findMatchingStack()

	if err != nil {
		return fmt.Errorf("failed to find matching stack: %w", err)
	}

	if !found {
		return fmt.Errorf("no stacks found for DB identifier")
	}

	slog.Info("Found matching CloudFormation stack, deleting", "stackName", stackName)

	err = k.cfn.DeleteStack(stackName)

	if err != nil {
		return fmt.Errorf("failed to delete CloudFormation stack: %w", err)
	}

	if timeout == 0 {
		slog.Info("Skipped wait for stack deletion")

		return nil
	}

	slog.Info("Waiting for CloudFormation stack deletion to complete", "timeout", timeout.Seconds())

	err = k.cfn.WaitForStackDeletion(stackName, timeout)

	if err != nil {
		return fmt.Errorf("failed while waiting for stack deletion: %w", err)
	}

	return nil
}
