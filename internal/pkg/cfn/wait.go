package cfn

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
)

/*
operation represents the type of CloudFormation stack operation being performed.
It is used to distinguish between different stack operations such as creation and deletion.
*/
type operation string

/*
completeWaiter encapsulates the waiters for different CloudFormation stack operations.
It includes a reference to the operation type and the appropriate waiter objects
from the CloudFormation SDK for creation and deletion operations.
*/
type completeWaiter struct {
	operation operation // type of operation ("creation" or "deletion")

	create awsfactory.StackCreateCompleteWaiter // waiter for stack creation completion
	delete awsfactory.StackDeleteCompleteWaiter // waiter for stack deletion completion
}

const (
	operationCreate = operation("creation") // stack creation operation
	operationDelete = operation("deletion") // stack deletion operation
)

/*
WaitForStackCreation waits for a CloudFormation stack creation to complete.
*/
func (c *CloudFormation) WaitForStackCreation(stackName string, timeout time.Duration) error {
	createWaiter, err := c.factory.NewStackCreateCompleteWaiter()

	if err != nil {
		return fmt.Errorf("failed to create StackCreateComplete waiter: %w", err)
	}

	waiter := completeWaiter{
		operation: operationCreate,
		create:    createWaiter,
	}

	return c.waitForStackOperation(stackName, timeout, waiter)
}

/*
WaitForStackDeletion waits for a CloudFormation stack deletion to complete.
*/
func (c *CloudFormation) WaitForStackDeletion(stackName string, timeout time.Duration) error {
	deleteWaiter, err := c.factory.NewStackDeleteCompleteWaiter()

	if err != nil {
		return fmt.Errorf("failed to create StackDeleteComplete waiter: %w", err)
	}

	waiter := completeWaiter{
		operation: operationDelete,
		delete:    deleteWaiter,
	}

	return c.waitForStackOperation(stackName, timeout, waiter)
}

/*
waitForStackOperation waits for a CloudFormation stack operation (create or delete) to complete.
*/
func (c *CloudFormation) waitForStackOperation(stackName string, timeout time.Duration, waiter completeWaiter) error {
	slog.Debug("Waiting for stack operation to complete",
		"stackName", stackName,
		"timeout", timeout.Seconds(),
		"operation", waiter.operation,
	)

	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	timerCtx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func() {
		startTime := time.Now()

		ticker := time.NewTicker(10 * time.Second)

		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				slog.Info("Waiting for stack operation to complete",
					"stackName", stackName,
					"operation", waiter.operation,
					"elapsed", time.Since(startTime).Seconds(),
				)
			case <-timerCtx.Done():
				return
			}
		}
	}()

	var err error

	ctx := context.Background()

	switch waiter.operation {
	case operationCreate:
		optFunc := func(opt *cloudformation.StackCreateCompleteWaiterOptions) {
			opt.MinDelay = 10 * time.Second
			opt.MaxDelay = 15 * time.Second
		}

		err = waiter.create.Wait(ctx, input, timeout, optFunc)
	case operationDelete:
		optFunc := func(opt *cloudformation.StackDeleteCompleteWaiterOptions) {
			opt.MinDelay = 10 * time.Second
			opt.MaxDelay = 15 * time.Second
		}

		err = waiter.delete.Wait(ctx, input, timeout, optFunc)
	default:
		return fmt.Errorf("unknown operation '%s'", waiter.operation)
	}

	if err != nil {
		return fmt.Errorf("error while waiting for stack '%s' %s to complete: %w", stackName, waiter.operation, err)
	}

	slog.Debug("CloudFormation stack operation completed successfully")

	return nil
}
