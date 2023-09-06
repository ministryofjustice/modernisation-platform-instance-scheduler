package main

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstype "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

func TestGetNonProductionAccounts(t *testing.T) {
	tests := []struct {
		testTitle    string
		environments string
		skipAccounts string
		want         map[string]string
	}{
		{
			testTitle:    "testing removal of production accounts and accounts to be skipped",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "test-account-test",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
		},
		{
			testTitle:    "testing duplicate accounts in skipAccounts",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "test-account-test, test-account-test",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
		},
		{
			testTitle:    "testing empty skipAccounts",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-test\":\"883115264813\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012", "test-account-test": "883115264813"},
		},
		{
			testTitle:    "testing empty environments and empty skipAccounts",
			environments: "{\"account_ids\":{}}",
			skipAccounts: "",
			want:         map[string]string{},
		},
		{
			testTitle:    "testing empty environments and non empty skipAccounts",
			environments: "{\"account_ids\":{}}",
			skipAccounts: "test-account-preproduction",
			want:         map[string]string{},
		},
		{
			testTitle:    "testing skipAccounts being of length 1",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "t",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
		},
		{
			testTitle:    "testing invalid skipAccounts",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "tdsh&*tew-ljijle32^@srw",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
		},
		{
			testTitle:    "testing account number in skipAccounts",
			environments: "{\"account_ids\":{\"test-account-development\":\"123456789098\",\"test-account-preproduction\":\"983456789012\",\"test-account-production\":\"45400687236096\"}}",
			skipAccounts: "123456789098",
			want:         map[string]string{"test-account-development": "123456789098", "test-account-preproduction": "983456789012"},
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.testTitle, func(t *testing.T) {
			got := getNonProductionAccounts(subtest.environments, subtest.skipAccounts)
			if !reflect.DeepEqual(subtest.want, got) {
				t.Errorf("want %v, got %v", subtest.want, got)
			}
		})
	}
}

type mockIEC2InstancesAPI struct {
	DescribeInstancesOutput *ec2.DescribeInstancesOutput
	StartInstancesOutput    *ec2.StartInstancesOutput
	StopInstancesOutput     *ec2.StopInstancesOutput
}

func (m *mockIEC2InstancesAPI) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesOutput, nil
}

func (m *mockIEC2InstancesAPI) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	return m.StopInstancesOutput, nil
}

func (m *mockIEC2InstancesAPI) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	return m.StartInstancesOutput, nil
}

type mockIRDSInstancesAPI struct {
	DescribeDBInstancesOutput *rds.DescribeDBInstancesOutput
	StartDBInstancesOutput    *rds.StartDBInstancesOutput
	StopDBInstancesOutput     *rds.StopDBInstancesOutput
}

func (m *mockIRDSInstancesAPI) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.DescribeDBInstancesOutput, nil
}

func (m *mockIRDSInstancesAPI) StopDBInstances(ctx context.Context, params *rds.StopDBInstancesInput, optFns ...func(*rds.Options)) (*rds.StopDBInstancesOutput, error) {
	return m.StopDBInstancesOutput, nil
}

func (m *mockIRDSInstancesAPI) StartDBInstances(ctx context.Context, params *rds.StartDBInstancesInput, optFns ...func(*rds.Options)) (*rds.StartDBInstancesOutput, error) {
	return m.StartDBInstancesOutput, nil
}

