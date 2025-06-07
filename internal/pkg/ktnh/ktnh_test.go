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

func Test_shortenIdentifier(t *testing.T) {
	testCases := []struct {
		name         string
		dbIdentifier string
		expected     string
	}{
		{
			name:         "Empty",
			dbIdentifier: "",
			expected:     "",
		},
		{
			name:         "Short",
			dbIdentifier: "12345",
			expected:     "12345",
		},
		{
			name:         "Same length as truncation length",
			dbIdentifier: "1234567890",
			expected:     "1234567890",
		},
		{
			name:         "Ending with number",
			dbIdentifier: "1234567890123",
			expected:     "1234567890",
		},
		{
			name:         "Ending with lowercase letter",
			dbIdentifier: "123456789abcd",
			expected:     "123456789a",
		},
		{
			name:         "Ending with uppercase letter",
			dbIdentifier: "123456789ABCD",
			expected:     "123456789A",
		},
		{
			name:         "Ending with non-alphanumeric",
			dbIdentifier: "123456789-abcd",
			expected:     "123456789-a",
		},
		{
			name:         "Ending with multiple non-alphanumeric",
			dbIdentifier: "123456789--abc",
			expected:     "123456789--",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shortenIdentifier(tc.dbIdentifier)

			assert.Equal(t, tc.expected, got, "Shortened identifier does not match expected value")
		})
	}
}

func Test_generateStackName(t *testing.T) {
	testCases := []struct {
		name            string
		stackNamePrefix string
		option          *stackNameOption
		expected        string
	}{
		{
			name:            "Identifier and qualifier",
			stackNamePrefix: "A",
			option: &stackNameOption{
				dbIdentifierShort: "full-1",
				qualifier:         "abcdef",
			},
			expected: "A-full-1-abcdef",
		},
		{
			name:            "Identifier",
			stackNamePrefix: "B",
			option: &stackNameOption{
				dbIdentifierShort: "partial-2",
			},
			expected: "B-partial-2-.+",
		},
		{
			name:            "Qualifier",
			stackNamePrefix: "C",
			option: &stackNameOption{
				qualifier: "ghijkl",
			},
			expected: "C-.+",
		},
		{
			name:            "Prefix only",
			stackNamePrefix: "D",
			option:          &stackNameOption{},
			expected:        "D-.+",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k := &ktnh{
				stackNamePrefix: tc.stackNamePrefix,
			}

			got := k.generateStackName(tc.option)

			assert.Equal(t, tc.expected, got, "Generated stack name does not match expected value")
		})
	}
}

