package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const INSTANCE_SCHEDULER_VERSION string = "1.0"

/*
ENV variable INSTANCE_SCHEDULING_SKIP_ACCOUNTS: A comma-separated list of account names to be skipped from instance scheduling. For example:
"xhibit-portal-development,another-development,".
As can be observed in the example above, every account name needs a leading comma, hence the last comma in the list.

Use the same scheduling as for bastion instances:

stop = "0 20 * * *" # 20.00 UTC time or 21.00 London time
start = "0 5 * * *" # 5.00 UTC time or 6.00 London time

CLI examples:
aws secretsmanager get-secret-value --secret-id environment_management --profile mod --region eu-west-2
*/

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

func getSecret(cfg aws.Config, secretId string) string {
	client := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretId),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err)
	}
	return *result.SecretString
}

// This function does not return the expected value at the moment. Will be refactored and used in the future.
func getParameter(cfg aws.Config, parameterName string) string {
	client := ssm.NewFromConfig(cfg)
	input := &ssm.GetParameterInput{
		Name: aws.String(parameterName),
	}
	result, err := client.GetParameter(context.TODO(), input)
	if err != nil {
		log.Fatal(err)
	}
	return *result.Parameter.Value
}

func getNonProductionAccounts(cfg aws.Config, skipAccountNames string) map[string]string {
	accounts := make(map[string]string)
	// Get accounts secret
	environments := getSecret(cfg, os.Getenv("INSTANCE_SCHEDULING_ENVIRONMENT_MANAGEMENT_SECRET_ID"))

	var allAccounts map[string]interface{}
	json.Unmarshal([]byte(environments), &allAccounts)

	for _, record := range allAccounts {
		if rec, ok := record.(map[string]interface{}); ok {
			for key, val := range rec {
				// Skip if the account's name ends with "-production", for example: performance-hub-production will be skipped
				if !strings.HasSuffix(key, "-production") && (len(skipAccountNames) < 1 || !strings.Contains(skipAccountNames, key)) {
					accounts[key] = val.(string)
				}
			}
		}
	}
	return accounts
}

func stopInstance(client *ec2.Client, instanceId string) {
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

func startInstance(client *ec2.Client, instanceId string) {
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

func stopStartTestInstancesInMemberAccount(client *ec2.Client, action string) {
	input := &ec2.DescribeInstancesInput{}

	result, err := client.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Print("ERROR: Could not retrieve information about Amazon EC2 instances in member account:\n", err)
		return
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

			if instanceIsPartOfAutoScalingGroup {
				log.Print(skippedAutoScaledMessage)
				skippedAutoScaledInstances = append(skippedAutoScaledInstances, *i.InstanceId)
			} else if (instanceSchedulingTag == "") || (instanceSchedulingTag == "default") {
				// Tag key: instance-scheduling
				// Valid values: default (same as absence of tag), skip-scheduling, skip-auto-stop, skip-auto-start
				log.Print(actedUponMessage)
				instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
				if action == "Stop" {
					stopInstance(client, *i.InstanceId)
				} else if action == "Start" {
					startInstance(client, *i.InstanceId)
				} else if action == "Test" {
					log.Printf("Successfully tested instance with Id %v\n", *i.InstanceId)
				}
			} else if instanceSchedulingTag == "skip-scheduling" {
				log.Print(skippedMessage)
				skippedInstances = append(skippedInstances, *i.InstanceId)
			} else if instanceSchedulingTag == "skip-auto-stop" {
				if action == "Stop" {
					log.Print(skippedMessage)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				} else if action == "Start" {
					log.Print(actedUponMessage)
					instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
					startInstance(client, *i.InstanceId)
				} else if action == "Test" {
					log.Printf("Successfully tested instance with Id %v\n", *i.InstanceId)
				}
			} else if instanceSchedulingTag == "skip-auto-start" {
				if action == "Stop" {
					log.Print(actedUponMessage)
					instancesActedUpon = append(instancesActedUpon, *i.InstanceId)
					stopInstance(client, *i.InstanceId)
				} else if action == "Start" {
					log.Print(skippedMessage)
					skippedInstances = append(skippedInstances, *i.InstanceId)
				} else if action == "Test" {
					log.Printf("Successfully tested instance with Id %v\n", *i.InstanceId)
				}
			}
		}
	}

	acted := "Started"
	if action == "Stop" {
		acted = "Stopped"
	} else if action == "Test" {
		acted = "Tested"
	}
	if len(instancesActedUpon) > 0 {
		log.Printf("%v %v instances: %v\n", acted, len(instancesActedUpon), instancesActedUpon)
	} else {
		log.Printf("WARN: No instances found to %v!\n", action)
	}
	if len(skippedInstances) > 0 {
		log.Printf("Skipped %v instances due to instance-scheduling tag: %v\n", len(skippedInstances), skippedInstances)
	}
	if len(skippedAutoScaledInstances) > 0 {
		log.Printf("Skipped %v instances due to aws:autoscaling:groupName tag: %v\n", len(skippedAutoScaledInstances), skippedAutoScaledInstances)
	}
}

