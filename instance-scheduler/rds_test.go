package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstype "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

type mockIRDSInstancesAPI struct {
	DescribeDBInstancesOutput *rds.DescribeDBInstancesOutput
	StartDBInstanceOutput     *rds.StartDBInstanceOutput
	StopDBInstanceOutput      *rds.StopDBInstanceOutput
}

func (m *mockIRDSInstancesAPI) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.DescribeDBInstancesOutput, nil
}

func (m *mockIRDSInstancesAPI) StopDBInstance(ctx context.Context, params *rds.StopDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StopDBInstanceOutput, error) {
	return m.StopDBInstanceOutput, nil
}

func (m *mockIRDSInstancesAPI) StartDBInstance(ctx context.Context, params *rds.StartDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StartDBInstanceOutput, error) {
	return m.StartDBInstanceOutput, nil
}
func TestStopStartTestRDSInstancesInMemberAccount(t *testing.T) {
	tests := []struct {
		testTitle     string
		client        *mockIRDSInstancesAPI
		action        string
		expectedCount RDSInstanceCount
	}{
		{
			testTitle: "RDS testing Test action",
			client: &mockIRDSInstancesAPI{
				DescribeDBInstancesOutput: &rds.DescribeDBInstancesOutput{
					DBInstances: []rdstype.DBInstance{
						// RDS instance-scheduling = default, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("default"),
								},
							},
						},
						// no RDS instance-scheduling tag set, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-2"),
						},
						// RDS instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-7"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-scheduling"),
								},
							},
						},
						// RDS instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-3"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String(""),
								},
							},
						},
						// RDS instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-4"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("invalid-value"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-stop, therefore skip auto stop,  skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-5"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-stop"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-6"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-start"),
								},
							},
						},
					},
				},
			},
			action:        "test",
			expectedCount: RDSInstanceCount{4, 3},
		},
		{
			testTitle: "RDS testing Stop action",
			client: &mockIRDSInstancesAPI{
				DescribeDBInstancesOutput: &rds.DescribeDBInstancesOutput{
					DBInstances: []rdstype.DBInstance{
						// RDS instance-scheduling = default, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("default"),
								},
							},
						},
						// no RDS instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-2"),
						},
						// RDS instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
						{
							DBInstanceIdentifier: aws.String("i-2162279001"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-scheduling"),
								},
							},
						},
						// RDS instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-3"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String(""),
								},
							},
						},
						// RDS instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-4"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("invalid-value"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-stop, therefore skip auto stop, skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-5"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-stop"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-start, therefore skip auto start, but not stop, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-6"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-start"),
								},
							},
						},
					},
				},
			},
			action:        "stop",
			expectedCount: RDSInstanceCount{5, 2},
		},
		{
			testTitle: "RDS testing Start action",
			client: &mockIRDSInstancesAPI{
				DescribeDBInstancesOutput: &rds.DescribeDBInstancesOutput{
					DBInstances: []rdstype.DBInstance{
						// instance-scheduling = default, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("default"),
								},
							},
						},
						// no RDS instance-scheduling and no aws:autoscaling:groupName tags, therefore schedule an instance, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-2"),
						},
						// RDS instance-scheduling = skip-scheduling, therefore skip scheduling, skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-7"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-scheduling"),
								},
							},
						},
						// RDS instance-scheduling is set to an empty string, therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-3"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String(""),
								},
							},
						},
						// RDS instance-scheduling = "invalid-value", therefore ignore the tag and auto schedule, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-4"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("invalid-value"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-stop, therefore skip auto stop, but not start, acted upon: 1
						{
							DBInstanceIdentifier: aws.String("test-database-5"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-stop"),
								},
							},
						},
						// RDS instance-scheduling = skip-auto-start, therefore skip auto start, skipped: 1
						{
							DBInstanceIdentifier: aws.String("test-database-6"),
							TagList: []rdstype.Tag{
								{
									Key:   aws.String("instance-scheduling"),
									Value: aws.String("skip-auto-start"),
								},
							},
						},
					},
				},
			},
			action:        "start",
			expectedCount: RDSInstanceCount{5, 2},
		},
	}

	for _, subtest := range tests {
		t.Run(subtest.testTitle, func(t *testing.T) {
			actualInstanceCount := StopStartTestRDSInstancesInMemberAccount(subtest.client, subtest.action)
			if want, got := subtest.expectedCount, actualInstanceCount; want != *got {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}
