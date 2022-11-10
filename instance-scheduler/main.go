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
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const INSTANCE_SCHEDULER_VERSION string = "1.1.8"

/*
CLI examples:
aws ssm get-parameter --name environment_management_arn --with-decryption --profile core-shared-services-production --region eu-west-2
aws secretsmanager get-secret-value --secret-id environment_management --profile mod --region eu-west-2
*/

/*
Interface that defines the set of Amazon SSM API operations required by the getParameter function.
ISSMGetParameter is satisfied byt the Amazon SSM client's GetParameter method.
*/
type ISSMGetParameter interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func getParameter(client ISSMGetParameter, parameterName string) string {
	result, err := client.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	return *result.Parameter.Value
}

/*
Interface that defines the set of Amazon secretsmanager API operations required by the getSecret function.
ISecretManagerGetSecretValue is satisfied byt the Amazon secretsmanager client's GetSecretValue method.
*/
type ISecretManagerGetSecretValue interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func getSecret(client ISecretManagerGetSecretValue, secretId string) string {
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretId),
		VersionStage: aws.String("AWSCURRENT"),
	})

	if err != nil {
		log.Fatal(err)
	}
	return *result.SecretString
}

func getNonProductionAccounts(environments string, skipAccountNames string) map[string]string {
	accounts := make(map[string]string)

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
	roleARN := fmt.Sprintf("arn:aws:iam::%v:role/InstanceSchedulerAccess", accountId)
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	ec2Client := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{}
	_, err := ec2Client.DescribeInstances(context.TODO(), input)
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

type InstanceSchedulingRequest struct {
	Action string `json:"action"`
}

func handler(ctx context.Context, request InstanceSchedulingRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("BEGIN: Instance scheduling v%v\n", INSTANCE_SCHEDULER_VERSION)
	log.Printf("Action=%v\n", request.Action)
	skipAccounts := os.Getenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS")
	log.Printf("INSTANCE_SCHEDULING_SKIP_ACCOUNTS=%v\n", skipAccounts)

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatal(err)
	}

	// Get the secret ID (ARN) of the environment management secret from the parameter store
	ssmClient := ssm.NewFromConfig(cfg)
	secretId := getParameter(ssmClient, "environment_management_arn")

	// Get the environment management secret that holds the account IDs
	secretsManagerClient := secretsmanager.NewFromConfig(cfg)
	environments := getSecret(secretsManagerClient, secretId)

	accounts := getNonProductionAccounts(environments, skipAccounts)
	memberAccountNames := []string{}
	nonMemberAccountNames := []string{}
	for accName, accId := range accounts {
		ec2Client := getEc2ClientForMemberAccount(cfg, accName, accId)
		if ec2Client == nil {
			nonMemberAccountNames = append(nonMemberAccountNames, accName)
		} else {
			memberAccountNames = append(memberAccountNames, accName)
			log.Printf("BEGIN: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
			stopStartTestInstancesInMemberAccount(ec2Client, request.Action)
			log.Printf("END: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
		}
	}

	if len(memberAccountNames) > 0 {
		log.Printf("END: Instance scheduling for %v member accounts: %v\n", len(memberAccountNames), memberAccountNames)
	} else {
		log.Println("WARN: END: Instance scheduling: No member account was found!")
	}
	if len(nonMemberAccountNames) > 0 {
		log.Printf("Ignored %v non-member accounts lacking InstanceSchedulerAccess role: %v\n", len(nonMemberAccountNames), nonMemberAccountNames)
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Action=%v\n", request.Action),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
