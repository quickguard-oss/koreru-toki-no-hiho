package cfn

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appmock "github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/mock"
)

func Test_WaitForStackCreation(t *testing.T) {
	testCases := []struct {
		name      string
		stackName string
		timeout   time.Duration
		mockSetup func(*appmock.MockCloudFormationFactory, *appmock.MockStackCreateCompleteWaiter)
		wantErr   bool
	}{
		{
			name:      "Success",
			stackName: "success-stack",
			timeout:   time.Minute * 5,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {
				f.On("NewStackCreateCompleteWaiter").
					Return(w, nil)

				params := &cloudformation.DescribeStacksInput{
					StackName: aws.String("success-stack"),
				}

				w.On("Wait", mock.Anything, params, time.Minute*5, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Timeout",
			stackName: "timeout-stack",
			timeout:   time.Second * 30,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {
				f.On("NewStackCreateCompleteWaiter").
					Return(w, nil)

				params := &cloudformation.DescribeStacksInput{
					StackName: aws.String("timeout-stack"),
				}

				w.On("Wait", mock.Anything, params, time.Second*30, mock.Anything).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:      "Factory error",
			stackName: "factory-error-stack",
			timeout:   time.Minute * 5,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackCreateCompleteWaiter) {
				f.On("NewStackCreateCompleteWaiter").
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockCloudFormationFactory)
			mockWaiter := new(appmock.MockStackCreateCompleteWaiter)

			tc.mockSetup(mockFactory, mockWaiter)

			c := NewCloudFormation(mockFactory)

			err := c.WaitForStackCreation(tc.stackName, tc.timeout)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}

			mockFactory.AssertExpectations(t)
			mockWaiter.AssertExpectations(t)
		})
	}
}

func Test_WaitForStackDeletion(t *testing.T) {
	testCases := []struct {
		name      string
		stackName string
		timeout   time.Duration
		mockSetup func(*appmock.MockCloudFormationFactory, *appmock.MockStackDeleteCompleteWaiter)
		wantErr   bool
	}{
		{
			name:      "success",
			stackName: "success-stack",
			timeout:   time.Minute * 5,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackDeleteCompleteWaiter) {
				f.On("NewStackDeleteCompleteWaiter").
					Return(w, nil)

				params := &cloudformation.DescribeStacksInput{
					StackName: aws.String("success-stack"),
				}

				w.On("Wait", mock.Anything, params, time.Minute*5, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "Timeout",
			stackName: "timeout-stack",
			timeout:   time.Second * 30,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackDeleteCompleteWaiter) {
				f.On("NewStackDeleteCompleteWaiter").
					Return(w, nil)

				params := &cloudformation.DescribeStacksInput{
					StackName: aws.String("timeout-stack"),
				}

				w.On("Wait", mock.Anything, params, time.Second*30, mock.Anything).
					Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:      "Factory error",
			stackName: "factory-error-stack",
			timeout:   time.Minute * 5,
			mockSetup: func(f *appmock.MockCloudFormationFactory, w *appmock.MockStackDeleteCompleteWaiter) {
				f.On("NewStackDeleteCompleteWaiter").
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFactory := new(appmock.MockCloudFormationFactory)
			mockWaiter := new(appmock.MockStackDeleteCompleteWaiter)

			tc.mockSetup(mockFactory, mockWaiter)

			c := NewCloudFormation(mockFactory)

			err := c.WaitForStackDeletion(tc.stackName, tc.timeout)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
			}

			mockFactory.AssertExpectations(t)
			mockWaiter.AssertExpectations(t)
		})
	}
}
