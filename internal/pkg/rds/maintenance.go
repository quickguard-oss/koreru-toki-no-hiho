package rds

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

/*
parseARN extracts the DB identifier and type from an ARN.
Returns the DB identifier and type ("cluster" or "db").
*/
func parseARN(arn string) (dbIdentifier string, dbType string) {
	parts := strings.Split(arn, ":")

	lastIndex := len(parts) - 1

	dbIdentifier = parts[lastIndex]
	dbType = parts[lastIndex-1]

	return
}

/*
GetPendingMaintenanceActions checks if Aurora clusters and RDS instances have pending maintenance actions.
It accepts three parameters:
- clusters: a slice of Aurora cluster IDs
- instances: a slice of standalone RDS instance IDs
- clusterMembers: a map where keys are cluster IDs and values are slices of member instance IDs

It returns a map where the key is in the format "${dbType}:${dbIdentifier}" (e.g., "cluster:my-cluster" or "db:my-instance")
and the value is a boolean indicating whether there are any pending maintenance actions.
*/
func (r *RDS) GetPendingMaintenanceActions(clusters []string, instances []string, clusterMembers map[string][]string) (map[string]bool, error) {
	result := map[string]bool{}

	if len(clusters)+len(instances) == 0 {
		return result, nil
	}

	slog.Debug("Starting to check for pending maintenance actions")

	clusterMemberInstanceIds := []string{}

	instanceToCluster := map[string]string{}

	for clusterId, members := range clusterMembers {
		clusterMemberInstanceIds = slices.Concat(clusterMemberInstanceIds, members)

		for _, memberId := range members {
			instanceToCluster[memberId] = clusterId
		}
	}

	// NOTE: Sort the slice to ensure a consistent order for testing purposes
	//       This eliminates non-deterministic behavior caused by map iteration.
	slices.Sort(clusterMemberInstanceIds)

	paginator, err := r.factory.NewDescribePendingMaintenanceActionsPaginator(&rds.DescribePendingMaintenanceActionsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: clusters,
			},
			{
				Name:   aws.String("db-instance-id"),
				Values: slices.Concat(instances, clusterMemberInstanceIds),
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create DescribePendingMaintenanceActions paginator: %w", err)
	}

	ctx := context.Background()

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to execute DescribePendingMaintenanceActions API: %w", err)
		}

		for _, action := range output.PendingMaintenanceActions {
			dbIdentifier, dbType := parseARN(
				aws.ToString(action.ResourceIdentifier),
			)

			if dbType == "cluster" {
				result["cluster:"+dbIdentifier] = true
			} else {
				clusterId, isClusterMember := instanceToCluster[dbIdentifier]

				if isClusterMember {
					result["cluster:"+clusterId] = true
				} else {
					result["db:"+dbIdentifier] = true
				}
			}
		}
	}

	slog.Debug("Checking pending maintenance actions successful")

	return result, nil
}
