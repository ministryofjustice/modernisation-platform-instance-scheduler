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
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const INSTANCE_SCHEDULER_VERSION string = "1.2.1"

/*
CLI examples:
aws ssm get-parameter --name environment_management_arn --with-decryption --profile core-shared-services-production --region eu-west-2
aws secretsmanager get-secret-value --secret-id environment_management --profile mod --region eu-west-2
*/

// ISSMGetParameter
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

// ISecretManagerGetSecretValue
/*
Interface that defines the set of Amazon secretsmanager API operations required by the getSecret function.
ISecretManagerGetSecretValue is satisfied by the Amazon secretsmanager client's GetSecretValue method.
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

type InstanceCount struct {
	actedUpon         int
	skipped           int
	skippedAutoScaled int
}

// IEC2InstancesAPI
/*
Interface that defines the set of Amazon EC2 API operations required by the startInstance, stopInstance and stopStartTestInstancesInMemberAccount
functions.
IEC2InstancesAPI is satisfied by the Amazon EC2 client's StopInstances,  StartInstances and DescribeInstances methods.
*/
type IEC2InstancesAPI interface {
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// IRDSInstancesAPI
/*
Interface that defines the set of Amazon RDS API operations required by the startInstance, stopInstance and stopStartTestInstancesInMemberAccount
functions.
IRDSInstancesAPI is satisfied by the Amazon RDS client's StopInstances,  StartInstances and DescribeInstances methods.
*/

type IRDSInstancesAPI interface {
	StopDBInstance(ctx context.Context, params *rds.StopDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StopDBInstanceOutput, error)
	StartDBInstance(ctx context.Context, params *rds.StartDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StartDBInstanceOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

func stopStartTestInstancesInMemberAccount(client IEC2InstancesAPI, rdsClient IRDSInstancesAPI, action string) *InstanceCount {
	action = strings.ToLower(action)
	count := &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	switch action {
	case "test", "start", "stop":
		break
	default:
		log.Print("ERROR: Invalid Action. Must be one of 'start' 'stop' 'test'")
		return count
	}

	// Handle EC2 instances
	ec2Result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Print("ERROR: Could not retrieve information about Amazon EC2 instances in member account:\n", err)
		return count
	}

	for _, r := range ec2Result.Reservations {
		for _, i := range r.Instances {
			// Existing EC2 instance tag-based processing code
			// ...
			// Apply start or stop actions as needed
			if instanceSchedulingTag == "skip-scheduling" {
				if action == "stop" {
					stopInstance(client, *i.InstanceId)
				} else if action == "start" {
					startInstance(client, *i.InstanceId)
				}
			}
		}
	}

	// Handle RDS instances
	rdsResult, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Print("ERROR: Could not retrieve information about Amazon RDS instances in member account:\n", err)
		return count
	}

	for _, instance := range rdsResult.DBInstances {
		// Check tags, skip auto-scaled instances, etc.
		// Apply start or stop actions as needed
		if *instance.DBClusterIdentifier == "skip-scheduling" {
			// Apply start or stop actions as needed
			if action == "stop" {
				stopRDSInstance(rdsClient, *instance.DBInstanceIdentifier)
			} else if action == "start" {
				startRDSInstance(rdsClient, *instance.DBInstanceIdentifier)
			}
		}
	}

	// Return the instance count
	return count
}

func getEc2ClientForMemberAccount(cfg aws.Config, accountName string, accountId string) (*ec2.Client, IRDSInstancesAPI) {
	roleARN := fmt.Sprintf("arn:aws:iam::%v:role/InstanceSchedulerAccess", accountId)
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	ec2Client := ec2.NewFromConfig(cfg)
	rdsClient := rds.NewFromConfig(cfg)

	// Check if the account has permissions to describe instances (both EC2 and RDS)
	ec2Input := &ec2.DescribeInstancesInput{}
	_, ec2Err := ec2Client.DescribeInstances(context.TODO(), ec2Input)
	if ec2Err != nil {
		if strings.Contains(ec2Err.Error(), "is not authorized to perform: ec2:DescribeInstances") {
			log.Printf("WARN: account %v (%v) lacks permissions to describe EC2 instances\n", accountName, accountId)
		} else {
			log.Fatal(ec2Err)
		}
	}

	rdsInput := &rds.DescribeDBInstancesInput{}
	_, rdsErr := rdsClient.DescribeDBInstances(context.TODO(), rdsInput)
	if rdsErr != nil {
		if strings.Contains(rdsErr.Error(), "is not authorized to perform: rds:DescribeDBInstances") {
			log.Printf("WARN: account %v (%v) lacks permissions to describe RDS instances\n", accountName, accountId)
		} else {
			log.Fatal(rdsErr)
		}
	}

	return ec2Client, rdsClient
}

type InstanceSchedulingRequest struct {
	Action string `json:"action"`
}

type InstanceSchedulingResponse struct {
	Action                string   `json:"action"`
	MemberAccountNames    []string `json:"member_account_names"`
	NonMemberAccountNames []string `json:"non_member_account_names"`
	ActedUpon             int      `json:"acted_upon"`
	Skipped               int      `json:"skipped"`
	SkippedAutoScaled     int      `json:"skipped_auto_scaled"`
}

func handler(request InstanceSchedulingRequest) (events.APIGatewayProxyResponse, error) {
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
	totalCount := &InstanceCount{actedUpon: 0, skipped: 0, skippedAutoScaled: 0}
	for accName, accId := range accounts {
		ec2Client, rdsClient := getEc2ClientForMemberAccount(cfg, accName, accId)
		if ec2Client == nil || rdsClient == nil {
			nonMemberAccountNames = append(nonMemberAccountNames, accName)
		} else {
			memberAccountNames = append(memberAccountNames, accName)
			log.Printf("BEGIN: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
			count := stopStartTestInstancesInMemberAccount(ec2Client, rdsClient, request.Action)
			totalCount.actedUpon += count.actedUpon
			totalCount.skipped += count.skipped
			totalCount.skippedAutoScaled += count.skippedAutoScaled
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

	body := &InstanceSchedulingResponse{
		Action:                request.Action,
		MemberAccountNames:    memberAccountNames,
		NonMemberAccountNames: nonMemberAccountNames,
		ActedUpon:             totalCount.actedUpon,
		Skipped:               totalCount.skipped,
		SkippedAutoScaled:     totalCount.skippedAutoScaled,
	}
	bodyJson, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		Body:       string(bodyJson),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
