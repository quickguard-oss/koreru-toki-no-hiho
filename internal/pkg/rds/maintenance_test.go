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

func Test_parseARN(t *testing.T) {
	testCases := []struct {
		name               string
		arn                string
		expectedIdentifier string
		expectedType       string
	}{
		{
			name:               "Aurora cluster ARN",
			arn:                "arn:aws:rds:ap-northeast-1:123456789012:cluster:my-aurora-cluster",
			expectedIdentifier: "my-aurora-cluster",
			expectedType:       "cluster",
		},
		{
			name:               "RDS instance ARN",
			arn:                "arn:aws:rds:ap-northeast-1:123456789012:db:my-rds-instance",
			expectedIdentifier: "my-rds-instance",
			expectedType:       "db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotIdentifier, gotType := parseARN(tc.arn)

			assert.Equal(t, tc.expectedIdentifier, gotIdentifier, "Resource identifier should match the expected value")
			assert.Equal(t, tc.expectedType, gotType, "Resource type should match the expected value")
		})
	}
}

func Test_GetPendingMaintenanceActions(t *testing.T) {
	testCases := []struct {
		name           string
		clusters       []string
		instances      []string
		clusterMembers map[string][]string
		mockSetup      func(*appmock.MockRDSFactory, *appmock.MockDescribePendingMaintenanceActionsPaginator)
		expected       map[string]bool
		wantErr        bool
	}{
		{
			name: "Clusters and instances",
			clusters: []string{
				"parent-cluster-1",
				"parent-cluster-2",
				"parent-cluster-3",
				"parent-cluster-4",
			},
			instances: []string{
				"standalone-instance-1",
				"standalone-instance-2",
				"standalone-instance-3",
			},
			clusterMembers: map[string][]string{
				"parent-cluster-1": {
					"member-instance-1a",
					"member-instance-1b",
				},
				"parent-cluster-2": {
					"member-instance-2a",
				},
				"parent-cluster-3": {
					"member-instance-3a",
					"member-instance-3b",
				},
				"parent-cluster-4": {
					"member-instance-4a",
				},
			},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"parent-cluster-1",
								"parent-cluster-2",
								"parent-cluster-3",
								"parent-cluster-4",
							},
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"standalone-instance-1",
								"standalone-instance-2",
								"standalone-instance-3",
								"member-instance-1a",
								"member-instance-1b",
								"member-instance-2a",
								"member-instance-3a",
								"member-instance-3b",
								"member-instance-4a",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []types.ResourcePendingMaintenanceActions{
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:cluster:parent-cluster-1"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("system-update"),
									Description: aws.String("a"),
								},
								{
									Action:      aws.String("hardware-maintenance"),
									Description: aws.String("b"),
								},
							},
						},
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:db:member-instance-3b"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("os-upgrade"),
									Description: aws.String("c"),
								},
							},
						},
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:db:standalone-instance-2"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("ca-certificate-rotation"),
									Description: aws.String("d"),
								},
							},
						},
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:db:standalone-instance-3"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("db-upgrade"),
									Description: aws.String("e"),
								},
								{
									Action:      aws.String("os-upgrade"),
									Description: aws.String("f"),
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

				result2 := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []types.ResourcePendingMaintenanceActions{
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:cluster:parent-cluster-4"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("system-update"),
									Description: aws.String("g"),
								},
							},
						},
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:db:member-instance-4a"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("ca-certificate-rotation"),
									Description: aws.String("h"),
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
			expected: map[string]bool{
				"cluster:parent-cluster-1": true,
				"cluster:parent-cluster-3": true,
				"db:standalone-instance-2": true,
				"db:standalone-instance-3": true,
				"cluster:parent-cluster-4": true,
			},
			wantErr: false,
		},
		{
			name: "Clusters only",
			clusters: []string{
				"cluster-1",
				"cluster-2",
			},
			instances:      []string{},
			clusterMembers: map[string][]string{},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
								"cluster-2",
							},
						},
						{
							Name:   aws.String("db-instance-id"),
							Values: nil, // NOTE: nil because `slices.Concat` returns nil for empty results, not an empty slice
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []types.ResourcePendingMaintenanceActions{
						{
							ResourceIdentifier: aws.String("arn:aws:rds:us-west-2:123456789012:cluster:cluster-1"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("system-update"),
									Description: aws.String("a"),
								},
							},
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result1, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: map[string]bool{
				"cluster:cluster-1": true,
			},
			wantErr: false,
		},
		{
			name:     "Instances only",
			clusters: []string{},
			instances: []string{
				"instance-1",
				"instance-2",
			},
			clusterMembers: map[string][]string{},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("db-cluster-id"),
							Values: []string{},
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"instance-1",
								"instance-2",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []types.ResourcePendingMaintenanceActions{
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:db:instance-1"),
							PendingMaintenanceActionDetails: []types.PendingMaintenanceAction{
								{
									Action:      aws.String("os-upgrade"),
									Description: aws.String("a"),
								},
							},
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result1, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: map[string]bool{
				"db:instance-1": true,
			},
			wantErr: false,
		},
		{
			name:           "Empty",
			clusters:       []string{},
			instances:      []string{},
			clusterMembers: map[string][]string{},
			mockSetup:      func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {},
			expected:       map[string]bool{},
			wantErr:        false,
		},
		{
			name: "Paginator creation error",
			clusters: []string{
				"cluster-1",
			},
			instances: []string{
				"instance-1",
			},
			clusterMembers: map[string][]string{},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
							},
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"instance-1",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
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
			instances: []string{
				"instance-1",
			},
			clusterMembers: map[string][]string{},
			mockSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []types.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"cluster-1",
							},
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"instance-1",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &rds.DescribePendingMaintenanceActionsOutput{}

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
			mockPaginator := new(appmock.MockDescribePendingMaintenanceActionsPaginator)

			tc.mockSetup(mockFactory, mockPaginator)

			r := NewRDS(mockFactory)

			got, err := r.GetPendingMaintenanceActions(tc.clusters, tc.instances, tc.clusterMembers)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Pending maintenance actions map should match expected values")
			}

			mockFactory.AssertExpectations(t)
			mockPaginator.AssertExpectations(t)
		})
	}
}
