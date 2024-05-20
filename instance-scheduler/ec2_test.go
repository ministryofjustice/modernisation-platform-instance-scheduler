package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

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

func TestStopStartTestInstancesInMemberAccount(t *testing.T) {
	tests := []struct {
		testTitle     string
		client        *mockIEC2InstancesAPI
		action        string
		expectedCount InstanceCount
	}{
		{
			testTitle: "testing Test action",
			client: &mockIEC2InstancesAPI{
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
			testTitle: "testing Stop action",
			client: &mockIEC2InstancesAPI{
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
			testTitle: "testing Start action",
			client: &mockIEC2InstancesAPI{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []ec2type.Reservation{
						{
							ReservationId: aws.String("r-0899f7abdd9be06d8"),
							Instances: []ec2type.Instance{
								// aws:autoscaling:groupName is set, therefore skip scheduling, skipped auto scaled: 1
								{
									InstanceId: aws.String("i-6567788001"),
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
									InstanceId: aws.String("i-6562788001"),
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
									InstanceId: aws.String("i-2162279010"),
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
								// instance-scheduling = skip-auto-stop, therefore skip auto stop, but not start, acted upon: 1
								{
									InstanceId: aws.String("i-1265579100"),
									Tags: []ec2type.Tag{
										{
											Key:   aws.String("instance-scheduling"),
											Value: aws.String("skip-auto-stop"),
										},
									},
								},
								// instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
								{
									InstanceId: aws.String("i-9262279010"),
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
			action:        "Start",
			expectedCount: InstanceCount{5, 2, 2},
		},
		{
			testTitle: "testing if action input is not case sensitive when passing start",
			client: &mockIEC2InstancesAPI{
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
			// aws:autoscaling:groupName tag is set, but action is an empty string, therefore InstanceCount: {0,0,0}
			testTitle: "testing empty action input",
			client: &mockIEC2InstancesAPI{
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
			client: &mockIEC2InstancesAPI{
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
