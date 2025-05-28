package ktnh

import (
	"fmt"
	"testing"
	"time"

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

func Test_Freeze(t *testing.T) {
	testCases := []struct {
		name                     string
		dbIdentifier             string
		dbIdentifierShort        string
		stackNamePrefix          string
		qualifier                string
		templateBody             string
		timeout                  time.Duration
		mockDetermineDBTypeSetup func(*appmock.MockRDSClient)
		mockListStacksSetup      func(*appmock.MockCloudFormationFactory, *appmock.MockListStacksPaginator)
		mockGetTemplateSetup     func(*appmock.MockCloudFormationFactory, *appmock.MockCloudFormationClient)
		mockCreateStackSetup     func(*appmock.MockCloudFormationClient)
		mockWaitSetup            func(*appmock.MockCloudFormationFactory, *appmock.MockStackCreateCompleteWaiter)
		wantErr                  bool
	}{
		{
			name:              "With wait",
			dbIdentifier:      "db-1-1234567890",
			dbIdentifierShort: "db-1-12345",
			stackNamePrefix:   "A",
			qualifier:         "abcdef",
			templateBody:      "{a: 1}",
			timeout:           time.Minute * 5,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
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
					StackName: aws.String("A-db-1-12345-mnopqr"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-1-1234567890'
    DBType: 'rds'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()
			},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("A-db-1-12345-abcdef"),
					TemplateBody: aws.String("{a: 1}"),
					Capabilities: []cfntypes.Capability{cfntypes.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockWaitSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {
				f.On("NewStackCreateCompleteWaiter").
					Return(w, nil)

				params := &cloudformation.DescribeStacksInput{
					StackName: aws.String("A-db-1-12345-abcdef"),
				}

				w.On("Wait", mock.Anything, params, time.Minute*5, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:              "Without wait",
			dbIdentifier:      "db-2-1234567890",
			dbIdentifierShort: "db-2-12345",
			stackNamePrefix:   "B",
			qualifier:         "ghijkl",
			templateBody:      "{b: 2}",
			timeout:           0,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
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
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)
			},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("B-db-2-12345-ghijkl"),
					TemplateBody: aws.String("{b: 2}"),
					Capabilities: []cfntypes.Capability{cfntypes.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockWaitSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {},
			wantErr:       false,
		},
		{
			name:              "Error during finding stack",
			dbIdentifier:      "db-3-1234567890",
			dbIdentifierShort: "db-3-12345",
			stackNamePrefix:   "C",
			qualifier:         "mnopqr",
			templateBody:      "{c: 3}",
			timeout:           time.Minute * 5,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-3-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{}

				c.On("DescribeDBClusters", mock.Anything, params, mock.Anything).
					Return(result, fmt.Errorf("Error"))
			},
			mockListStacksSetup:  func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {},
			mockWaitSetup:        func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {},
			wantErr:              true,
		},
		{
			name:              "Stack exists",
			dbIdentifier:      "db-4-1234567890",
			dbIdentifierShort: "db-4-12345",
			stackNamePrefix:   "D",
			qualifier:         "stuvwx",
			templateBody:      "{d: 4}",
			timeout:           time.Minute * 5,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
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
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []cfntypes.StackSummary{
						{
							StackName: aws.String("D-db-4-12345-stuvwx"),
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
					StackName: aws.String("D-db-4-12345-stuvwx"),
				}

				templateBody1 := `
Metadata:
  KTNH:
    Generator: 'koreru-toki-no-hiho'
    Version: '1'
    DBIdentifier: 'db-4-1234567890'
    DBType: 'aurora'
`

				result1 := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody1),
				}

				c.On("GetTemplate", mock.Anything, params1, mock.Anything).
					Return(result1, nil).
					Once()
			},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {},
			mockWaitSetup:        func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {},
			wantErr:              true,
		},
		{
			name:              "Error during stack creation",
			dbIdentifier:      "db-5-1234567890",
			dbIdentifierShort: "db-5-12345",
			stackNamePrefix:   "E",
			qualifier:         "zyxwvu",
			templateBody:      "{e: 5}",
			timeout:           time.Minute * 5,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
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
					Return(p, nil)

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)
			},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("E-db-5-12345-zyxwvu"),
					TemplateBody: aws.String("{e: 5}"),
					Capabilities: []cfntypes.Capability{cfntypes.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, fmt.Errorf("Error"))
			},
			mockWaitSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {},
			wantErr:       true,
		},
		{
			name:              "Error during waiter",
			dbIdentifier:      "db-6-1234567890",
			dbIdentifierShort: "db-6-12345",
			stackNamePrefix:   "F",
			qualifier:         "tsrqpo",
			templateBody:      "{f: 6}",
			timeout:           time.Minute * 5,
			mockDetermineDBTypeSetup: func(c *appmock.MockRDSClient) {
				params := &rds.DescribeDBClustersInput{
					DBClusterIdentifier: aws.String("db-6-1234567890"),
				}

				result := &rds.DescribeDBClustersOutput{
					DBClusters: []rdstypes.DBCluster{
						{
							Engine: aws.String("aurora-postgresql"),
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
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {
				f.On("GetClient").
					Return(c)
			},
			mockCreateStackSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("F-db-6-12345-tsrqpo"),
					TemplateBody: aws.String("{f: 6}"),
					Capabilities: []cfntypes.Capability{cfntypes.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			mockWaitSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {
				f.On("NewStackCreateCompleteWaiter").
					Return(nil, fmt.Errorf("Error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClientRDS := new(appmock.MockRDSClient)

			tc.mockDetermineDBTypeSetup(mockClientRDS)

			mockFactoryRDS := new(appmock.MockRDSFactory)

			mockFactoryRDS.On("GetClient").Return(mockClientRDS)

			mockFactoryCloudFormation := new(appmock.MockCloudFormationFactory)

			mockPaginator := new(appmock.MockListStacksPaginator)

			tc.mockListStacksSetup(mockFactoryCloudFormation, mockPaginator)

			mockClientCloudFormation := new(appmock.MockCloudFormationClient)

			tc.mockGetTemplateSetup(mockFactoryCloudFormation, mockClientCloudFormation)

			tc.mockCreateStackSetup(mockClientCloudFormation)

			mockWaiter := new(appmock.MockStackCreateCompleteWaiter)

			tc.mockWaitSetup(mockFactoryCloudFormation, mockWaiter)

			k := &ktnh{
				dbIdentifier:      tc.dbIdentifier,
				dbIdentifierShort: tc.dbIdentifierShort,
				stackNamePrefix:   tc.stackNamePrefix,
				rds:               apprds.NewRDS(mockFactoryRDS),
				cfn:               appcfn.NewCloudFormation(mockFactoryCloudFormation),
			}

			err := k.Freeze(tc.templateBody, tc.qualifier, tc.timeout)

			mockClientRDS.AssertExpectations(t)
			mockFactoryRDS.AssertExpectations(t)

			mockClientCloudFormation.AssertExpectations(t)
			mockPaginator.AssertExpectations(t)
			mockWaiter.AssertExpectations(t)
			mockFactoryCloudFormation.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}
		})
	}
}
