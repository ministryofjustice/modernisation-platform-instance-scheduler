package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockLoadDefaultConfigWithError() (aws.Config, error) {
	return aws.Config{}, errors.New("Mock Error!")
}

func mockLoadDefaultConfig() (aws.Config, error) {
	return aws.Config{}, nil
}
func mockGetEnv(key string) string {
	return "skip-me,skip-me-too"
}

type MockISSMGetParameter struct {
	mock.Mock
}

func (m *MockISSMGetParameter) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return new(ssm.GetParameterOutput), nil
}

func mockCreateSSMClient(aws.Config) ISSMGetParameter {
	return new(MockISSMGetParameter)
}

func mockHandlerGetParameter(client ISSMGetParameter, parameterName string) string {
	return "test parameter"
}

func mockCreateSecretManagerClient(cfg aws.Config) ISecretManagerGetSecretValue {
	return nil
}

func mockGetSecret(client ISecretManagerGetSecretValue, secretId string) string {
	return `{
		"account_ids": {
			"test-account-development": "1",
			"test-account-preproduction": "2",
			"test-account-test": "3",
			"test-account-production": "4"
		}
	}`
}

type MockGetEc2ClientForMemberAccount struct {
	mock.Mock
	IEC2InstancesAPI
}

func mockGetEc2ClientForMemberAccount(cfg aws.Config, accountName string, accountId string) IEC2InstancesAPI {
	return new(MockGetEc2ClientForMemberAccount)
}

func mockGetEc2ClientForMemberAccountError(cfg aws.Config, accountName string, accountId string) IEC2InstancesAPI {
	return nil
}

type MockGetRDSClientForMemberAccount struct {
	mock.Mock
	IRDSInstancesAPI
}

func mockGetRdsClientForMemberAccount(cfg aws.Config, accountName string, accountId string) IRDSInstancesAPI {
	return new(MockGetRDSClientForMemberAccount)
}

func mockGetRdsClientForMemberAccountError(cfg aws.Config, accountName string, accountId string) IRDSInstancesAPI {
	return nil
}

func mockStopStartTestInstancesInMemberAccount(client IEC2InstancesAPI, action string) *InstanceCount {
	return &InstanceCount{
		actedUpon:         1,
		skipped:           1,
		skippedAutoScaled: 1,
	}
}

func mockStopStartTestRDSInstancesInMemberAccount(RDSClient IRDSInstancesAPI, action string) *RDSInstanceCount {
	return &RDSInstanceCount{
		RDSActedUpon: 1,
		RDSSkipped:   1,
	}
}

