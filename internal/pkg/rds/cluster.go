package rds

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

/*
GetClusterMembers retrieves all DB instances that belong to the given DB clusters.
It returns a map where the key is the cluster ID and the value is a slice of instance IDs.
*/
func (r *RDS) GetClusterMembers(clusters []string) (map[string][]string, error) {
	result := map[string][]string{}

	if len(clusters) == 0 {
		return result, nil
	}

	slog.Debug("Retrieving DB instances that belong to DB clusters")

	paginator, err := r.factory.NewDescribeDBClustersPaginator(&rds.DescribeDBClustersInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: clusters,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create DescribeDBClusters paginator: %w", err)
	}

	ctx := context.Background()

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to execute DescribeDBClusters API: %w", err)
		}

		for _, cluster := range output.DBClusters {
			clusterId := aws.ToString(cluster.DBClusterIdentifier)

			result[clusterId] = make([]string, len(cluster.DBClusterMembers))

			for i, instance := range cluster.DBClusterMembers {
				result[clusterId][i] = aws.ToString(instance.DBInstanceIdentifier)
			}

			slog.Debug("Retrieved DB instances for cluster",
				"cluster", clusterId,
				"instanceCount", len(result[clusterId]),
			)
		}
	}

	slog.Debug("Retrieved DB instances that belong to DB clusters")

	return result, nil
}
