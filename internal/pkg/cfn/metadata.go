package cfn

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"gopkg.in/yaml.v3"
)

/*
MetadataVerifyOption defines options for metadata verification.
*/
type MetadataVerifyOption struct {
	DBIdentifier string // DB cluster/instance identifier
	DBType       string // type of the DB (see `internal/pkg/rds`)
}

/*
ktnhMetadata defines the structure of the `Metadata.KTNH` section in CloudFormation templates.
*/
type ktnhMetadata struct {
	Generator    string `yaml:"Generator"`    // generator name
	Version      string `yaml:"Version"`      // version of the generator
	DBIdentifier string `yaml:"DBIdentifier"` // DB cluster/instance identifier
	DBType       string `yaml:"DBType"`       // type of the DB (see `internal/pkg/rds`)
}

/*
cloudFormationTemplate defines the structure of CloudFormation templates.
*/
type cloudFormationTemplate struct {
	Metadata struct {
		KTNH ktnhMetadata `yaml:"KTNH"` // `Metadata.KTNH` section
	} `yaml:"Metadata"`
}

/*
VerifyMetadata veryfies the metadata of a CloudFormation template.
*/
func VerifyMetadata(metadata *ktnhMetadata, option *MetadataVerifyOption) (bool, error) {
	slog.Debug("Verifying template metadata")

	err := validateRequiredMetadataFields(metadata)

	if err != nil {
		return false, fmt.Errorf("invalid metadata: %w", err)
	}

	// TODO: Verify version as well. (What rules should be applied?)

	if metadata.Generator != generatorName {
		slog.Debug("Generator name mismatch",
			"expected", generatorName,
			"actual", metadata.Generator,
		)

		return false, nil
	}

	if (option.DBIdentifier != "") && (metadata.DBIdentifier != option.DBIdentifier) {
		slog.Debug("DB identifier mismatch",
			"expected", option.DBIdentifier,
			"actual", metadata.DBIdentifier,
		)

		return false, nil
	}

	if (option.DBType != "") && (metadata.DBType != option.DBType) {
		slog.Debug("DB type mismatch",
			"expected", option.DBType,
			"actual", metadata.DBType,
		)

		return false, nil
	}

	slog.Debug("Metadata verification successful")

	return true, nil
}

/*
validateRequiredMetadataFields checks that all required metadata fields are present and non-empty.
*/
func validateRequiredMetadataFields(metadata *ktnhMetadata) error {
	slog.Debug("Validating required metadata fields")

	if metadata.Generator == "" {
		return fmt.Errorf("metadata field 'Generator' is empty")
	}

	if metadata.Version == "" {
		return fmt.Errorf("metadata field 'Version' is empty")
	}

	if metadata.DBIdentifier == "" {
		return fmt.Errorf("metadata field 'DBIdentifier' is empty")
	}

	if metadata.DBType == "" {
		return fmt.Errorf("metadata field 'DBType' is empty")
	}

	slog.Debug("All required metadata fields are present")

	return nil
}

/*
parseTemplate parses the CloudFormation template.
*/
func parseTemplate(templateBody string) (*cloudFormationTemplate, error) {
	var template cloudFormationTemplate

	slog.Debug("Parsing template body")

	err := yaml.Unmarshal([]byte(templateBody), &template)

	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	slog.Debug("Template body parsed successfully")

	return &template, nil
}

/*
GetKTNHMetadata retrieves the metadata from a CloudFormation stack.
*/
func (c *CloudFormation) GetKTNHMetadata(stackName string) (*ktnhMetadata, error) {
	slog.Debug("Retrieving metadata from CloudFormation stack")

	templateBody, err := c.getStackTemplate(stackName)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve template: %w", err)
	}

	template, err := parseTemplate(templateBody)

	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata from template: %w", err)
	}

	slog.Debug("Metadata retrieved successfully")

	slog.Debug("Extracted metadata from template",
		"Generator", template.Metadata.KTNH.Generator,
		"Version", template.Metadata.KTNH.Version,
		"DBIdentifier", template.Metadata.KTNH.DBIdentifier,
		"DBType", template.Metadata.KTNH.DBType,
	)

	return &template.Metadata.KTNH, nil
}

/*
getStackTemplate retrieves the template body of a given stack.
*/
func (c *CloudFormation) getStackTemplate(stackName string) (string, error) {
	slog.Debug("Retrieving stack template", "stackName", stackName)

	ctx := context.Background()

	output, err := c.factory.GetClient().GetTemplate(ctx, &cloudformation.GetTemplateInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute GetTemplate API for stack '%s': %w", stackName, err)
	}

	slog.Debug("Stack template retrieved successfully")

	return aws.ToString(output.TemplateBody), nil
}
