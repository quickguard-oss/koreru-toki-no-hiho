/*
Package ktnh implements the core functionality of the ktnh command-line tool,
serving as the main engine for managing Aurora clusters and RDS instances in AWS.

The package handles all CloudFormation stack operations including creating, listing,
and deleting stacks.
It serves as the bridge between the CLI interface and the AWS resources, providing
all the necessary business logic for the ktnh command.
*/
package ktnh

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/cfn"
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/rds"
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/utils"
)

/*
ktnh is the core component of the ktnh command that manages the lifecycle of
Aurora clusters and RDS instances.
*/
type ktnh struct {
	dbIdentifier      string              // DB cluster/instance identifier
	dbIdentifierShort string              // shortened DB identifier for display
	stackNamePrefix   string              // prefix for CloudFormation stack name
	cfn               *cfn.CloudFormation // CloudFormation operations wrapper
	rds               *rds.RDS            // RDS operations wrapper
}

/*
stackNameOption provides configuration for generating CloudFormation stack names.
It contains the shortened database identifier and a qualifier string that together
form a unique stack name when combined with the stack name prefix.
If either field is an empty string, a wildcard pattern (".+") will be used instead,
which can be useful for matching multiple stacks in search operations.
*/
type stackNameOption struct {
	dbIdentifierShort string // shortened DB identifier for display
	qualifier         string // unique qualifier for the CloudFormation stack
}

/*
dbIdentifierTruncateLength defines the maximum length for displaying DB identifiers.
*/
const dbIdentifierTruncateLength = 10

/*
NewKtnh creates and returns a new instance of ktnh.
*/
func NewKtnh(dbIdentifier string, stackNamePrefix string) (*ktnh, error) {
	cfnFactory, err := awsfactory.NewCloudFormationFactory()

	if err != nil {
		return nil, fmt.Errorf("failed to create CloudFormation factory: %w", err)
	}

	rdsFactory, err := awsfactory.NewRDSFactory()

	if err != nil {
		return nil, fmt.Errorf("failed to create RDS factory: %w", err)
	}

	return &ktnh{
		dbIdentifier:      dbIdentifier,
		dbIdentifierShort: shortenIdentifier(dbIdentifier),
		stackNamePrefix:   stackNamePrefix,
		cfn:               cfn.NewCloudFormation(cfnFactory),
		rds:               rds.NewRDS(rdsFactory),
	}, nil
}

/*
shortenIdentifier shortens the DB identifier by truncating it to the specified length.
If the last character after truncation is not alphanumeric, it extends the length by one
to increase the chance of ending with an alphanumeric character.
*/
func shortenIdentifier(dbIdentifier string) string {
	shortened := utils.Truncate(dbIdentifier, dbIdentifierTruncateLength)

	if shortened == "" {
		return shortened
	}

	if len(shortened) == len(dbIdentifier) {
		return shortened
	}

	isAlphanumeric, _ := regexp.MatchString("[A-Za-z0-9]$", shortened)

	if isAlphanumeric {
		return shortened
	}

	return utils.Truncate(dbIdentifier, dbIdentifierTruncateLength+1)
}

/*
generateStackName generates a stack name using the prefix, DB identifier, and qualifier.
*/
func (k *ktnh) generateStackName(option *stackNameOption) string {
	items := []string{k.stackNamePrefix}

	if option.dbIdentifierShort == "" {
		items = append(items, ".+")
	} else {
		items = append(items, option.dbIdentifierShort)

		if option.qualifier == "" {
			items = append(items, ".+")
		} else {
			items = append(items, option.qualifier)
		}
	}

	stackName := strings.Join(items, "-")

	slog.Debug("Generated stack name", "stackName", stackName)

	return stackName
}

/*
findMatchingStack finds the CloudFormation stack matching the DB identifier.
Returns the stack name, whether a stack was found, and any error encountered.
*/
func (k *ktnh) findMatchingStack() (string, bool, error) {
	slog.Debug("Finding matching stack")

	dbType, err := k.rds.DetermineDBType(k.dbIdentifier)

	if err != nil {
		return "", false, fmt.Errorf("failed to determine DB type: %w", err)
	}

	pattern := fmt.Sprintf(
		"^%s$",
		k.generateStackName(&stackNameOption{
			dbIdentifierShort: k.dbIdentifierShort,
		}),
	)

	slog.Debug("Generated stack name pattern for matching", "pattern", pattern)

	re, err := regexp.Compile(pattern)

	if err != nil {
		return "", false, fmt.Errorf("failed to compile regex pattern '%s': %w", pattern, err)
	}

	verifyOotion := cfn.MetadataVerifyOption{
		DBIdentifier: k.dbIdentifier,
		DBType:       string(dbType),
	}

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

		isMatched, err := cfn.VerifyMetadata(metadata, &verifyOotion)

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

		return true
	}

	stacks, err := k.cfn.ListStacks(evaluator)

	if err != nil {
		return "", false, fmt.Errorf("failed to list CloudFormation stacks: %w", err)
	}

	stacksCount := len(stacks)

	slog.Debug("Found stacks matching criteria", "count", stacksCount)

	if stacksCount == 0 {
		return "", false, nil
	} else if 2 <= stacksCount {
		return "", false, fmt.Errorf("multiple stacks found for DB identifier")
	}

	stackName := stacks[0]

	slog.Debug("Found single matching stack", "stackName", stackName)

	return stackName, true, nil
}