func Test_findMatchingStack(t *testing.T) {
	testCases := []struct {
		name                     string
		dbIdentifier             string
		dbIdentifierShort        string
		stackNamePrefix          string
		mockDetermineDBTypeSetup func(*appmock.MockRDSFactory, *appmock.MockRDSClient)
		mockListStacksSetup      func(*appmock.MockCloudFormationFactory, *appmock.MockListStacksPaginator)
		mockGetTemplateSetup     func(*appmock.MockCloudFormationFactory, *appmock.MockCloudFormationClient)
		expectedStackName        string
		expectedFound            bool
		wantErr                  bool
	}{
		{
			name:              "Stack found",
			dbIdentifier:      "db-1-1234567890",
			dbIdentifierShort: "db-1-12345",
			stackNamePrefix:   "A",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-1-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("A-db-1-12345-abcdef"),
						},
						{
							StackName: aws.String("A-db-2-12345-ghijkl"),
						},
						{
							StackName: aws.String("A-db-1-12345-mnopqr"),
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
					StackName: aws.String("A-db-1-12345-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-1-1234567890'
    DBType: 'aurora'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("A-db-1-12345-mnopqr"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-1-1234567890'
    DBType: 'rds'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			expectedStackName: "A-db-1-12345-abcdef",
			expectedFound:     true,
			wantErr:           false,
		},
		{
			name:              "No stack found",
			dbIdentifier:      "db-2-1234567890",
			dbIdentifierShort: "db-2-12345",
			stackNamePrefix:   "B",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-2-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
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
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expectedStackName:    "",
			expectedFound:        false,
			wantErr:              false,
		},
		{
			name:              "Error during determining DB type",
			dbIdentifier:      "db-3-1234567890",
			dbIdentifierShort: "db-3-12345",
			stackNamePrefix:   "C",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-3-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, assert.AnError)
			},
			mockListStacksSetup:  func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expectedStackName:    "",
			expectedFound:        false,
			wantErr:              true,
		},
		{
			name:              "Invalid stack name pattern",
			dbIdentifier:      "db-4-1234567890",
			dbIdentifierShort: "db-4-12345",
			stackNamePrefix:   "[invalid regex",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-4-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup:  func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expectedStackName:    "",
			expectedFound:        false,
			wantErr:              true,
		},
		{
			name:              "Error during listing stacks",
			dbIdentifier:      "db-5-1234567890",
			dbIdentifierShort: "db-5-12345",
			stackNamePrefix:   "D",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-5-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(nil, assert.AnError)
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expectedStackName:    "",
			expectedFound:        false,
			wantErr:              true,
		},
		{
			name:              "Multiple stacks found",
			dbIdentifier:      "db-6-1234567890",
			dbIdentifierShort: "db-6-12345",
			stackNamePrefix:   "E",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-6-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("E-db-6-12345-abcdef"),
						},
						{
							StackName: aws.String("E-db-6-12345-ghijkl"),
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
					StackName: aws.String("E-db-6-12345-abcdef"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-6-1234567890'
    DBType: 'aurora'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("E-db-6-12345-ghijkl"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-6-1234567890'
    DBType: 'aurora'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			expectedStackName: "",
			expectedFound:     false,
			wantErr:           true,
		},
		{
			name:              "Error during retrieving template",
			dbIdentifier:      "db-7-1234567890",
			dbIdentifierShort: "db-7-12345",
			stackNamePrefix:   "F",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-7-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("F-db-7-12345-abcdef"),
						},
						{
							StackName: aws.String("F-db-7-12345-ghijkl"),
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
					StackName: aws.String("F-db-7-12345-abcdef"),
				}

				result1 := &cloudformation.GetTemplateOutput{}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, assert.AnError).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("F-db-7-12345-ghijkl"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-7-1234567890'
    DBType: 'aurora'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			expectedStackName: "F-db-7-12345-ghijkl",
			expectedFound:     true,
			wantErr:           false,
		},
		{
			name:              "Error during metadata verification",
			dbIdentifier:      "db-8-1234567890",
			dbIdentifierShort: "db-8-12345",
			stackNamePrefix:   "G",
			mockDetermineDBTypeSetup: func(f *appmock.MockRDSFactory, c *appmock.MockRDSClient) {
				f.On("GetClient").
					Return(c)

				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-8-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-mysql"),
						},
					},
				}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("G-db-8-12345-mnopqr"),
						},
						{
							StackName: aws.String("G-db-8-12345-stuvwx"),
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
					StackName: aws.String("G-db-8-12345-mnopqr"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-8-1234567890'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()

				params2 := &cloudformation.GetTemplateInput{
					StackName: aws.String("G-db-8-12345-stuvwx"),
				}

				templateBody2 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-8-1234567890'
    DBType: 'aurora'
`

				result2 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody2),
				}

				c.On("GetTemplate", mock.Anything, params2, mock.Anything).
					Return(result2, nil).
					Once()
			},
			expectedStackName: "G-db-8-12345-stuvwx",
			expectedFound:     true,
			wantErr:           false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactoryRDS := new(appmock.MockRDSFactory)
			mockClientRDS := new(appmock.MockRDSClient)
			mockFactoryCloudFormation := new(appmock.MockCloudFormationFactory)
			mockClientCloudFormation := new(appmock.MockCloudFormationClient)
			mockPaginator := new(appmock.MockListStacksPaginator)

			tc.mockDetermineDBTypeSetup(mockFactoryRDS, mockClientRDS)
			tc.mockListStacksSetup(mockFactoryCloudFormation, mockPaginator)
			tc.mockGetTemplateSetup(mockFactoryCloudFormation, mockClientCloudFormation)

			k := &ktnh{
				dbIdentifier:      tc.dbIdentifier,
				dbIdentifierShort: tc.dbIdentifierShort,
				stackNamePrefix:   tc.stackNamePrefix,
				rds:               apprds.NewRDS(mockFactoryRDS),
				cfn:               appcfn.NewCloudFormation(mockFactoryCloudFormation),
			}

			stackName, found, err := k.findMatchingStack()

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expectedFound, found, "Found flag does not match expected value")
				assert.Equal(t, tc.expectedStackName, stackName, "Stack name does not match expected value")
			}

			mockFactoryRDS.AssertExpectations(t)
			mockClientRDS.AssertExpectations(t)
			mockFactoryCloudFormation.AssertExpectations(t)
			mockClientCloudFormation.AssertExpectations(t)
			mockPaginator.AssertExpectations(t)
		})
	}
}
