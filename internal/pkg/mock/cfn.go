package mock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/stretchr/testify/mock"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
)

/*
MockCloudFormationFactory is a mock implementation of the `CloudFormationFactory` (internal/pkg/awsfactory) interface.
*/
type MockCloudFormationFactory struct {
	mock.Mock
}

/*
MockCloudFormationClient is a mock implementation of the `CloudFormationClient` (internal/pkg/awsfactory) interface.
*/
type MockCloudFormationClient struct {
	mock.Mock
}

/*
MockListStacksPaginator is a mock implementation of the `ListStacksPaginator` (internal/pkg/awsfactory) interface.
*/
type MockListStacksPaginator struct {
	mock.Mock
}

/*
MockStackCreateCompleteWaiter is a mock implementation of the `StackCreateCompleteWaiter` (internal/pkg/awsfactory) interface.
*/
type MockStackCreateCompleteWaiter struct {
	mock.Mock
}

/*
MockStackDeleteCompleteWaiter is a mock implementation of the `StackDeleteCompleteWaiter` (internal/pkg/awsfactory) interface.
*/
type MockStackDeleteCompleteWaiter struct {
	mock.Mock
}

func (m *MockCloudFormationFactory) GetClient() awsfactory.CloudFormationClient {
	args := m.Called()

	return args.Get(0).(*MockCloudFormationClient)
}

func (m *MockCloudFormationFactory) NewListStacksPaginator(params *cloudformation.ListStacksInput) (awsfactory.ListStacksPaginator, error) {
	args := m.Called(params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MockListStacksPaginator), args.Error(1)
}

func (m *MockCloudFormationFactory) NewStackCreateCompleteWaiter() (awsfactory.StackCreateCompleteWaiter, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MockStackCreateCompleteWaiter), args.Error(1)
}

func (m *MockCloudFormationFactory) NewStackDeleteCompleteWaiter() (awsfactory.StackDeleteCompleteWaiter, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MockStackDeleteCompleteWaiter), args.Error(1)
}

func (m *MockCloudFormationClient) CreateStack(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*cloudformation.CreateStackOutput), args.Error(1)
}

func (m *MockCloudFormationClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*cloudformation.DeleteStackOutput), args.Error(1)
}

func (m *MockCloudFormationClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*cloudformation.DescribeStacksOutput), args.Error(1)
}

func (m *MockCloudFormationClient) GetTemplate(ctx context.Context, params *cloudformation.GetTemplateInput, optFns ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*cloudformation.GetTemplateOutput), args.Error(1)
}

func (m *MockCloudFormationClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*cloudformation.ListStacksOutput), args.Error(1)
}

func (m *MockListStacksPaginator) HasMorePages() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockListStacksPaginator) NextPage(ctx context.Context, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	args := m.Called(ctx, optFns)

	return args.Get(0).(*cloudformation.ListStacksOutput), args.Error(1)
}

func (m *MockStackCreateCompleteWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackCreateCompleteWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)

	return args.Error(0)
}

func (m *MockStackDeleteCompleteWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)

	return args.Error(0)
}
