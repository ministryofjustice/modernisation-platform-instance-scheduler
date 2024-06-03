package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const INSTANCE_SCHEDULER_VERSION string = "1.2.1"

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
	RDSActedUpon          int      `json:"rds_acted_upon"`
	RDSSkipped            int      `json:"rds_skipped"`
}

type InstanceScheduler struct {
	loadDefaultConfig func() (aws.Config, error)
}

func (instanceScheduler *InstanceScheduler) handler(request InstanceSchedulingRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("INFO: Starting Instance Scheduling...")

	instanceSchedulingResponse := &InstanceSchedulingResponse{
		Action:                request.Action,
		MemberAccountNames:    []string{},
		NonMemberAccountNames: []string{},
		ActedUpon:             0,
		Skipped:               0,
		SkippedAutoScaled:     0,
		RDSActedUpon:          0,
		RDSSkipped:            0,
	}

	action, err := parseAction(request.Action)
	if err != nil {
		body, _ := json.Marshal(instanceSchedulingResponse)
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: 400,
		}, err
	}

	cfg, err := instanceScheduler.loadDefaultConfig()
	if err != nil {
		body, _ := json.Marshal(instanceSchedulingResponse)
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: 500,
		}, err
	}

	skipAccounts := os.Getenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS")
	log.Printf("INSTANCE_SCHEDULING_SKIP_ACCOUNTS=%v\n", skipAccounts)

	// Get the secret ID (ARN) of the environment management secret from the parameter store
	ssmClient := ssm.NewFromConfig(cfg)
	secretId := getParameter(ssmClient, "environment_management_arn")

	// Get the environment management secret that holds the account IDs
	secretsManagerClient := secretsmanager.NewFromConfig(cfg)
	environments := getSecret(secretsManagerClient, secretId)

	accounts := getNonProductionAccounts(environments, skipAccounts)
	for accName, accId := range accounts {
		ec2Client := getEc2ClientForMemberAccount(cfg, accName, accId)
		rdsClient := getRDSClientForMemberAccount(cfg, accName, accId)

		if ec2Client == nil || rdsClient == nil {
			instanceSchedulingResponse.NonMemberAccountNames = append(instanceSchedulingResponse.NonMemberAccountNames, accName)
		} else {
			instanceSchedulingResponse.MemberAccountNames = append(instanceSchedulingResponse.MemberAccountNames, accName)
			log.Printf("INFO: Instance scheduling for member account: accountName=%v\n", accName)
			count := stopStartTestInstancesInMemberAccount(ec2Client, action)
			instanceSchedulingResponse.ActedUpon += count.actedUpon
			instanceSchedulingResponse.Skipped += count.skipped
			instanceSchedulingResponse.SkippedAutoScaled += count.skippedAutoScaled

			rdsCount := StopStartTestRDSInstancesInMemberAccount(rdsClient, action)
			instanceSchedulingResponse.RDSActedUpon += rdsCount.RDSActedUpon
			instanceSchedulingResponse.RDSSkipped += rdsCount.RDSSkipped
		}
	}

	log.Printf("INFO: Instance scheduling for %v member accounts: %v\n", len(instanceSchedulingResponse.MemberAccountNames), instanceSchedulingResponse.MemberAccountNames)
	log.Printf("INFO: Ignored %v non-member accounts lacking InstanceSchedulerAccess role: %v\n", len(instanceSchedulingResponse.NonMemberAccountNames), instanceSchedulingResponse.NonMemberAccountNames)

	body, _ := json.Marshal(instanceSchedulingResponse)
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}, nil
}

func main() {
	InstanceScheduler := InstanceScheduler{loadDefaultConfig: LoadDefaultConfig}
	lambda.Start(InstanceScheduler.handler)
}
