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

func (m *MockRDSFactory) GetClient() awsfactory.RDSClient {
	args := m.Called()

	return args.Get(0).(*MockRDSClient)
}

func (m *MockRDSClient) DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*rds.DescribeDBClustersOutput), args.Error(1)
}

func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)

	return args.Get(0).(*rds.DescribeDBInstancesOutput), args.Error(1)
}
