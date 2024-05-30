package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstype "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type RDSInstanceCount struct {
	RDSActedUpon int
	RDSSkipped   int
}

type IRDSInstancesAPI interface {
	StopDBInstance(ctx context.Context, params *rds.StopDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StopDBInstanceOutput, error)
	StartDBInstance(ctx context.Context, params *rds.StartDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StartDBInstanceOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

func StopStartTestRDSInstancesInMemberAccount(RDSClient IRDSInstancesAPI, action string) *RDSInstanceCount {
	action = strings.ToLower(action)

	if action == "stop" {
		result, err := RDSClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
		if err != nil {
			log.Print("ERROR: Could not retrieve information about Amazon RDS instances in member account:\n", err)
			return &RDSInstanceCount{RDSActedUpon: 0, RDSSkipped: 0}
		}

		instancesActedUpon := []string{}
		skippedInstances := []string{}

		for _, RDSInstance := range result.DBInstances {
			log.Printf("INFO: RDS Instance Identifier: [ %v ]\n", *RDSInstance.DBInstanceIdentifier)
			instanceSchedulingTag, skipInstance, skippedInstancesModified := parseRDSInstanceTags(RDSInstance, skippedInstances)
			skippedInstances = skippedInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-stop" {
				skippedInstances = append(skippedInstances, *RDSInstance.DBInstanceIdentifier)
				log.Printf("INFO: Skipped RDS instance because instance-scheduling tag having value 'skip-auto-stop'\n")
				continue
			}

			instancesActedUpon = append(instancesActedUpon, *RDSInstance.DBInstanceIdentifier)
			stopRDSInstance(RDSClient, *RDSInstance.DBInstanceIdentifier)
			log.Printf("INFO: Stopped RDS instance because instance-scheduling tag is absent\n")
			continue
		}

		log.Printf("INFO: Stopped %v instances: %v\n", len(instancesActedUpon), instancesActedUpon)
		log.Printf("INFO: Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)

		return &RDSInstanceCount{RDSActedUpon: len(instancesActedUpon), RDSSkipped: len(skippedInstances)}
	}

	if action == "start" {
		result, err := RDSClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
		if err != nil {
			log.Print("ERROR: Could not retrieve information about Amazon RDS instances in member account:\n", err)
			return &RDSInstanceCount{RDSActedUpon: 0, RDSSkipped: 0}
		}

		instancesActedUpon := []string{}
		skippedInstances := []string{}

		for _, RDSInstance := range result.DBInstances {
			log.Printf("INFO: RDS Instance Identifier: [ %v ]\n", *RDSInstance.DBInstanceIdentifier)
			instanceSchedulingTag, skipInstance, skippedInstancesModified := parseRDSInstanceTags(RDSInstance, skippedInstances)
			skippedInstances = skippedInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-start" {
				skippedInstances = append(skippedInstances, *RDSInstance.DBInstanceIdentifier)
				log.Printf("INFO: Skipped RDS instance because instance-scheduling tag having value 'skip-auto-start'\n")
				continue
			}

			instancesActedUpon = append(instancesActedUpon, *RDSInstance.DBInstanceIdentifier)
			startRDSInstance(RDSClient, *RDSInstance.DBInstanceIdentifier)
			log.Printf("INFO: Started RDS instance because instance-scheduling tag is absent\n")
			continue
		}

		log.Printf("INFO: Started %v RDS instances: %v\n", len(instancesActedUpon), instancesActedUpon)
		log.Printf("INFO: Skipped %v RDS instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)

		return &RDSInstanceCount{RDSActedUpon: len(instancesActedUpon), RDSSkipped: len(skippedInstances)}

	}

	if action == "test" {
		result, err := RDSClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
		if err != nil {
			log.Print("ERROR: Could not retrieve information about Amazon RDS instances in member account:\n", err)
			return &RDSInstanceCount{RDSActedUpon: 0, RDSSkipped: 0}
		}

		instancesActedUpon := []string{}
		skippedInstances := []string{}

		for _, RDSInstance := range result.DBInstances {
			log.Printf("INFO: RDS Instance Identifier: [ %v ]\n", *RDSInstance.DBInstanceIdentifier)
			instanceSchedulingTag, skipInstance, skippedInstancesModified := parseRDSInstanceTags(RDSInstance, skippedInstances)
			skippedInstances = skippedInstancesModified

			if skipInstance {
				continue
			}

			if instanceSchedulingTag == "skip-auto-stop" || instanceSchedulingTag == "skip-auto-start" {
				skippedInstances = append(skippedInstances, *RDSInstance.DBInstanceIdentifier)
				log.Printf("INFO: Skipped RDS instance with DB instance identifier %v because instance-scheduling tag having value 'skip-auto-stop' or 'skip-auto-start'", *RDSInstance.DBInstanceIdentifier)
				continue
			}

			instancesActedUpon = append(instancesActedUpon, *RDSInstance.DBInstanceIdentifier)
			log.Printf("INFO: Successfully tested RDS instance with DB instance identifier %v because instance-scheduling tag is absent\n")
			continue
		}

		log.Printf("INFO: Started %v RDS instances: %v\n", len(instancesActedUpon), instancesActedUpon)
		log.Printf("INFO: Skipped %v RDS instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)

		return &RDSInstanceCount{RDSActedUpon: len(instancesActedUpon), RDSSkipped: len(skippedInstances)}
	}

	log.Fatalf("Invalid action: [ %v ]", action)
	return nil
}

func parseRDSInstanceTags(instance rdstype.DBInstance, RDSskippedInstances []string) (string, bool, []string) {
	var instanceSchedulingTag string
	isSkipSchedulingTag := false
	for _, tag := range instance.TagList {
		if *tag.Key == "instance-scheduling" {
			instanceSchedulingTag = *tag.Value
		}
		if *tag.Key == "instance-scheduling" && *tag.Value == "skip-scheduling" {
			log.Printf("Skip instance because instance-scheduling tag having value 'skip-scheduling'\n")
			RDSskippedInstances = append(RDSskippedInstances, *instance.DBInstanceIdentifier)
			isSkipSchedulingTag = true
		}
	}

	return instanceSchedulingTag, isSkipSchedulingTag, RDSskippedInstances
}

func startRDSInstance(client IRDSInstancesAPI, dbInstanceIdentifier string) {
	input := &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	_, err := client.StartDBInstance(context.TODO(), input)
	if err == nil {
		log.Printf("Successfully started RDS instance with Identifier %v\n", dbInstanceIdentifier)
	} else {
		log.Printf("ERROR: Could not start RDS instance: %v\n", err)
	}
}

func stopRDSInstance(client IRDSInstancesAPI, dbInstanceIdentifier string) {
	input := &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	_, err := client.StopDBInstance(context.TODO(), input)
	if err == nil {
		log.Printf("Successfully stopped RDS instance with Identifier %v\n", dbInstanceIdentifier)
	} else {
		log.Printf("ERROR: Could not stop RDS instance: %v\n", err)
	}
}

func getRDSClientForMemberAccount(cfg aws.Config, accountName string, accountId string) IRDSInstancesAPI {
	roleARN := fmt.Sprintf("arn:aws:iam::%v:role/InstanceSchedulerAccess", accountId)
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	// Create RDS client
	rdsClient := rds.NewFromConfig(cfg)
	rdsInput := &rds.DescribeDBInstancesInput{}
	_, rdsErr := rdsClient.DescribeDBInstances(context.TODO(), rdsInput)
	if rdsErr != nil {
		if strings.Contains(rdsErr.Error(), "is not authorized to perform: sts:AssumeRole on resource") {
			log.Printf("WARN: account %v (%v) is ignored because it does not have the role InstanceSchedulerAccess, therefore is not a member account\n", accountName, accountId)
			return nil
		} else {
			log.Fatal(rdsErr)
		}
	}
	return rdsClient
}
