package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"strconv"
	"testing"
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

func TestHandlerFunction(t *testing.T) {
	t.Run("Testing the lambda function", func(t *testing.T) {
		_, err := handler(context.TODO(), InstanceSchedulingRequest{Action: "Test"})
		if err != nil {
			t.Fatal("Everything should be ok")
		}
	})
}
