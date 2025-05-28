package cfn

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_GetKTNHMetadata(t *testing.T) {
	testCases := []struct {
		name      string
		stackName string
		mockSetup func(*appmock.MockCloudFormationClient)
		expected  *ktnhMetadata
		wantErr   bool
	}{
		{
			name:      "Valid",
			stackName: "valid-stack",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.GetTemplateInput{
					StackName: aws.String("valid-stack"),
				}

				templateBody := `
Metadata:
  KTNH:
    Generator: 'test-generator'
    Version: '10'
    DBIdentifier: 'test-db'
    DBType: 'aurora'
`

				result := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String(templateBody),
				}

				c.On("GetTemplate", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			expected: &ktnhMetadata{
				Generator:    "test-generator",
				Version:      "10",
				DBIdentifier: "test-db",
				DBType:       "aurora",
			},
			wantErr: false,
		},
		{
			name:      "API error",
			stackName: "api-error-stack",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.GetTemplateInput{
					StackName: aws.String("api-error-stack"),
				}

				result := &cloudformation.GetTemplateOutput{}

				c.On("GetTemplate", mock.Anything, params, mock.Anything).
					Return(result, fmt.Errorf("Error"))
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:      "Invalid template",
			stackName: "invalid-template-stack",
			mockSetup: func(c *appmock.MockCloudFormationClient) {
				params := &cloudformation.GetTemplateInput{
					StackName: aws.String("invalid-template-stack"),
				}

				result := &cloudformation.GetTemplateOutput{
					TemplateBody: aws.String("Invalid: {[}"),
				}

				c.On("GetTemplate", mock.Anything, params, mock.Anything).
					Return(result, nil)
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(appmock.MockCloudFormationClient)

			tc.mockSetup(mockClient)

			mockFactory := new(appmock.MockCloudFormationFactory)

			mockFactory.On("GetClient").Return(mockClient)

			c := NewCloudFormation(mockFactory)

			got, err := c.GetKTNHMetadata(tc.stackName)

			mockClient.AssertExpectations(t)
			mockFactory.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Metadata does not match expected value")
			}
		})
	}
}

func Test_VerifyMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		metadata *ktnhMetadata
		option   *MetadataVerifyOption
		expected bool
		wantErr  bool
	}{
		{
			name: "Valid metadata",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				Version:      "1",
				DBIdentifier: "db-1",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-1",
				DBType:       "aurora",
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "No options",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				Version:      "1",
				DBIdentifier: "db-2",
				DBType:       "aurora",
			},
			option:   &MetadataVerifyOption{},
			expected: true,
			wantErr:  false,
		},
		{
			name: "Empty Generator field",
			metadata: &ktnhMetadata{
				Version:      "1",
				DBIdentifier: "db-3",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-3",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "Empty Version field",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				DBIdentifier: "db-4",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-4",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "Empty DBIdentifier field",
			metadata: &ktnhMetadata{
				Generator: "koreru-toki-no-hiho",
				Version:   "1",
				DBType:    "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-5",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "Empty DBType field",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				Version:      "1",
				DBIdentifier: "db-6",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-6",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "Generator mismatch",
			metadata: &ktnhMetadata{
				Generator:    "another-generator",
				Version:      "1",
				DBIdentifier: "db-7",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-7",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "DBIdentifier mismatch",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				Version:      "1",
				DBIdentifier: "db-8",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "another-db-identifier",
				DBType:       "aurora",
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "DBType mismatch",
			metadata: &ktnhMetadata{
				Generator:    "koreru-toki-no-hiho",
				Version:      "1",
				DBIdentifier: "db-9",
				DBType:       "aurora",
			},
			option: &MetadataVerifyOption{
				DBIdentifier: "db-9",
				DBType:       "rds",
			},
			expected: false,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := VerifyMetadata(tc.metadata, tc.option)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Verification result does not match expected value")
			}
		})
	}
}
