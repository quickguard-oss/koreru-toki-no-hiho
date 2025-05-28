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
}

/*
RDSClient defines the interface for RDS operations.
*/
type RDSClient interface {
	DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
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
