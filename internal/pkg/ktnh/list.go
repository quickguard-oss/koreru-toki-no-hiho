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
	dbIdentifier string // DB cluster/instance identifier
	dbType       string // type of the DB (see `internal/pkg/rds`)
	stackName    string // CloudFormation stack name
}

/*
Constants for list command table headers
*/
const (
	headerDBIdentifier = "ID"
	headerDBType       = "TYPE"
	headerStackName    = "STACK"
)

/*
formatDatabaseInfoForDisplay formats the databases information for display purposes.
Each database's information is presented in ASCII table format for user-friendly console output.
It returns a slice of strings where each string represents a row in the table.
*/
func formatDatabaseInfoForDisplay(databases []displayDBInfo) []string {
	lines := []string{}

	slog.Debug("Formatting databases information for display output")

	if len(databases) == 0 {
		slog.Debug("No managed databases found")

		return lines
	}

	dbIdentifierMaxWidth := len(headerDBIdentifier)
	dbTypeMaxWidth := len(headerDBType)

	for _, db := range databases {
		if dbIdentifierMaxWidth < len(db.dbIdentifier) {
			dbIdentifierMaxWidth = len(db.dbIdentifier)
		}

		if dbTypeMaxWidth < len(db.dbType) {
			dbTypeMaxWidth = len(db.dbType)
		}
	}

	// NOTE: Create format string like "%-10s" for left-aligned display
	rowFormat := fmt.Sprintf("%%-%ds   %%-%ds   %%s", dbIdentifierMaxWidth, dbTypeMaxWidth)

	lines = append(lines, fmt.Sprintf(rowFormat, headerDBIdentifier, headerDBType, headerStackName))

	for _, db := range databases {
		lines = append(lines, fmt.Sprintf(rowFormat, db.dbIdentifier, db.dbType, db.stackName))
	}

	slog.Debug("Formatted databases information for output")

	return lines
}

/*
List returns a list of managed databases in a human-readable format.
*/
func (s *ktnh) List() ([]string, error) {
	databases, err := s.collectManagedDatabases()

	if err != nil {
		return nil, fmt.Errorf("failed to collect managed databases: %w", err)
	}

	return formatDatabaseInfoForDisplay(databases), nil
}

/*
collectManagedDatabases finds all databases managed by ktnh.
*/
func (s *ktnh) collectManagedDatabases() ([]displayDBInfo, error) {
	slog.Debug("Finding all managed databases")

	pattern := fmt.Sprintf(
		"^%s$",
		s.generateStackName(&stackNameOption{}),
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

		metadata, err := s.cfn.GetKTNHMetadata(stackName)

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

	_, err = s.cfn.ListStacks(evaluator)

	if err != nil {
		return nil, fmt.Errorf("failed to list CloudFormation stacks: %w", err)
	}

	slog.Debug("Found managed databases", "count", len(databases))

	return databases, nil
}
