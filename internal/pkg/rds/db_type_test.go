package rds

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_DetermineDBType(t *testing.T) {
	testCases := []struct {
		name         string
		dbIdentifier string
		mockSetup    func(*appmock.MockRDSFactory, *appmock.MockRDSClient)
		expected     dbType
		wantErr      bool
	}{
		{
			name:         "Aurora cluster",
			dbIdentifier: "aurora-cluster-db",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("aurora-cluster-db"),
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
			expected: dbTypeAurora,
			wantErr:  false,
		},
		{
			name:         "RDS instance",
			dbIdentifier: "rds-instance-db",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("rds-instance-db"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, fmt.Errorf("DBClusterNotFoundFault"))

				params2 := &rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: aws.String("rds-instance-db"),
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
			expected: dbTypeRDS,
			wantErr:  false,
		},
		{
			name:         "Not found",
			dbIdentifier: "not-found-db",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("not-found-db"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, fmt.Errorf("DBClusterNotFoundFault"))

				params2 := &rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: aws.String("not-found-db"),
				}

				result2 := &rds.DescribeDBInstancesOutput{}

				c.On("DescribeDBInstances", mock.Anything, params2, mock.Anything).
					Return(result2, fmt.Errorf("DBInstanceNotFound"))
			},
			expected: "",
			wantErr:  true,
		},
		{
			name:         "Error - DescribeDBClusters",
			dbIdentifier: "cluster-error-db",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("cluster-error-db"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, assert.AnError)
			},
			expected: "",
			wantErr:  true,
		},
		{
			name:         "Error - DescribeDBInstances",
			dbIdentifier: "instance-error-db",
			mockSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params1 := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("instance-error-db"),
				}

				result1 := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params1, mock.Anything).
					Return(result1, fmt.Errorf("DBClusterNotFoundFault"))

				params2 := &rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: aws.String("instance-error-db"),
				}

				result2 := &rds.DescribeDBInstancesOutput{}

				c.On("DescribeDBInstances", mock.Anything, params2, mock.Anything).
					Return(result2, assert.AnError)
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockRDSFactory)
			mockClient := new(appmock.MockRDSClient)

			tc.mockSetup(mockFactory, mockClient)

			r := NewRDS(mockFactory)

			got, err := r.DetermineDBType(tc.dbIdentifier)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "DB type detection result does not match expected value")
			}

			mockFactory.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}
