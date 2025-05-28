package cfn

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_CreateStack(t *testing.T) {
	testCases := []struct {
		name         string
		stackName    string
		templateBody string
		mockSetup    func(*appmock.MockCloudFormationClient)
		wantErr      bool
	}{
		{
			name:         "Success",
			stackName:    "success-stack",
			templateBody: "{a: 1}",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("success-stack"),
					TemplateBody: aws.String("{a: 1}"),
					Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			wantErr: false,
		},
		{
			name:         "API error",
			stackName:    "api-error-stack",
			templateBody: "[]",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.CreateStackInput{
					StackName:    aws.String("api-error-stack"),
					TemplateBody: aws.String("[]"),
					Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
				}

				result := &cloudformation.CreateStackOutput{}

				c.On("CreateStack", mock.Anything, params, mock.Anything).
					Return(result, fmt.Errorf("Error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(appmock.MockCloudFormationClient)

			tc.mockSetup(mockClient)

			mockFactory := new(appmock.MockCloudFormationFactory)

			mockFactory.On("GetClient").Return(mockClient)

			c := NewCloudFormation(mockFactory)

			err := c.CreateStack(tc.stackName, tc.templateBody)

			mockClient.AssertExpectations(t)
			mockFactory.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}
		})
	}
}

func Test_DeleteStack(t *testing.T) {
	testCases := []struct {
		name      string
		stackName string
		mockSetup func(*appmock.MockCloudFormationClient)
		wantErr   bool
	}{
		{
			name:      "Success",
			stackName: "success-stack",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.DeleteStackInput{
					StackName: aws.String("success-stack"),
				}

				result := &cloudformation.DeleteStackOutput{}

				c.On("DeleteStack", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			wantErr: false,
		},
		{
			name:      "API error",
			stackName: "api-error-stack",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.DeleteStackInput{
					StackName: aws.String("api-error-stack"),
				}

				result := &cloudformation.DeleteStackOutput{}

				c.On("DeleteStack", mock.Anything, params, mock.Anything).
					Return(result, fmt.Errorf("Error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(appmock.MockCloudFormationClient)

			tc.mockSetup(mockClient)

			mockFactory := new(appmock.MockCloudFormationFactory)

			mockFactory.On("GetClient").Return(mockClient)

			c := NewCloudFormation(mockFactory)

			err := c.DeleteStack(tc.stackName)

			mockClient.AssertExpectations(t)
			mockFactory.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}
		})
	}
}

func Test_ListStacks(t *testing.T) {
	testCases := []struct {
		name           string
		evaluator      stackEvaluator
		mockSetup      func(*appmock.MockCloudFormationFactory, *appmock.MockListStacksPaginator)
		expectedStacks []string
		wantErr        bool
	}{
		{
			name: "With evaluator",
			evaluator: func(stackName string) bool {
				return stackName == "stack1" || stackName == "stack3"
			},
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("stack1"),
						},
						{
							StackName: aws.String("stack2"),
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
			expectedStacks: []string{
				"stack1",
			},
			wantErr: false,
		},
		{
			name:      "Multiple pages",
			evaluator: nil,
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result1 := &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("stack1"),
						},
						{
							StackName: aws.String("stack2"),
						},
					},
				}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result1, nil).
					Once()

				p.On("HasMorePages").
					Return(true).
					Once()

				result2 := &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("stack3"),
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
			expectedStacks: []string{
				"stack1",
				"stack2",
				"stack3",
			},
			wantErr: false,
		},
		{
			name:      "No stacks found",
			evaluator: nil,
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
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
			expectedStacks: []string{},
			wantErr:        false,
		},
		{
			name: "Evaluator filters all stacks",
			evaluator: func(stackName string) bool {
				return false
			},
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("stack1"),
						},
						{
							StackName: aws.String("stack2"),
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
			expectedStacks: []string{},
			wantErr:        false,
		},
		{
			name:      "Paginator creation error",
			evaluator: nil,
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(nil, fmt.Errorf("Error"))
			},
			expectedStacks: nil,
			wantErr:        true,
		},
		{
			name:      "NextPage error",
			evaluator: nil,
			mockSetup: func(f *appmock.MockCloudFormationFactory, p *appmock.MockListStacksPaginator) {
				f.On("NewListStacksPaginator", mock.Anything).
					Return(p, nil)

				p.On("HasMorePages").
					Return(true).
					Once()

				result := &cloudformation.ListStacksOutput{}

				p.On("NextPage", mock.Anything, mock.Anything).
					Return(result, fmt.Errorf("Error")).
					Once()
			},
			expectedStacks: nil,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPaginator := new(appmock.MockListStacksPaginator)
			mockFactory := new(appmock.MockCloudFormationFactory)

			tc.mockSetup(mockFactory, mockPaginator)

			c := NewCloudFormation(mockFactory)

			got, err := c.ListStacks(tc.evaluator)

			mockPaginator.AssertExpectations(t)
			mockFactory.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.ElementsMatch(t, tc.expectedStacks, got)
			}
		})
	}
}
