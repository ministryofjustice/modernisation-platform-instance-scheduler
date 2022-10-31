package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"strconv"
	"testing"
)

//func TestGetNonProductionAccounts(t *testing.T) {
//	// Load the Shared AWS Configuration (~/.aws/config)
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	got := getNonProductionAccounts(cfg, "oasys-development,apex-development,data-and-insights-wepi-development,xhibit-portal-development,nomis-preproduction,testing-test,oasys-preproduction,tariff-development,mlra-development,nomis-development,example-development,performance-hub-preproduction,xhibit-portal-preproduction,delius-iaps-development,performance-hub-development,ppud-development,refer-monitor-development,digital-prison-reporting-development,oasys-test,equip-development,threat-and-vulnerability-mgmt-development,nomis-test,ccms-ebs-development,maatdb-development")
//	want := map[string]string{
//		"something": "something",
//	}
//	fmt.Println(got)
//	fmt.Println("")
//	fmt.Println(want)
//
//}

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
						Parameter: aws.String("test-parameter-arn"),
					}.Value, nil
				})
			},
			name: "test-parameter-name",
			want: "test-parameter-arn",
		},
	}

	for i, subtest := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			secret := getParameter(subtest.client(t), subtest.name)
			if want, got := subtest.want, secret; want != got {
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
