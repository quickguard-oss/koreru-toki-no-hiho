package ktnh

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
	apprds "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/rds"
)

func Test_Template(t *testing.T) {
	testCases := []struct {
		name         string
		dbIdentifier string
		mockSetup    func(*appmock.MockRDSFactory, *appmock.MockRDSClient)
		wantErr      bool
	}{
		{
			name:         "Aurora",
			dbIdentifier: "db-1",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-1"),
				}

				result1 := &rds.DescribeDBClustersOutput{
					DBClusters: []types.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, nil)
			},
			wantErr: false,
		},
		{
			name:         "RDS",
			dbIdentifier: "db-2",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-2"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, fmt.Errorf("DBClusterNotFoundFault"))

				params2 := &rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: aws.String("db-2"),
				}

				result2 := &rds.DescribeDBInstancesOutput{
					DBInstances: []types.DBInstance{
						{
							Engine: aws.String("mysql"),
						},
					},
				}

				c.On("DescribeDBInstances", mock.Anything, params2, mock.Anything).
					Return(result2, nil)
			},
			wantErr: false,
		},
		{
			name:         "Error during determining DB type",
			dbIdentifier: "db-3",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-3"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockRDSFactory)
			mockClient := new(appmock.MockRDSClient)

			tc.mockSetup(mockFactory, mockClient)

			k := &ktnh{
				dbIdentifier:      tc.dbIdentifier,
				dbIdentifierShort: shortenIdentifier(tc.dbIdentifier),
				rds:               apprds.NewRDS(mockFactory),
			}

			templateBody, qualifier, err := k.Template()

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				var parsedYaml any

				yamlErr := yaml.Unmarshal([]byte(templateBody), &parsedYaml)

				assert.NoError(t, yamlErr, "Template should be parsable as YAML")
				assert.NotNil(t, parsedYaml, "Template must not be nil")

				assert.Regexp(t, "^[A-Za-z0-9]{6}$", qualifier, "Qualifier should only contain alphanumeric characters and be exactly 6 characters long")
			}

			mockFactory.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}
