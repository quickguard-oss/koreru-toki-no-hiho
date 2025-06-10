package ktnh

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appcfn "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/cfn"
	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
	apprds "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/rds"
)

func Test_List(t *testing.T) {
	testCases := []struct {
		name                                       string
		stackNamePrefix                            string
		mockListStacksSetup                        func(*appmock.MockCloudFormationFactory, *appmock.MockListStacksPaginator)
		mockGetTemplateSetup                       func(*appmock.MockCloudFormationFactory, *appmock.MockCloudFormationClient)
		mockDescribeDBClustersSetup                func(*appmock.MockRDSFactory, *appmock.MockDescribeDBClustersPaginator)
		mockDescribePendingMaintenanceActionsSetup func(*appmock.MockRDSFactory, *appmock.MockDescribePendingMaintenanceActionsPaginator)
		expected                                   [][]string
		wantErr                                    bool
	}{
		{
			name:            "Multiple databases",
			stackNamePrefix: "A",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("A-db1-abcdef"),
						},
						{
							StackName: aws.String("x-db2-ghijkl"),
						},
						{
							StackName: aws.String("A-db3-mnopqr"),
						},
						{
							StackName: aws.String("A-db4-stuvwx"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)

				params1 := &cloudformation.GetTemplateInput{
					StackName: aws.String("A-db1-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db1'
    DBType: 'aurora'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()

				params3 := &cloudformation.GetTemplateInput{
					StackName: aws.String("A-db3-mnopqr"),
				}

				templateBody3 := `
Metadata:
  KTNH:
    Generator: 'test-generator'
    Version: '1'
    DBIdentifier: 'db3'
    DBType: 'aurora'
`

				result3 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody3),
				}

				c.On("GetTemplate", mock.Anything, params3, mock.Anything).
					Return(result3, nil).
					Once()

				params4 := &cloudformation.GetTemplateInput{
					StackName: aws.String("A-db4-stuvwx"),
				}

				templateBody4 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db4'
    DBType: 'rds'
`

				result4 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody4),
				}

				c.On("GetTemplate", mock.Anything, params4, mock.Anything).
					Return(result4, nil).
					Once()
			},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db1",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribeDBClustersOutput{}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db1",
							},
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"db4",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []rdstypes.ResourcePendingMaintenanceActions{
						{
							ResourceIdentifier: aws.String("arn:aws:rds:ap-northeast-1:123456789012:cluster:db1"),
							PendingMaintenanceActionDetails: []rdstypes.PendingMaintenanceAction{
								{
									Action:      aws.String("system-update"),
									Description: aws.String("a"),
								},
							},
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: [][]string{
				{"db1", "aurora", "A-db1-abcdef", "pending"},
				{"db4", "rds", "A-db4-stuvwx", "none"},
			},
			wantErr: false,
		},
		{
			name:            "No databases",
			stackNamePrefix: "B",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup:                       func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			mockDescribeDBClustersSetup:                func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {},
			expected: [][]string{},
			wantErr:  false,
		},
		{
			name:                        "Invalid stack name",
			stackNamePrefix:             "[invalid-regex",
			mockListStacksSetup:         func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {},
			mockGetTemplateSetup:        func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {},
			expected: nil,
			wantErr:  true,
		},
		{
			name:            "Error during listing stacks",
			stackNamePrefix: "C",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(nil, assert.AnError)
			},
			mockGetTemplateSetup:                       func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			mockDescribeDBClustersSetup:                func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {},
			expected: nil,
			wantErr:  true,
		},
		{
			name:            "Error during retrieving template",
			stackNamePrefix: "D",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("D-db1-abcdef"),
						},
						{
							StackName: aws.String("D-db2-ghijkl"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)

				params1 := &cloudformation.GetTemplateInput{
					StackName: aws.String("D-db1-abcdef"),
				}

				result1 := &cloudformation.GetTemplateOutput{}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, assert.AnError).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("D-db2-ghijkl"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db2'
    DBType: 'aurora'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db2",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribeDBClustersOutput{}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db2",
							},
						},
						{
							Name:   aws.String("db-instance-id"),
							Values: nil,
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []rdstypes.ResourcePendingMaintenanceActions{},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: [][]string{
				{"db2", "aurora", "D-db2-ghijkl", "none"},
			},
			wantErr: false,
		},
		{
			name:            "Error during metadata verification",
			stackNamePrefix: "E",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("E-db1-abcdef"),
						},
						{
							StackName: aws.String("E-db2-ghijkl"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)

				params1 := &cloudformation.GetTemplateInput{
					StackName: aws.String("E-db1-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db1'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("E-db2-ghijkl"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db2'
    DBType: 'aurora'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db2",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribeDBClustersOutput{}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db2",
							},
						},
						{
							Name:   aws.String("db-instance-id"),
							Values: nil,
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &rds.DescribePendingMaintenanceActionsOutput{
					PendingMaintenanceActions: []rdstypes.ResourcePendingMaintenanceActions{},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			expected: [][]string{
				{"db2", "aurora", "E-db2-ghijkl", "none"},
			},
			wantErr: false,
		},
		{
			name:            "Error during retrieving DB cluster members",
			stackNamePrefix: "F",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("F-db1-abcdef"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)

				params1 := &cloudformation.GetTemplateInput{
					StackName: aws.String("F-db1-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db1'
    DBType: 'aurora'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()
			},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {
				params := &rds.DescribeDBClustersInput{
					Filters: []rdstypes.Filter{
						{
							Name: aws.String("db-cluster-id"),
							Values: []string{
								"db1",
							},
						},
					},
				}

				f.On("NewDescribeDBClustersPaginator", params).
					Return(nil, assert.AnError)
			},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {},
			expected: [][]string{
				{"db1", "aurora", "F-db1-abcdef", "(unknown)"},
			},
			wantErr: false,
		},
		{
			name:            "Error during retrieving pending maintenance actions",
			stackNamePrefix: "G",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("G-db1-abcdef"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)

				params1 := &cloudformation.GetTemplateInput{
					StackName: aws.String("G-db1-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db1'
    DBType: 'rds'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()
			},
			mockDescribeDBClustersSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribeDBClustersPaginator) {},
			mockDescribePendingMaintenanceActionsSetup: func(f *appmock.MockRDSFactory, p *appmock.MockDescribePendingMaintenanceActionsPaginator) {
				params := &rds.DescribePendingMaintenanceActionsInput{
					Filters: []rdstypes.Filter{
						{
							Name:   aws.String("db-cluster-id"),
							Values: nil, // NOTE: in this path, it's passed as a nil slice, not an empty slice
						},
						{
							Name: aws.String("db-instance-id"),
							Values: []string{
								"db1",
							},
						},
					},
				}

				f.On("NewDescribePendingMaintenanceActionsPaginator", params).
					Return(nil, assert.AnError)
			},
			expected: [][]string{
				{"db1", "rds", "G-db1-abcdef", "(unknown)"},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactoryRDS := new(appmock.MockRDSFactory)
			mockDescribeDBClustersPaginator := new(appmock.MockDescribeDBClustersPaginator)
			mockDescribePendingMaintenanceActionsPaginator := new(appmock.MockDescribePendingMaintenanceActionsPaginator)
			mockFactoryCloudFormation := new(appmock.MockCloudFormationFactory)
			mockClientCloudFormation := new(appmock.MockCloudFormationClient)
			mockListStacksPaginator := new(appmock.MockListStacksPaginator)

			tc.mockDescribeDBClustersSetup(mockFactoryRDS, mockDescribeDBClustersPaginator)
			tc.mockDescribePendingMaintenanceActionsSetup(mockFactoryRDS, mockDescribePendingMaintenanceActionsPaginator)
			tc.mockListStacksSetup(mockFactoryCloudFormation, mockListStacksPaginator)
			tc.mockGetTemplateSetup(mockFactoryCloudFormation, mockClientCloudFormation)

			k := &ktnh{
				stackNamePrefix: tc.stackNamePrefix,
				rds:             apprds.NewRDS(mockFactoryRDS),
				cfn:             appcfn.NewCloudFormation(mockFactoryCloudFormation),
			}

			_, got, err := k.List()

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Output body does not match expected body")
			}

			mockFactoryRDS.AssertExpectations(t)
			mockDescribeDBClustersPaginator.AssertExpectations(t)
			mockDescribePendingMaintenanceActionsPaginator.AssertExpectations(t)
			mockFactoryCloudFormation.AssertExpectations(t)
			mockClientCloudFormation.AssertExpectations(t)
			mockListStacksPaginator.AssertExpectations(t)
		})
	}
}
