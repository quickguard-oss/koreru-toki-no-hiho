package awsfactory

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

/*
RDSFactory defines the main interface for creating Amazon RDS service clients and helpers.
*/
type RDSFactory interface {
	GetClient() RDSClient
	NewDescribeDBClustersPaginator(params *rds.DescribeDBClustersInput) (DescribeDBClustersPaginator, error)
	NewDescribePendingMaintenanceActionsPaginator(params *rds.DescribePendingMaintenanceActionsInput) (DescribePendingMaintenanceActionsPaginator, error)
}

/*
RDSClient defines the interface for RDS operations.
*/
type RDSClient interface {
	DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribePendingMaintenanceActions(ctx context.Context, params *rds.DescribePendingMaintenanceActionsInput, optFns ...func(*rds.Options)) (*rds.DescribePendingMaintenanceActionsOutput, error)
}

/*
DescribeDBClustersPaginator defines the interface for paginating through DB clusters.
*/
type DescribeDBClustersPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
}

/*
DescribePendingMaintenanceActionsPaginator defines the interface for paginating through pending maintenance actions.
*/
type DescribePendingMaintenanceActionsPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*rds.Options)) (*rds.DescribePendingMaintenanceActionsOutput, error)
}

/*
defaultRDSFactory is the default implementation of the RDSFactory interface.
*/
type defaultRDSFactory struct {
	client RDSClient // RDS client
}

/*
NewRDSFactory creates and returns a new instance of defaultRDSFactory.
*/
func NewRDSFactory() (RDSFactory, error) {
	client, err := initializeRDSClient()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize RDS client: %w", err)
	}

	return &defaultRDSFactory{
		client: client,
	}, nil
}

/*
initializeRDSClient initializes the RDS client.
*/
func initializeRDSClient() (RDSClient, error) {
	slog.Debug("Initializing RDS client")

	err := loadAWSConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	client := rds.NewFromConfig(cfg)

	slog.Debug("RDS client initialized")

	return client, nil
}

/*
GetClient returns an instance of the RDS client.
*/
func (f *defaultRDSFactory) GetClient() RDSClient {
	return f.client
}

/*
NewDescribeDBClustersPaginator creates a new instance of the DescribeDBClustersPaginator.
*/
func (f *defaultRDSFactory) NewDescribeDBClustersPaginator(params *rds.DescribeDBClustersInput) (DescribeDBClustersPaginator, error) {
	slog.Debug("Creating new DescribeDBClusters paginator")

	client, err := f.getTypedClient()

	if err != nil {
		return nil, fmt.Errorf("failed to get typed client: %w", err)
	}

	paginator := rds.NewDescribeDBClustersPaginator(client, params)

	slog.Debug("DescribeDBClusters paginator created successfully")

	return paginator, nil
}

/*
NewDescribePendingMaintenanceActionsPaginator creates a new instance of the DescribePendingMaintenanceActionsPaginator.
*/
func (f *defaultRDSFactory) NewDescribePendingMaintenanceActionsPaginator(params *rds.DescribePendingMaintenanceActionsInput) (DescribePendingMaintenanceActionsPaginator, error) {
	slog.Debug("Creating new DescribePendingMaintenanceActions paginator")

	client, err := f.getTypedClient()

	if err != nil {
		return nil, fmt.Errorf("failed to get typed client: %w", err)
	}

	paginator := rds.NewDescribePendingMaintenanceActionsPaginator(client, params)

	slog.Debug("DescribePendingMaintenanceActions paginator created successfully")

	return paginator, nil
}

/*
getTypedClient returns the RDS client as the concrete type *rds.Client.
*/
func (f *defaultRDSFactory) getTypedClient() (*rds.Client, error) {
	slog.Debug("Retrieving typed RDS client")

	typedClient, ok := f.client.(*rds.Client)

	if !ok {
		return nil, fmt.Errorf("invalid RDS client type")
	}

	slog.Debug("Typed RDS client retrieved successfully")

	return typedClient, nil
}
