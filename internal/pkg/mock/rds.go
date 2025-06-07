package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/stretchr/testify/mock"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
)

/*
MockRDSFactory is a mock implementation of the `RDSFactory` (internal/pkg/awsfactory) interface.
*/
type MockRDSFactory struct {
	mock.Mock
}

/*
MockRDSClient is a mock implementation of the `RDSClient` (internal/pkg/awsfactory) interface.
*/
type MockRDSClient struct {
	mock.Mock
}

/*
MockDescribeDBClustersPaginator is a mock implementation of the `DescribeDBClustersPaginator` (internal/pkg/awsfactory) interface.
*/
type MockDescribeDBClustersPaginator struct {
	mock.Mock
}

/*
MockDescribePendingMaintenanceActionsPaginator is a mock implementation of the `DescribePendingMaintenanceActionsPaginator` (internal/pkg/awsfactory) interface.
*/
type MockDescribePendingMaintenanceActionsPaginator struct {
	mock.Mock
}

func (m *MockRDSFactory) GetClient() awsfactory.RDSClient {
	args := m.Called()

	return args.Get(0).(*MockRDSClient)
}

func (m *MockRDSFactory) NewDescribeDBClustersPaginator(params *rds.DescribeDBClustersInput) (awsfactory.DescribeDBClustersPaginator, error) {
	args := m.Called(params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MockDescribeDBClustersPaginator), args.Error(1)
}

func (m *MockRDSFactory) NewDescribePendingMaintenanceActionsPaginator(params *rds.DescribePendingMaintenanceActionsInput) (awsfactory.DescribePendingMaintenanceActionsPaginator, error) {
	args := m.Called(params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MockDescribePendingMaintenanceActionsPaginator), args.Error(1)
}

func (m *MockRDSClient) DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*rds.DescribeDBClustersOutput), args.Error(1)
}

func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*rds.DescribeDBInstancesOutput), args.Error(1)
}

func (m *MockRDSClient) DescribePendingMaintenanceActions(ctx context.Context, params *rds.DescribePendingMaintenanceActionsInput, optFns ...func(*rds.Options)) (*rds.DescribePendingMaintenanceActionsOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*rds.DescribePendingMaintenanceActionsOutput), args.Error(1)
}

func (m *MockDescribeDBClustersPaginator) HasMorePages() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockDescribeDBClustersPaginator) NextPage(ctx context.Context, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
	args := m.Called(ctx, optFns)

	return args.Get(0).(*rds.DescribeDBClustersOutput), args.Error(1)
}

func (m *MockDescribePendingMaintenanceActionsPaginator) HasMorePages() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockDescribePendingMaintenanceActionsPaginator) NextPage(ctx context.Context, optFns ...func(*rds.Options)) (*rds.DescribePendingMaintenanceActionsOutput, error) {
	args := m.Called(ctx, optFns)

	return args.Get(0).(*rds.DescribePendingMaintenanceActionsOutput), args.Error(1)
}
