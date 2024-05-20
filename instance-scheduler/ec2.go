package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

type InstanceCount struct {
	actedUpon         int
	skipped           int
	skippedAutoScaled int
}

type IEC2InstancesAPI interface {
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func stopStartTestInstancesInMemberAccount(client IEC2InstancesAPI, action string) *InstanceCount {
	action = strings.ToLower(action)
	count := &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	switch action {
	case "test", "start", "stop":
		break
	default:
		log.Print("ERROR: Invalid Action. Must be one of 'start' 'stop' 'test'")
		return count
	}
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})

	if err != nil {
		log.Println("ERROR: Could not retrieve information about Amazon EC2 instances in member account:", err)
		return count
	}

	instancesActedUpon := []string{}
	skippedInstances := []string{}
	skippedAutoScaledInstances := []string{}
	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			var instanceSchedulingTag string
			instanceIsPartOfAutoScalingGroup := false
			for _, tag := range i.Tags {
				if *tag.Key == "aws:autoscaling:groupName" {
					instanceIsPartOfAutoScalingGroup = true
					break
				} else if *tag.Key == "instance-scheduling" {
					instanceSchedulingTag = *tag.Value
				}
			}

			instanceSchedulingTagDescr := fmt.Sprintf("with instance-scheduling tag having value '%v'", instanceSchedulingTag)
			if instanceSchedulingTag == "" {
				instanceSchedulingTagDescr = "with instance-scheduling tag being absent"
			}

			actedUponMessage := fmt.Sprintf("%v instance %v (ReservationId: %v) %v\n", action, *i.InstanceId, *r.ReservationId, instanceSchedulingTagDescr)
			skippedMessage := fmt.Sprintf("Skipped instance %v (ReservationId: %v) %v\n", *i.InstanceId, *r.ReservationId, instanceSchedulingTagDescr)
			skippedAutoScaledMessage := fmt.Sprintf("Skipped instance %v (ReservationId: %v) with aws:autoscaling:groupName tag because it is part of an Auto Scaling group\n", *i.InstanceId, *r.ReservationId)

			// Tag key: instance-scheduling
			// Valid values: default (same as absence of tag), skip-scheduling, skip-auto-stop, skip-auto-start

			if instanceIsPartOfAutoScalingGroup {
				log.Print(skippedAutoScaledMessage)
				skippedAutoScaledInstances = append(skippedAutoScaledInstances, *i.InstanceId)
			} else if instanceSchedulingTag == "skip-scheduling" {
				log.Print(skippedMessage)
				skippedInstances = append(skippedInstances, *i.InstanceId)
			} else if instanceSchedulingTag == "skip-auto-stop" {
				if action == "stop" {
					log.Print(skippedMessage)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				} else if action == "start" {
					log.Print(actedUponMessage)
					instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
					startInstance(client, *i.InstanceId)
				} else if action == "test" {
					log.Printf("Successfully tested skipping instance with Id %v\n", *i.InstanceId)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				}
			} else if instanceSchedulingTag == "skip-auto-start" {
				if action == "stop" {
					log.Print(actedUponMessage)
					instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
					stopInstance(client, *i.InstanceId)
				} else if action == "start" {
					log.Print(skippedMessage)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				} else if action == "test" {
					log.Printf("Successfully tested skipping instance with Id %v\n", *i.InstanceId)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				}
			} else { // if instance-scheduling tag is missing, or the value of the tag either default, not valid or empty the instance will be actioned
				log.Print(actedUponMessage)
				instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
				if action == "stop" {
					stopInstance(client, *i.InstanceId)
				} else if action == "start" {
					startInstance(client, *i.InstanceId)
				} else if action == "test" {
					log.Printf("Successfully tested instance with Id %v\n", *i.InstanceId)
				}
			}

		}
	}

	acted := "Started"
	if action == "stop" {
		acted = "Stopped"
	} else if action == "test" {
		acted = "Tested"
	}
	if len(instancesActedUpon) > 0 {
		log.Printf("%v %v instances: %v\n", acted, len(instancesActedUpon), instancesActedUpon)
		count.actedUpon = len(instancesActedUpon)
	} else {
		log.Printf("WARN: No instances found to %v!\n", action)
	}
	if len(skippedInstances) > 0 {
		log.Printf("Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)
		count.skipped = len(skippedInstances)
	}
	if len(skippedAutoScaledInstances) > 0 {
		log.Printf("Skipped %v instances due to aws:autoscaling:groupName tag: %v\n", len(skippedAutoScaledInstances), skippedAutoScaledInstances)
		count.skippedAutoScaled = len(skippedAutoScaledInstances)
	}
	return count
}

func startInstance(client IEC2InstancesAPI, instanceId string) {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
		DryRun: aws.Bool(true),
	}

	_, err := client.StartInstances(context.TODO(), input)

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
		log.Println("User has permission to start an instance.")
		input.DryRun = aws.Bool(false)
		_, err = client.StartInstances(context.TODO(), input)
	}

	if err == nil {
		log.Printf("Successfully started instance with Id %v\n", instanceId)
	} else {
		log.Printf("ERROR: Could not start instance: %v\n", err)
	}
}

func stopInstance(client IEC2InstancesAPI, instanceId string) {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
		DryRun: aws.Bool(true),
	}

	_, err := client.StopInstances(context.TODO(), input)

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
		log.Println("User has permission to stop instances.")
		input.DryRun = aws.Bool(false)
		_, err = client.StopInstances(context.TODO(), input)
	}

	if err == nil {
		log.Printf("Successfully stopped instance with Id %v\n", instanceId)
	} else {
		log.Printf("ERROR: Could not stop instance: %v\n", err)
	}
}

func getEc2ClientForMemberAccount(cfg aws.Config, accountName string, accountId string) IEC2InstancesAPI {
	roleARN := fmt.Sprintf("arn:aws:iam::%v:role/InstanceSchedulerAccess", accountId)
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)
	ec2Input := &ec2.DescribeInstancesInput{}
	_, err := ec2Client.DescribeInstances(context.TODO(), ec2Input)
	if err != nil {
		if strings.Contains(err.Error(), "is not authorized to perform: sts:AssumeRole on resource") {
			log.Printf("WARN: account %v (%v) is ignored because it does not have the role InstanceSchedulerAccess, therefore is not a member account\n", accountName, accountId)
			return nil
		} else {
			log.Fatal(err)
		}
	}

	return ec2Client
}
