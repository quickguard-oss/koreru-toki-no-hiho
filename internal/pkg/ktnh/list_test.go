package ktnh

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appcfn "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/cfn"
	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_List(t *testing.T) {
	testCases := []struct {
		name                 string
		stackNamePrefix      string
		mockListStacksSetup  func(*appmock.MockCloudFormationFactory, *appmock.MockListStacksPaginator)
		mockGetTemplateSetup func(*appmock.MockCloudFormationFactory, *appmock.MockCloudFormationClient)
		expected             []string
		wantErr              bool
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
					StackSummaries: []types.StackSummary{
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
			expected: []string{
				"ID    TYPE     STACK",
				"db1   aurora   A-db1-abcdef",
				"db4   rds      A-db4-stuvwx",
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
					StackSummaries: []types.StackSummary{},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, nil).
					Once()

				p.On("HasMorePages").
					Return(false).
					Once()
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expected:             []string{},
			wantErr:              false,
		},
		{
			name:                 "Invalid stack name",
			stackNamePrefix:      "[invalid-regex",
			mockListStacksSetup:  func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expected:             nil,
			wantErr:              true,
		},
		{
			name:            "Error during listing stacks",
			stackNamePrefix: "C",
			mockListStacksSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(nil, fmt.Errorf("Error"))
			},
			mockGetTemplateSetup: func(f *appmock.MockCloudFormationFactory, c *appmock.MockCloudFormationClient) {},
			expected:             nil,
			wantErr:              true,
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
					StackSummaries: []types.StackSummary{
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
					Return(result1, fmt.Errorf("Error")).
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
			expected: []string{
				"ID    TYPE     STACK",
				"db2   aurora   D-db2-ghijkl",
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
					StackSummaries: []types.StackSummary{
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
			expected: []string{
				"ID    TYPE     STACK",
				"db2   aurora   E-db2-ghijkl",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockCloudFormationFactory)

			mockPaginator := new(appmock.MockListStacksPaginator)

			tc.mockListStacksSetup(mockFactory, mockPaginator)

			mockClient := new(appmock.MockCloudFormationClient)

			tc.mockGetTemplateSetup(mockFactory, mockClient)

			k := &ktnh{
				stackNamePrefix: tc.stackNamePrefix,
				cfn:             appcfn.NewCloudFormation(mockFactory),
			}

			got, err := k.List()

			mockClient.AssertExpectations(t)
			mockPaginator.AssertExpectations(t)
			mockFactory.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Output lines do not match expected lines")
			}
		})
	}
}
