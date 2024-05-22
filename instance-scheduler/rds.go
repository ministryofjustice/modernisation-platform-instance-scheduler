package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/rds"
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

func StopStartTestRDSInstancesInMemberAccount(rdsClient IRDSInstancesAPI, action string) *RDSInstanceCount {
	action = strings.ToLower(action)
	rdscount := &RDSInstanceCount{RDSActedUpon: 0, RDSSkipped: 0}

	result, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Print("ERROR: Could not retrieve information about Amazon RDS instances in member account:\n", err)
		return rdscount
	}

	RDSinstancesActedUpon := []string{}
	RDSskippedInstances := []string{}

	for _, rdsInstance := range result.DBInstances {
		var instanceSchedulingTag string
		for _, tag := range rdsInstance.TagList {
			if *tag.Key == "instance-scheduling" {
				instanceSchedulingTag = *tag.Value
				break
			}
		}

		instanceSchedulingTagDescr := fmt.Sprintf("with instance-scheduling tag having value '%v'", instanceSchedulingTag)
		if instanceSchedulingTag == "" {
			instanceSchedulingTagDescr = "with instance-scheduling tag being absent"
		}

		actedUponMessage := fmt.Sprintf("%v instance %v %v\n", action, *rdsInstance.DBInstanceIdentifier, instanceSchedulingTagDescr)
		skippedMessage := fmt.Sprintf("Skipped instance %v %v\n", *rdsInstance.DBInstanceIdentifier, instanceSchedulingTagDescr)

		// Tag key: instance-scheduling
		// Valid values: default (same as absence of tag), skip-scheduling, skip-auto-stop, skip-auto-start

		if instanceSchedulingTag == "skip-scheduling" {
			log.Print(skippedMessage)
			RDSskippedInstances = append(RDSskippedInstances, *rdsInstance.DBInstanceIdentifier)
		} else if instanceSchedulingTag == "skip-auto-stop" {
			if action == "stop" {
				log.Print(skippedMessage)
				RDSskippedInstances = append(RDSskippedInstances, *rdsInstance.DBInstanceIdentifier)
			} else if action == "start" {
				log.Print(actedUponMessage)
				RDSinstancesActedUpon = append(RDSinstancesActedUpon, *rdsInstance.DBInstanceIdentifier)
				startRDSInstance(rdsClient, *rdsInstance.DBInstanceIdentifier)
			} else if action == "test" {
				log.Printf("Successfully tested skipping instance with Id %v\n", *rdsInstance.DBInstanceIdentifier)
				RDSskippedInstances = append(RDSskippedInstances, *rdsInstance.DBInstanceIdentifier)
			}
		} else if instanceSchedulingTag == "skip-auto-start" {
			if action == "stop" {
				log.Print(actedUponMessage)
				RDSinstancesActedUpon = append(RDSinstancesActedUpon, *rdsInstance.DBInstanceIdentifier)
				stopRDSInstance(rdsClient, *rdsInstance.DBInstanceIdentifier)
			} else if action == "start" {
				log.Print(skippedMessage)
				RDSskippedInstances = append(RDSskippedInstances, *rdsInstance.DBInstanceIdentifier)
			} else if action == "test" {
				log.Printf("Successfully tested skipping instance with Id %v\n", *rdsInstance.DBInstanceIdentifier)
				RDSskippedInstances = append(RDSskippedInstances, *rdsInstance.DBInstanceIdentifier)
			}

		} else { // if instance-scheduling tag is missing, or the value of the tag either default, not valid or empty the RDS instance will be actioned
			log.Print(actedUponMessage)
			RDSinstancesActedUpon = append(RDSinstancesActedUpon, *rdsInstance.DBInstanceIdentifier)

			if action == "stop" {
				stopRDSInstance(rdsClient, *rdsInstance.DBInstanceIdentifier)
			} else if action == "start" {
				startRDSInstance(rdsClient, *rdsInstance.DBInstanceIdentifier)
			} else if action == "test" {
				log.Printf("Successfully tested RDS instance with Id %v\n", *rdsInstance.DBInstanceIdentifier)
			}
		}
	}

	acted := "Started"
	if action == "stop" {
		acted = "Stopped"
	} else if action == "test" {
		acted = "Tested"
	}
	if len(RDSinstancesActedUpon) > 0 {
		log.Printf("%v %v RDS instances: %v\n", acted, len(RDSinstancesActedUpon), RDSinstancesActedUpon)
		rdscount.RDSActedUpon = len(RDSinstancesActedUpon)
	} else {
		log.Printf("WARN: No RDS instances found to %v!\n", action)
	}
	if len(RDSskippedInstances) > 0 {
		log.Printf("Skipped %v RDS instances due to instance-scheduling tag: %v\n", len(RDSskippedInstances), RDSskippedInstances)
		rdscount.RDSSkipped = len(RDSskippedInstances)
	}

	return rdscount
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
