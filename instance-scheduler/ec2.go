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
	ec2type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
	if action == "stop" {
		return stopEc2Instances(client)
	}
	if action == "start" {
		return startEc2Instances(client)
	}
	if action == "test" {
		return testEc2Instances(client)
	}
	log.Fatalf("Invalid action: [ %v ]", action)
	return nil
}

func parseInstanceTags(instance ec2type.Instance, skippedInstances []string, skippedAutoScaledInstances []string) (string, bool, []string, []string) {
	var instanceSchedulingTag string
	isPartOfAutoScalingGroup := false
	isSkipSchedulingTag := false
	for _, tag := range instance.Tags {
		if *tag.Key == "aws:autoscaling:groupName" {
			log.Printf("Skip instance because aws:autoscaling:groupName tag because it is part of an Auto Scaling group\n")
			skippedAutoScaledInstances = append(skippedAutoScaledInstances, *instance.InstanceId)
			isPartOfAutoScalingGroup = true
		}
		if *tag.Key == "instance-scheduling" {
			instanceSchedulingTag = *tag.Value
		}
		if *tag.Key == "instance-scheduling" && *tag.Value == "skip-scheduling" {
			log.Printf("Skip instance because instance-scheduling tag having value 'skip-scheduling'\n")
			skippedInstances = append(skippedInstances, *instance.InstanceId)
			isSkipSchedulingTag = true
		}
	}

	isSkippable := bool(isPartOfAutoScalingGroup || isSkipSchedulingTag)
	return instanceSchedulingTag, isSkippable, skippedInstances, skippedAutoScaledInstances
}

func startEc2Instances(client IEC2InstancesAPI) *InstanceCount {
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Println("ERROR: Could not retrieve information about Amazon EC2 instances in member account:", err)
		return &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	}

	instancesActedUpon := []string{}
	skippedInstances := []string{}
	skippedAutoScaledInstances := []string{}
	for _, r := range result.Reservations {
		log.Printf("INFO: Reservation ID: [ %v ]\n", *r.ReservationId)
		for _, i := range r.Instances {
			log.Printf("INFO: Instance ID: [ %v ]\n", *i.InstanceId)
			instanceSchedulingTag, skipInstance, skippedInstancesModified, skippedAutoScaledInstancesModified := parseInstanceTags(i, skippedInstances, skippedAutoScaledInstances)
			skippedInstances = skippedInstancesModified
			skippedAutoScaledInstances = skippedAutoScaledInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-stop" {
				log.Printf("INFO: Skipped instance because instance-scheduling tag having value 'skip-auto-stop'\n")
				skippedInstances = append(skippedInstances, *i.InstanceId)
				continue
			}

			log.Printf("INFO: Stopped instance because instance-scheduling tag is absent\n")
			instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
			startInstance(client, *i.InstanceId)
		}
	}

	log.Printf("INFO: Started %v instances: %v\n", len(instancesActedUpon), instancesActedUpon)
	log.Printf("INFO: Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)
	log.Printf("INFO: Skipped %v instances due to aws:autoscaling:groupName tag: %v\n", len(skippedAutoScaledInstances), skippedAutoScaledInstances)

	return &InstanceCount{actedUpon: len(instancesActedUpon), skipped: len(skippedInstances), skippedAutoScaled: len(skippedAutoScaledInstances)}
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

func stopEc2Instances(client IEC2InstancesAPI) *InstanceCount {
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Println("ERROR: Could not retrieve information about Amazon EC2 instances in member account:", err)
		return &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	}

	instancesActedUpon := []string{}
	skippedInstances := []string{}
	skippedAutoScaledInstances := []string{}
	for _, r := range result.Reservations {
		log.Printf("INFO: Reservation ID: [ %v ]\n", *r.ReservationId)
		for _, i := range r.Instances {
			log.Printf("INFO: Instance ID: [ %v ]\n", *i.InstanceId)
			instanceSchedulingTag, skipInstance, skippedInstancesModified, skippedAutoScaledInstancesModified := parseInstanceTags(i, skippedInstances, skippedAutoScaledInstances)
			skippedInstances = skippedInstancesModified
			skippedAutoScaledInstances = skippedAutoScaledInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-stop" {
				log.Printf("INFO: Skipped instance because instance-scheduling tag having value 'skip-auto-stop'\n")
				skippedInstances = append(skippedInstances, *i.InstanceId)
				continue
			}

			log.Printf("INFO: Stopped instance because instance-scheduling tag is absent\n")
			instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
			stopInstance(client, *i.InstanceId)
		}
	}

	log.Printf("INFO: Stopped %v instances: %v\n", len(instancesActedUpon), instancesActedUpon)
	log.Printf("INFO: Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)
	log.Printf("INFO: Skipped %v instances due to aws:autoscaling:groupName tag: %v\n", len(skippedAutoScaledInstances), skippedAutoScaledInstances)

	return &InstanceCount{actedUpon: len(instancesActedUpon), skipped: len(skippedInstances), skippedAutoScaled: len(skippedAutoScaledInstances)}
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

func testEc2Instances(client IEC2InstancesAPI) *InstanceCount {
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Println("ERROR: Could not retrieve information about Amazon EC2 instances in member account:", err)
		return &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	}

	instancesActedUpon := []string{}
	skippedInstances := []string{}
	skippedAutoScaledInstances := []string{}
	for _, r := range result.Reservations {
		log.Printf("INFO: Reservation ID: [ %v ]\n", *r.ReservationId)
		for _, i := range r.Instances {
			log.Printf("INFO: Instance ID: [ %v ]\n", *i.InstanceId)
			instanceSchedulingTag, skipInstance, skippedInstancesModified, skippedAutoScaledInstancesModified := parseInstanceTags(i, skippedInstances, skippedAutoScaledInstances)
			skippedInstances = skippedInstancesModified
			skippedAutoScaledInstances = skippedAutoScaledInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-stop" || instanceSchedulingTag == "skip-auto-start" {
				log.Printf("INFO: Skipped instance because instance-scheduling tag having value 'skip-auto-start' or \n")
				skippedInstances = append(skippedInstances, *i.InstanceId)
				continue
			}
			instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
			log.Printf("INFO: Successfully tested instance with Id %v\n", *i.InstanceId)
			continue
		}
	}

	log.Printf("INFO: Tested %v instances: %v\n", len(instancesActedUpon), instancesActedUpon)
	log.Printf("INFO: Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)
	log.Printf("INFO: Skipped %v instances due to aws:autoscaling:groupName tag: %v\n", len(skippedAutoScaledInstances), skippedAutoScaledInstances)

	return &InstanceCount{actedUpon: len(instancesActedUpon), skipped: len(skippedInstances), skippedAutoScaled: len(skippedAutoScaledInstances)}
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