func TestHandlerUnit(t *testing.T) {
	t.Run("returns 400 error status and empty response when `InstanceSchedulingRequest.Action` is invalid", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{LoadDefaultConfig: mockLoadDefaultConfig}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "Invalid Action! 😱"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 400)
		assert.Equal(t, responseBody.Action, "Invalid Action! 😱")
		assert.Equal(t, responseBody.MemberAccountNames, []string{})
		assert.Equal(t, responseBody.NonMemberAccountNames, []string{})
		assert.Equal(t, responseBody.ActedUpon, 0)
		assert.Equal(t, responseBody.Skipped, 0)
		assert.Equal(t, responseBody.SkippedAutoScaled, 0)
		assert.Equal(t, responseBody.RDSActedUpon, 0)
		assert.Equal(t, responseBody.RDSSkipped, 0)
		assert.NotNil(t, err)
	})

	t.Run("returns 500 error status and empty response when default config cannot load", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{LoadDefaultConfig: mockLoadDefaultConfigWithError}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 500)
		assert.Equal(t, responseBody.Action, "test")
		assert.Equal(t, responseBody.MemberAccountNames, []string{})
		assert.Equal(t, responseBody.NonMemberAccountNames, []string{})
		assert.Equal(t, responseBody.ActedUpon, 0)
		assert.Equal(t, responseBody.Skipped, 0)
		assert.Equal(t, responseBody.SkippedAutoScaled, 0)
		assert.Equal(t, responseBody.RDSActedUpon, 0)
		assert.Equal(t, responseBody.RDSSkipped, 0)
		assert.NotNil(t, err)
	})

	t.Run("returns 200 status and empty response when no non-production accounts found", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:         mockLoadDefaultConfig,
			GetEnv:                    mockGetEnv,
			CreateSSMClient:           mockCreateSSMClient,
			GetParameter:              mockHandlerGetParameter,
			CreateSecretManagerClient: mockCreateSecretManagerClient,
			GetSecret: func(client ISecretManagerGetSecretValue, secretId string) string {
				return `{
					"account_ids": {}
				}`
			},
		}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 200)
		assert.Equal(t, responseBody.Action, "test")
		assert.Equal(t, responseBody.MemberAccountNames, []string{})
		assert.Equal(t, responseBody.NonMemberAccountNames, []string{})
		assert.Equal(t, responseBody.ActedUpon, 0)
		assert.Equal(t, responseBody.Skipped, 0)
		assert.Equal(t, responseBody.SkippedAutoScaled, 0)
		assert.Equal(t, responseBody.RDSActedUpon, 0)
		assert.Equal(t, responseBody.RDSSkipped, 0)
		assert.Nil(t, err)
	})

	t.Run("returns 200 status and empty response when all found accounts are skipped or production", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:         mockLoadDefaultConfig,
			GetEnv:                    mockGetEnv,
			CreateSSMClient:           mockCreateSSMClient,
			GetParameter:              mockHandlerGetParameter,
			CreateSecretManagerClient: mockCreateSecretManagerClient,
			GetSecret: func(client ISecretManagerGetSecretValue, secretId string) string {
				return "{\"account_ids\":{\"skip-me\":\"1\",\"skip-me-too\":\"2\",\"skip-me-production\":\"3\"}}"
			},
		}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 200)
		assert.Equal(t, responseBody.Action, "test")
		assert.Equal(t, responseBody.MemberAccountNames, []string{})
		assert.Equal(t, responseBody.NonMemberAccountNames, []string{})
		assert.Equal(t, responseBody.ActedUpon, 0)
		assert.Equal(t, responseBody.Skipped, 0)
		assert.Equal(t, responseBody.SkippedAutoScaled, 0)
		assert.Equal(t, responseBody.RDSActedUpon, 0)
		assert.Equal(t, responseBody.RDSSkipped, 0)
		assert.Nil(t, err)
	})

	t.Run("returns 200 status and returns full response and counts number of non-member accounts", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:            mockLoadDefaultConfig,
			GetEnv:                       mockGetEnv,
			CreateSSMClient:              mockCreateSSMClient,
			GetParameter:                 mockHandlerGetParameter,
			CreateSecretManagerClient:    mockCreateSecretManagerClient,
			GetSecret:                    mockGetSecret,
			GetEc2ClientForMemberAccount: mockGetEc2ClientForMemberAccountError,
			GetRDSClientForMemberAccount: mockGetRdsClientForMemberAccountError,
		}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 200)
		assert.Equal(t, responseBody.Action, "test")
		assert.ElementsMatch(t, responseBody.MemberAccountNames, []string{})
		assert.ElementsMatch(t, responseBody.NonMemberAccountNames, []string{"test-account-development", "test-account-preproduction", "test-account-test"})
		assert.Equal(t, responseBody.ActedUpon, 0)
		assert.Equal(t, responseBody.Skipped, 0)
		assert.Equal(t, responseBody.SkippedAutoScaled, 0)
		assert.Equal(t, responseBody.RDSActedUpon, 0)
		assert.Equal(t, responseBody.RDSSkipped, 0)
		assert.Nil(t, err)
	})

	t.Run("returns 200 status and returns full response and counts skipped instances for member accounts", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:                        mockLoadDefaultConfig,
			GetEnv:                                   mockGetEnv,
			CreateSSMClient:                          mockCreateSSMClient,
			GetParameter:                             mockHandlerGetParameter,
			CreateSecretManagerClient:                mockCreateSecretManagerClient,
			GetSecret:                                mockGetSecret,
			GetEc2ClientForMemberAccount:             mockGetEc2ClientForMemberAccount,
			GetRDSClientForMemberAccount:             mockGetRdsClientForMemberAccount,
			StopStartTestInstancesInMemberAccount:    mockStopStartTestInstancesInMemberAccount,
			StopStartTestRDSInstancesInMemberAccount: mockStopStartTestRDSInstancesInMemberAccount,
		}

		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})

		responseBody := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(response.Body), &responseBody)
		assert.Equal(t, response.StatusCode, 200)
		assert.Equal(t, responseBody.Action, "test")
		assert.ElementsMatch(t, responseBody.MemberAccountNames, []string{"test-account-development", "test-account-preproduction", "test-account-test"})
		assert.ElementsMatch(t, responseBody.NonMemberAccountNames, []string{})
		assert.Equal(t, responseBody.ActedUpon, 3)
		assert.Equal(t, responseBody.Skipped, 3)
		assert.Equal(t, responseBody.SkippedAutoScaled, 3)
		assert.Equal(t, responseBody.RDSActedUpon, 3)
		assert.Equal(t, responseBody.RDSSkipped, 3)
		assert.Nil(t, err)
	})
}
