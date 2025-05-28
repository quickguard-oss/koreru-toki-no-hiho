/*
Package cfn provides functionality for interacting with AWS CloudFormation.

It includes utilities for template generation, template data management,
and CloudFormation stack operations. This package is designed to simplify
the process of creating and managing CloudFormation stacks.
*/
package cfn

import (
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
)

/*
CloudFormation provides functionality for interacting with AWS CloudFormation service.
It handles template generation, stack operations, and other CloudFormation-related tasks.
*/
type CloudFormation struct {
	factory awsfactory.CloudFormationFactory // Interface instead of concrete client
}

const (
	generatorName    = "koreru-toki-no-hiho" // name of the generator
	generatorVersion = "1"                   // current version of the generator
)

/*
NewCloudFormation creates and returns a new instance of CloudFormation.
*/
func NewCloudFormation(factory awsfactory.CloudFormationFactory) *CloudFormation {
	return &CloudFormation{
		factory: factory,
	}
}
