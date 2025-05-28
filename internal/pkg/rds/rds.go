/*
Package rds provides functionality for interacting with Amazon RDS services.

This package offers utilities to determine database types (Aurora or RDS),
and interact with these services through the AWS SDK. It simplifies common
operations like identifying database instances and clusters, checking engine
types, and abstracting away the complexities of the underlying AWS API calls.
*/
package rds

import (
	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/awsfactory"
)

/*
RDS handles interactions with the Amazon RDS service.
*/
type RDS struct {
	factory awsfactory.RDSFactory // Interface instead of concrete client
}

/*
NewRDS creates and returns a new instance of RDS.
*/
func NewRDS(factory awsfactory.RDSFactory) *RDS {
	return &RDS{
		factory: factory,
	}
}
