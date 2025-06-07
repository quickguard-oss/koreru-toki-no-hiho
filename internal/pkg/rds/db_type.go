package rds

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

/*
dbType represents the type of database.
*/
type dbType string

const (
	dbTypeAurora dbType = "aurora" // Aurora cluster
	dbTypeRDS    dbType = "rds"    // RDS instance
)

/*
isAuroraEngine checks if the engine is an Aurora engine.
*/
func isAuroraEngine(engine string) (bool, error) {
	if engine == "" {
		return false, fmt.Errorf("engine is nil")
	}

	if strings.HasPrefix(engine, "aurora-") {
		slog.Debug("Engine is Aurora", "engine", engine)

		return true, nil
	}

	slog.Debug("Engine is RDS", "engine", engine)

	return false, nil
}

/*
DetermineDBType determines if the provided DB identifier is for an Aurora cluster or RDS instance.
*/
func (r *RDS) DetermineDBType(dbIdentifier string) (dbType, error) {
	slog.Debug("Determining DB type", "dbIdentifier", dbIdentifier)

	isAurora, err := r.isAuroraCluster(dbIdentifier)

	if err != nil {
		return "", fmt.Errorf("failed to check if Aurora cluster: %w", err)
	}

	if isAurora {
		slog.Debug("Identified as Aurora cluster")

		return dbTypeAurora, nil
	}

	isRDS, err := r.isRDSInstance(dbIdentifier)

	if err != nil {
		return "", fmt.Errorf("failed to check if RDS instance: %w", err)
	}

	if isRDS {
		slog.Debug("Identified as RDS instance")

		return dbTypeRDS, nil
	}

	slog.Debug("Unable to determine DB type")

	return "", fmt.Errorf("database '%s' was not found as either Aurora cluster or RDS instance", dbIdentifier)
}

/*
isAuroraCluster checks if the DB identifier is an Aurora cluster.
*/
func (r *RDS) isAuroraCluster(dbIdentifier string) (bool, error) {
	slog.Debug("Checking if DB is Aurora cluster")

	ctx := context.Background()

	output, err := r.factory.GetClient().DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbIdentifier),
	})

	if err != nil {
		if strings.Contains(err.Error(), "DBClusterNotFoundFault") {
			slog.Debug("DB cluster not found")

			return false, nil
		}

		return false, fmt.Errorf("failed to execute DescribeDBClusters API: %w", err)
	}

	if len(output.DBClusters) == 0 {
		slog.Debug("DB cluster not found")

		return false, nil
	}

	slog.Debug("DB cluster found")

	isAurora, err := isAuroraEngine(
		aws.ToString(output.DBClusters[0].Engine),
	)

	if err != nil {
		return false, fmt.Errorf("failed to determine if engine is Aurora: %w", err)
	}

	return isAurora, nil
}

/*
isRDSInstance checks if the DB identifier is an RDS instance.
*/
func (r *RDS) isRDSInstance(dbIdentifier string) (bool, error) {
	slog.Debug("Checking if DB is RDS instance")

	ctx := context.Background()

	output, err := r.factory.GetClient().DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
	})

	if err != nil {
		if strings.Contains(err.Error(), "DBInstanceNotFound") {
			slog.Debug("DB instance not found")

			return false, nil
		}

		return false, fmt.Errorf("failed to execute DescribeDBInstances API: %w", err)
	}

	if len(output.DBInstances) == 0 {
		slog.Debug("DB instance not found")

		return false, nil
	}

	slog.Debug("DB instance found")

	isAurora, err := isAuroraEngine(
		aws.ToString(output.DBInstances[0].Engine),
	)

	if err != nil {
		return false, fmt.Errorf("failed to determine if engine is Aurora: %w", err)
	}

	return !isAurora, nil
}
