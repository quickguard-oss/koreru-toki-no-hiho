package ktnh

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/cfn"
)

/*
displayDBInfo is a structure used for displaying managed database information.
*/
type displayDBInfo struct {
	dbIdentifier   string // DB cluster/instance identifier
	dbType         string // type of the DB (see `internal/pkg/rds`)
	stackName      string // CloudFormation stack name
	hasMaintenance bool   // whether there are pending maintenance actions
}

/*
convertDBsToStringRows transforms database information into a string slice.
It returns a header slice containing column names and a 2D slice
where each inner slice represents a row of database information.
*/
func convertDBsToStringRows(databases []displayDBInfo, isShowMaintenance bool) ([]string, [][]string) {
	slog.Debug("Converting databases information to string rows")

	body := make([][]string, len(databases))

	for i, db := range databases {
		var maintenanceStatus string

		if isShowMaintenance {
			if db.hasMaintenance {
				maintenanceStatus = "pending"
			} else {
				maintenanceStatus = "none"
			}
		} else {
			maintenanceStatus = "(unknown)"
		}

		body[i] = []string{
			db.dbIdentifier,
			db.dbType,
			db.stackName,
			maintenanceStatus,
		}
	}

	slog.Debug("Converted databases information to string rows")

	return []string{"id", "type", "stack", "maintenance"}, body
}

/*
List returns a list of managed databases.
*/
func (k *ktnh) List() ([]string, [][]string, error) {
	databases, err := k.collectManagedDatabases()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect managed databases: %w", err)
	}

	isShowMaintenance := true

	databasesWithMaintenance, err := k.updateMaintenanceStatus(databases)

	if err != nil {
		slog.Warn("Failed to add maintenance status", "error", err)

		isShowMaintenance = false
	}

	headers, body := convertDBsToStringRows(databasesWithMaintenance, isShowMaintenance)

	return headers, body, nil
}

/*
updateMaintenanceStatus updates the maintenance status for each database.
*/
func (k *ktnh) updateMaintenanceStatus(databases []displayDBInfo) ([]displayDBInfo, error) {
	slog.Debug("Updating maintenance status for databases")

	clusters, instances, clusterMembers, err := k.categorizeDBsByType(databases)

	if err != nil {
		return databases, fmt.Errorf("failed to categorize databases: %w", err)
	}

	pendingMaintenance, err := k.rds.GetPendingMaintenanceActions(clusters, instances, clusterMembers)

	if err != nil {
		return databases, fmt.Errorf("failed to get pending maintenance actions: %w", err)
	}

	databasesWithMaintenance := make([]displayDBInfo, len(databases))

	copy(databasesWithMaintenance, databases)

	for i, db := range databasesWithMaintenance {
		var prefix string

		if db.dbType == "aurora" {
			prefix = "cluster:"
		} else {
			prefix = "db:"
		}

		databasesWithMaintenance[i].hasMaintenance = pendingMaintenance[prefix+db.dbIdentifier]
	}

	slog.Debug("Updated maintenance status for databases")

	return databasesWithMaintenance, nil
}

/*
categorizeDBsByType separates DB identifiers into clusters and instances based on their type.
It returns:
- a slice of Aurora cluster IDs
- a slice of standalone RDS instance IDs
- a map of cluster IDs to their member instance IDs
- an error if any operation fails
*/
func (k *ktnh) categorizeDBsByType(databases []displayDBInfo) (
	clusters []string,
	instances []string,
	clusterMembers map[string][]string,
	err error,
) {
	slog.Debug("Categorizing databases by type")

	for _, db := range databases {
		if db.dbType == "aurora" {
			clusters = append(clusters, db.dbIdentifier)
		} else {
			instances = append(instances, db.dbIdentifier)
		}
	}

	clusterMembers, err = k.rds.GetClusterMembers(clusters)

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	slog.Debug("Categorized databases by type",
		"clusters", len(clusters),
		"instances", len(instances),
	)

	return
}

/*
collectManagedDatabases finds all databases managed by ktnh.
*/
func (k *ktnh) collectManagedDatabases() ([]displayDBInfo, error) {
	slog.Debug("Finding all managed databases")

	pattern := fmt.Sprintf(
		"^%s$",
		k.generateStackName(&stackNameOption{}),
	)

	slog.Debug("Generated stack name pattern for matching", "pattern", pattern)

	re, err := regexp.Compile(pattern)

	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern '%s': %w", pattern, err)
	}

	var databases []displayDBInfo

	verifyOption := cfn.MetadataVerifyOption{}

	evaluator := func(stackName string) bool {
		if !re.MatchString(stackName) {
			slog.Debug("Stack name does not match pattern")

			return false
		}

		metadata, err := k.cfn.GetKTNHMetadata(stackName)

		if err != nil {
			slog.Warn("Failed to retrieve metadata for stack during evaluation",
				"stackName", stackName,
				"error", err,
			)

			return false
		}

		isMatched, err := cfn.VerifyMetadata(metadata, &verifyOption)

		if err != nil {
			slog.Warn("Failed to verify metadata for stack during evaluation",
				"stackName", stackName,
				"error", err,
			)

			return false
		}

		if !isMatched {
			slog.Debug("Stack metadata does not match criteria")

			return false
		}

		databases = append(databases, displayDBInfo{
			dbIdentifier: metadata.DBIdentifier,
			dbType:       metadata.DBType,
			stackName:    stackName,
		})

		return true
	}

	_, err = k.cfn.ListStacks(evaluator)

	if err != nil {
		return nil, fmt.Errorf("failed to list CloudFormation stacks: %w", err)
	}

	slog.Debug("Found managed databases", "count", len(databases))

	return databases, nil
}
