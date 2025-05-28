/*
Package awsfactory provides a factory for creating AWS service clients.

It uses the AWS SDK for Go v2 to create clients for services like RDS and CloudFormation.
*/
package awsfactory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	// cfg holds the AWS configuration used by all AWS service clients
	cfg aws.Config

	// once ensures that the AWS configuration is loaded only once
	once sync.Once

	// counter counts how many times the AWS configuration has been loaded
	counter int
)

/*
loadAWSConfig loads the AWS configuration.
It uses a sync.Once to ensure the configuration is loaded only once.
*/
func loadAWSConfig() error {
	var err error

	ctx := context.Background()

	once.Do(func() {
		slog.Debug("Loading AWS configuration")

		counter++

		cfg, err = config.LoadDefaultConfig(ctx)

		if err != nil {
			return
		}

		slog.Debug("AWS configuration loaded successfully")
	})

	return err
}

/*
resetConfiguration resets the AWS configuration and initialization flag.
*/
func resetConfiguration() {
	cfg = aws.Config{}

	once = sync.Once{}

	counter = 0
}
