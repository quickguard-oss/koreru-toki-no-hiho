package awsfactory

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

/*
CloudFormationFactory defines the main interface for creating AWS CloudFormation service clients and helpers.
*/
type CloudFormationFactory interface {
	GetClient() CloudFormationClient
	NewListStacksPaginator(params *cloudformation.ListStacksInput) (ListStacksPaginator, error)
	NewStackCreateCompleteWaiter() (StackCreateCompleteWaiter, error)
	NewStackDeleteCompleteWaiter() (StackDeleteCompleteWaiter, error)
}

/*
CloudFormationClient defines the interface for CloudFormation operations.
*/
type CloudFormationClient interface {
	CreateStack(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	GetTemplate(ctx context.Context, params *cloudformation.GetTemplateInput, optFns ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error)
	ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
}

/*
ListStacksPaginator defines the interface for paginating through stack lists.
*/
type ListStacksPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
}

/*
StackCreateCompleteWaiter defines the interface for waiting for a stack creation to complete.
*/
type StackCreateCompleteWaiter interface {
	Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackCreateCompleteWaiterOptions)) error
}

/*
StackDeleteCompleteWaiter defines the interface for waiting for a stack deletion to complete.
*/
type StackDeleteCompleteWaiter interface {
	Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error
}

/*
defaultCloudFormationFactory is the default implementation of the CloudFormationFactory interface.
*/
type defaultCloudFormationFactory struct {
	client CloudFormationClient // CloudFormation client
}

/*
NewCloudFormationFactory creates and returns a new instance of defaultCloudFormationFactory.
*/
func NewCloudFormationFactory() (CloudFormationFactory, error) {
	client, err := initializeCloudFormationClient()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize CloudFormation client: %w", err)
	}

	return &defaultCloudFormationFactory{
		client: client,
	}, nil
}

/*
initializeCloudFormationClient initializes the CloudFormation client.
*/
func initializeCloudFormationClient() (CloudFormationClient, error) {
	slog.Debug("Initializing CloudFormation client")

	err := loadAWSConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)

	slog.Debug("CloudFormation client initialized")

	return client, nil
}

/*
GetClient returns an instance of the CloudFormation client.
*/
func (f *defaultCloudFormationFactory) GetClient() CloudFormationClient {
	return f.client
}

/*
NewListStacksPaginator creates a new instance of the ListStacksPaginator.
*/
func (f *defaultCloudFormationFactory) NewListStacksPaginator(params *cloudformation.ListStacksInput) (ListStacksPaginator, error) {
	slog.Debug("Creating new ListStacks paginator")

	client, err := f.getTypedClient()

	if err != nil {
		return nil, fmt.Errorf("failed to get typed client: %w", err)
	}

	paginator := cloudformation.NewListStacksPaginator(client, params)

	slog.Debug("ListStacks paginator created successfully")

	return paginator, nil
}

/*
NewStackCreateCompleteWaiter creates a new instance of the StackCreateCompleteWaiter.
*/
func (f *defaultCloudFormationFactory) NewStackCreateCompleteWaiter() (StackCreateCompleteWaiter, error) {
	slog.Debug("Creating new StackCreateComplete waiter")

	client, err := f.getTypedClient()

	if err != nil {
		return nil, fmt.Errorf("failed to get typed client: %w", err)
	}

	waiter := cloudformation.NewStackCreateCompleteWaiter(client)

	slog.Debug("StackCreateComplete waiter created successfully")

	return waiter, nil
}

/*
NewStackDeleteCompleteWaiter creates a new instance of the StackDeleteCompleteWaiter.
*/
func (f *defaultCloudFormationFactory) NewStackDeleteCompleteWaiter() (StackDeleteCompleteWaiter, error) {
	slog.Debug("Creating new StackDeleteComplete waiter")

	client, err := f.getTypedClient()

	if err != nil {
		return nil, fmt.Errorf("failed to get typed client: %w", err)
	}

	waiter := cloudformation.NewStackDeleteCompleteWaiter(client)

	slog.Debug("StackDeleteComplete waiter created successfully")

	return waiter, nil
}

/*
getTypedClient returns the CloudFormation client as the concrete type *cloudformation.Client.
*/
func (f *defaultCloudFormationFactory) getTypedClient() (*cloudformation.Client, error) {
	slog.Debug("Retrieving typed CloudFormation client")

	typedClient, ok := f.client.(*cloudformation.Client)

	if !ok {
		return nil, fmt.Errorf("invalid CloudFormation client type")
	}

	slog.Debug("Typed CloudFormation client retrieved successfully")

	return typedClient, nil
}
