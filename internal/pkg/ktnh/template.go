package ktnh

import (
	"fmt"
	"log/slog"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/cfn"
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/utils"
)

/*
qualifierLength defines the length of qualifier string.
A qualifier is a unique identifier for each CloudFormation stack.
*/
const qualifierLength = 6

/*
generateQualifier generates a unique qualifier for the CloudFormation stack.
*/
func generateQualifier() string {
	qualifier := utils.Truncate(utils.GenerateRandomStr(), qualifierLength)

	slog.Debug("Generated qualifier for CloudFormation stack", "qualifier", qualifier)

	return qualifier
}

/*
Template generates a CloudFormation template.
*/
func (s *ktnh) Template() (templateBody string, qualifier string, err error) {
	dbType, err := s.rds.DetermineDBType(s.dbIdentifier)

	if err != nil {
		return "", "", fmt.Errorf("failed to determine DB type: %w", err)
	}

	qualifier = generateQualifier()

	templateBody, err = cfn.GenerateTemplateBody(s.dbIdentifier, s.dbIdentifierShort, string(dbType), qualifier)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate CloudFormation template: %w", err)
	}

	return templateBody, qualifier, nil
}
