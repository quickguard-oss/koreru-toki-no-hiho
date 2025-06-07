package cfn

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
stackEvaluator is a function type that evaluates whether a stack should be included in the results
when listing CloudFormation stacks.
*/
type stackEvaluator func(stackName string) bool

/*
CreateStack creates a new CloudFormation stack without waiting for completion.
*/
func (c *CloudFormation) CreateStack(stackName string, templateBody string) error {
	slog.Debug("Starting CloudFormation stack creation", "stackName", stackName)

	ctx := context.Background()

	_, err := c.factory.GetClient().CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(templateBody),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
	})

	if err != nil {
		return fmt.Errorf("failed to execute CreateStack API for stack '%s': %w", stackName, err)
	}

	slog.Debug("CloudFormation stack creation initiated successfully")

	return nil
}

/*
DeleteStack deletes a CloudFormation stack without waiting for completion.
*/
func (c *CloudFormation) DeleteStack(stackName string) error {
	slog.Debug("Starting CloudFormation stack deletion", "stackName", stackName)

	ctx := context.Background()

	_, err := c.factory.GetClient().DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return fmt.Errorf("failed to execute DeleteStack API for stack '%s': %w", stackName, err)
	}

	slog.Debug("CloudFormation stack deletion initiated successfully")

	return nil
}

/*
ListStacks returns stack names that match the given evaluator function.
If no evaluator is provided, all stacks will be returned.
*/
func (c *CloudFormation) ListStacks(evaluator stackEvaluator) ([]string, error) {
	var matchingStacks []string

	slog.Debug("Starting CloudFormation stack listing")

	if evaluator == nil {
		slog.Debug("No evaluator provided, returning all stacks")

		evaluator = func(stackName string) bool {
			return true
		}
	}

	paginator, err := c.factory.NewListStacksPaginator(&cloudformation.ListStacksInput{
		StackStatusFilter: []types.StackStatus{
			types.StackStatusCreateInProgress,
			types.StackStatusCreateFailed,
			types.StackStatusCreateComplete,
			types.StackStatusRollbackInProgress,
			types.StackStatusRollbackFailed,
			types.StackStatusRollbackComplete,
			types.StackStatusDeleteInProgress,
			types.StackStatusDeleteFailed,
			//types.StackStatusDeleteComplete,
			types.StackStatusUpdateInProgress,
			types.StackStatusUpdateCompleteCleanupInProgress,
			types.StackStatusUpdateComplete,
			types.StackStatusUpdateFailed,
			types.StackStatusUpdateRollbackInProgress,
			types.StackStatusUpdateRollbackFailed,
			types.StackStatusUpdateRollbackCompleteCleanupInProgress,
			types.StackStatusUpdateRollbackComplete,
			types.StackStatusReviewInProgress,
			types.StackStatusImportInProgress,
			types.StackStatusImportComplete,
			types.StackStatusImportRollbackInProgress,
			types.StackStatusImportRollbackFailed,
			types.StackStatusImportRollbackComplete,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create ListStacks paginator: %w", err)
	}

	ctx := context.Background()

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to execute ListStacks API: %w", err)
		}

		for _, stack := range output.StackSummaries {
			stackName := aws.ToString(stack.StackName)

			slog.Debug("Evaluating stack", "stackName", stackName)

			result := evaluator(stackName)

			if result {
				matchingStacks = append(matchingStacks, stackName)
			}

			slog.Debug("Stack evaluation result", "matches", result)
		}
	}

	slog.Debug("CloudFormation stack listing completed")

	return matchingStacks, nil
}