func TestStopStartTestInstancesInMemberAccount(t *testing.T) {
	tests := []struct {
		testTitle     string
		ec2client     *mockIEC2InstancesAPI
		rdsclient     *mockIRDSInstancesAPI
		action        string
		expectedCount InstanceCount
	}{
		{
			testTitle: "testing Test action",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6567788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("bastion_linux_daily"),
										},
									},
								},
								// instance-scheduling = default, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562278100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
						{
							ReservationId: aws.String("r-0c318eab370f3d57a"),
							Instances: []ec2type.Instance{
								// both instance-scheduling and aws:autoscaling:groupName tags are set, skip scheduling due to autoscaling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6562788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("weblogic-CNOMT1"),
										},
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// no instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562279100"),
								},
								// instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
								{
									InstanceId: aws.String("i-2162279001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-scheduling"),
										},
									},
								},
								// instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7862279100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String(""),
										},
									},
								},
								// instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7863371100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("invalid-value"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not test, skipped: 1
								{
									InstanceId: aws.String("i-1265579001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, but not test, skipped: 1
								{
									InstanceId: aws.String("i-9262279001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "Test",
			expectedCount: InstanceCount{4, 3, 2},
		},
		{
			testTitle: "rdstesting Test action",
			rdsclient: &mockIRDSInstancesAPI{
				DescribeDBInstancesOutput: &rds.DescribeInstancesOutput{
					Reservations: []rdstype.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []rdstype.dbInstance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									dbInstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("rds-scalling"),
										},
									},
								},
								// instance-scheduling = default, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
						{
							ReservationId: aws.String("r-0c318eab370f3d57a"),
							Instances: []rdstype.dbInstance{
								// both instance-scheduling and aws:autoscaling:groupName tags are set, skip scheduling due to autoscaling, skipped auto scaled: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("rds-scalling"),
										},
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// no instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
								},
								// instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-scheduling"),
										},
									},
								},
								// instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String(""),
										},
									},
								},
								// instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("invalid-value"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not test, skipped: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, but not test, skipped: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "Test",
			expectedCount: InstanceCount{4, 3, 2},
		},
		{
			testTitle: "testing Stop action",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeDBInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6567788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("bastion_linux_daily"),
										},
									},
								},
								// instance-scheduling = default, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562278100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
						{
							ReservationId: aws.String("r-0c318eab370f3d57a"),
							Instances: []ec2type.Instance{
								// both instance-scheduling and aws:autoscaling:groupName tags are set, skip scheduling due to autoscaling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6562788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("weblogic-CNOMT1"),
										},
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// no instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562279100"),
								},
								// instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
								{
									InstanceId: aws.String("i-2162279001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-scheduling"),
										},
									},
								},
								// instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7862279100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String(""),
										},
									},
								},
								// instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7863371100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("invalid-value"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, skipped: 1
								{
									InstanceId: aws.String("i-1265579001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, but not stop, acted upon: 1
								{
									InstanceId: aws.String("i-9262279100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "Stop",
			expectedCount: InstanceCount{5, 2, 2},
		},
		{
			testTitle: "testing Stop action",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6567788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("bastion_linux_daily"),
										},
									},
								},
								// instance-scheduling = default, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562278100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
						{
							ReservationId: aws.String("r-0c318eab370f3d57a"),
							Instances: []ec2type.Instance{
								// both instance-scheduling and aws:autoscaling:groupName tags are set, skip scheduling due to autoscaling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6562788010"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("weblogic-CNOMT1"),
										},
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// no instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("i-6562279100"),
								},
								// instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
								{
									InstanceId: aws.String("i-2162279001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-scheduling"),
										},
									},
								},
								// instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7862279100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String(""),
										},
									},
								},
								// instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("i-7863371100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("invalid-value"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, skipped: 1
								{
									InstanceId: aws.String("i-1265579001"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, but not stop, acted upon: 1
								{
									InstanceId: aws.String("i-9262279100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "Stop",
			expectedCount: InstanceCount{5, 2, 2},
		},
		{
			testTitle: "RDS testing Start action",
			rdsclient: &mockIRDSInstancesAPI{
				DescribeDBInstancesOutput: &rds.DescribeInstancesOutput{
					Reservations: []rdstype.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []rdstype.dbInstance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("bastion_linux_daily"),
										},
									},
								},
								// instance-scheduling = default, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
						{
							ReservationId: aws.String("r-0c318eab370f3d57a"),
							Instances: []rdstype.dbInstance{
								// both instance-scheduling and aws:autoscaling:groupName tags are set, skip scheduling due to autoscaling, skipped auto scaled: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("weblogic-CNOMT1"),
										},
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// no instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
								{
									InstanceId: aws.String("test-database-no"),
								},
								// instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
								{
									InstanceId: aws.String("test-database-skip"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-scheduling"),
										},
									},
								},
								// instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String(""),
										},
									},
								},
								// instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("invalid-value"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not start, acted upon: 1
								{
									InstanceId: aws.String("test-database"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
								{
									InstanceId: aws.String("test-database-skip"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "Start",
			expectedCount: InstanceCount{5, 2, 2},
		},
		{
			testTitle: "testing if action input is not case sensitive when passing start",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								// instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
								{
									InstanceId: aws.String("i-9262279981"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not start, acted upon: 1
								{
									InstanceId: aws.String("i-1265579989"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "start",
			expectedCount: InstanceCount{1, 1, 0},
		},
		{
			testTitle: "RDS testing if action input is not case sensitive when passing start",
			rdsclient: &mockIRDSInstancesAPI{
				DescribeInstancesOutput: &rds.DescribeDBInstancesOutput{
					Reservations: []rdstype.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []rdstype.Instance{
								// instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
								{
									InstanceId: aws.String("test-database-skip"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-start"),
										},
									},
								},
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not start, acted upon: 1
								{
									InstanceId: aws.String("test-database-skip2"),
									Tags: []rdstype.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "start",
			expectedCount: InstanceCount{1, 1, 0},
		},
		{
			// aws:autoscaling:groupName tag is set, but action is an empty string, therefore InstanceCount: {0,0,0}
			testTitle: "testing empty action input",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								{
									InstanceId: aws.String("i-6567788909"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("aws:autoscaling:groupName"),
											Value: aws.String("bastion-linux"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "",
			expectedCount: InstanceCount{0, 0, 0},
		},
		{
			// instance-scheduling = default, but action value is invalid, therefore InstanceCount: {0,0,0}
			testTitle: "testing invalid action input",
			ec2client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								{
									InstanceId: aws.String("i-1265579989"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("default"),
										},
									},
								},
							},
						},
					},
				},
			},
			action:        "invalid",
			expectedCount: InstanceCount{0, 0, 0},
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.testTitle, func(t *testing.T) {
			actualInstanceCount := stopStartTestInstancesInMemberAccount(subtest.client, subtest.action)
			if want, got := subtest.expectedCount, actualInstanceCount; want != *got {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}
