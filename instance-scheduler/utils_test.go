package main

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
)

type mockGetParameter func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)

func (m mockGetParameter) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m(ctx, params, optFns...)
}

func TestGetParameter(t *testing.T) {
	tests := []struct {
		client func(t *testing.T) ISSMGetParameter
		name   string
		want   string
	}{
		{
			client: func(t *testing.T) ISSMGetParameter {
				return mockGetParameter(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					t.Helper()
					if params.Name == nil {
						t.Fatal("want Name to not be nil")
					}
					if want, got := "test-parameter-name", *params.Name; want != got {
						t.Errorf("want %v, got %v", want, got)
					}

					return &ssm.GetParameterOutput{
						Parameter: &types.Parameter{
							Value: aws.String("test-parameter-value"),
						},
					}, nil
				})
			},
			name: "test-parameter-name",
			want: "test-parameter-value",
		},
	}

	for i, subtest := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			parameter := getParameter(subtest.client(t), subtest.name)
			if want, got := subtest.want, parameter; want != got {
				t.Errorf("want %v, got %v", subtest.want, got)
			}
		})
	}
}

type mockGetSecretValue func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

func (m mockGetSecretValue) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

func TestGetSecret(t *testing.T) {
	cases := []struct {
		client   func(t *testing.T) ISecretManagerGetSecretValue
		secretId string
		want     string
	}{
		{
			client: func(t *testing.T) ISecretManagerGetSecretValue {
				return mockGetSecretValue(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					t.Helper()
					if params.SecretId == nil {
						t.Fatal("want SecretId to not be nil")
					}
					if want, got := "test-mod-platform-account-development", *params.SecretId; want != got {
						t.Errorf("want %v, got %v", want, got)
					}

					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String("123456789012"),
					}, nil
				})
			},
			secretId: "test-mod-platform-account-development",
			want:     "123456789012",
		},
	}

	for i, subtest := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			secret := getSecret(subtest.client(t), subtest.secretId)
			if want, got := subtest.want, secret; want != got {
				t.Errorf("want %v, got %v", subtest.want, got)
			}
		})
	}
}

// func TestGetNonProductionAccounts(t *testing.T) {
// 	tests := []struct {
// 		testTitle    string
// 		environments string
// 		skipAccounts string
// 		want         map[string]string
// 	}{
// 		{
// 			testTitle:    "testing removal of production accounts and accounts to be skipped",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "test-account-test",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
// 		},
// 		{
// 			testTitle:    "testing duplicate accounts in skipAccounts",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "test-account-test, test-account-test",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
// 		},
// 		{
// 			testTitle:    "testing empty skipAccounts",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012", "test-account-test": "883115264813"},
// 		},
// 		{
// 			testTitle:    "testing empty environments and empty skipAccounts",
// 			environments: "{\"account_ids\":{}}",
// 			skipAccounts: "",
// 			want:         map[string]string{},
// 		},
// 		{
// 			testTitle:    "testing empty environments and non empty skipAccounts",
// 			environments: "{\"account_ids\":{}}",
// 			skipAccounts: "test-account-preproduction",
// 			want:         map[string]string{},
// 		},
// 		{
// 			testTitle:    "testing skipAccounts being of length 1",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "t",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
// 		},
// 		{
// 			testTitle:    "testing invalid skipAccounts",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "tdsh&*tew-ljijle32^@srw",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
// 		},
// 		{
// 			testTitle:    "testing account number in skipAccounts",
// 			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
// 			skipAccounts: "123456789098",
// 			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
// 		},
// 	}

// 	for _, subtest := range tests {
// 		t.Run(subtest.testTitle, func(t *testing.T) {
// 			got := getNonProductionAccounts(subtest.environments, subtest.skipAccounts)
// 			if !reflect.DeepEqual(subtest.want, got) {
// 				t.Errorf("want %v, got %v", subtest.want, got)
// 			}
// 		})
// 	}
// }

func TestParseAction(t *testing.T) {
	tests := []struct {
		title       string
		action      string
		want        string
		expectError bool
	}{
		{
			title:       "returns 'test' for `TEST`",
			action:      "TEST",
			want:        "test",
			expectError: false,
		},
		{
			title:       "returns 'start' for `START`",
			action:      "START",
			want:        "start",
			expectError: false,
		},
		{
			title:       "returns 'stop' for `STOP`",
			action:      "STOP",
			want:        "stop",
			expectError: false,
		},
		{
			title:       "returns empty string and error for invalid action`",
			action:      "Invalid action name! 😱",
			want:        "",
			expectError: true,
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.title, func(t *testing.T) {
			got, err := parseAction(subtest.action)
			assert.Equal(t, subtest.want, got)
			if subtest.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