func getEc2ClientForMemberAccount(cfg aws.Config, accountName string, accountId string) *ec2.Client {
	roleARN := fmt.Sprintf("arn:aws:iam::%v:role/MemberInfrastructureAccess", accountId)
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	iamClient := iam.NewFromConfig(cfg)
	_, err := iamClient.ListRoles(context.TODO(), &iam.ListRolesInput{
		PathPrefix: aws.String("/")})

	if err != nil {
		if strings.Contains(err.Error(), "is not authorized to perform: sts:AssumeRole on resource") {
			log.Printf("WARN: account %v (%v) is ignored because it does not have the role MemberInfrastructureAccess, therefore is not a member account\n", accountName, accountId)
			return nil
		} else {
			log.Fatal(err)
		}
	}
	return ec2.NewFromConfig(cfg)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	action := os.Getenv("INSTANCE_SCHEDULING_ACTION")
	skipAccounts := os.Getenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS")
	log.Printf("BEGIN: Instance scheduling v%v\n", INSTANCE_SCHEDULER_VERSION)
	log.Printf("INSTANCE_SCHEDULING_ACTION=%v\n", action)
	log.Printf("INSTANCE_SCHEDULING_SKIP_ACCOUNTS=%v\n", skipAccounts)
	log.Printf("INSTANCE_SCHEDULING_ENVIRONMENT_MANAGEMENT_SECRET_ID=%v\n", os.Getenv("INSTANCE_SCHEDULING_ENVIRONMENT_MANAGEMENT_SECRET_ID"))

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatal(err)
	}

	accounts := getNonProductionAccounts(cfg, skipAccounts)
	memberAccountNames := []string{}
	nonMemberAccountNames := []string{}
	for accName, accId := range accounts {
		ec2Client := getEc2ClientForMemberAccount(cfg, accName, accId)
		if ec2Client == nil {
			nonMemberAccountNames = append(nonMemberAccountNames, accName)
		} else {
			memberAccountNames = append(memberAccountNames, accName)
			log.Printf("BEGIN: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
			stopStartTestInstancesInMemberAccount(ec2Client, action)
			log.Printf("END: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
		}
	}

	if len(memberAccountNames) > 0 {
		log.Printf("END: Instance scheduling for %v member accounts: %v\n", len(memberAccountNames), memberAccountNames)
	} else {
		log.Println("WARN: END: Instance scheduling: No member account was found!")
	}
	if len(nonMemberAccountNames) > 0 {
		log.Printf("Ignored %v non-member accounts lacking MemberInfrastructureAccess role: %v\n", len(nonMemberAccountNames), nonMemberAccountNames)
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("INSTANCE_SCHEDULING_ACTION=%v\n", action),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
