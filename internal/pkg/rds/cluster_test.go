package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_GetClusterMembers(t *testing.T) {
	testCases := []struct {
		name      string
		clusters  []string
		mockSetup func(*appmock.MockRDSFactory, *appmock.MockDescribeDBClustersPaginator)
		expected  map[string][]string
		wantErr   bool
	}{
		{
			name: "Clusters",
			clusters: []string{
				"cluster-1",
				"cluster-2",
				"cluster-3",
			},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
								"cluster-2",
								"cluster-3",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribeDBClustersOutput{
					DBClusters: []types.DBCluster{
						{
							DBClusterIdentifier: aws.String("cluster-1"),
							DBClusterMembers: []types.DBClusterMember{
								{
									DBInstanceIdentifier: aws.String("instance-1a"),
								},
								{
									DBInstanceIdentifier: aws.String("instance-1b"),
								},
							},
						},
						{
							DBClusterIdentifier: aws.String("cluster-2"),
							DBClusterMembers: []types.DBClusterMember{
								{
									DBInstanceIdentifier: aws.String("instance-2a"),
								},
							},
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result1, nil).
					Once()

				p.On("HasMorePages").
					Return(true).
					Once()

				result2 := &rds.DescribeDBClustersOutput{
					DBClusters: []types.DBCluster{
						{
							DBClusterIdentifier: aws.String("cluster-3"),
							DBClusterMembers: []types.DBClusterMember{
								{
									DBInstanceIdentifier: aws.String("instance-3a"),
								},
							},
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result2, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: map[string][]string{
				"cluster-1": {
					"instance-1a",
					"instance-1b",
				},
				"cluster-2": {
					"instance-2a",
				},
				"cluster-3": {
					"instance-3a",
				},
			},
			wantErr: false,
		},
		{
			name:      "Empty",
			clusters:  []string{},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {},
			expected:  map[string][]string{},
			wantErr:   false,
		},
		{
			name: "Paginator creation error",
			clusters: []string{
				"cluster-1",
			},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(nil, assert.AnError)
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "NextPage error",
			clusters: []string{
				"cluster-1",
			},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribeDBClustersOutput{}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result1, assert.AnError).
					Once()
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockRDSFactory)
			mockPaginator := new(appmock.MockDescribeDBClustersPaginator)

			tc.mockSetup(mockFactory, mockPaginator)

			r := NewRDS(mockFactory)

			got, err := r.GetClusterMembers(tc.clusters)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Cluster members mapping does not match expected result")
			}

			mockFactory.AssertExpectations(t)
			mockPaginator.AssertExpectations(t)
		})
	}
}
