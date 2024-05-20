package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
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
	totalCount := &InstanceSchedulingResponse{
		Action:                request.Action,
		MemberAccountNames:    []string{},
		NonMemberAccountNames: []string{},
		ActedUpon:             0,
		Skipped:               0,
		SkippedAutoScaled:     0,
		RDSActedUpon:          0,
		RDSSkipped:            0,
	}
	for accName, accId := range accounts {
		ec2Client := getEc2ClientForMemberAccount(cfg, accName, accId)
		rdsClient := getRDSClientForMemberAccount(cfg, accName, accId)

		if ec2Client == nil || rdsClient == nil {
			nonMemberAccountNames = append(nonMemberAccountNames, accName)
		} else {
			memberAccountNames = append(memberAccountNames, accName)
			log.Printf("BEGIN: Instance scheduling for member account: accountName=%v, accountId=%v\n", accName, accId)
			count := stopStartTestInstancesInMemberAccount(ec2Client, request.Action)
			totalCount.ActedUpon += count.actedUpon
			totalCount.Skipped += count.skipped
			totalCount.SkippedAutoScaled += count.skippedAutoScaled

			rdsCount := StopStartTestRDSInstancesInMemberAccount(rdsClient, request.Action)
			totalCount.RDSActedUpon += rdsCount.RDSActedUpon
			totalCount.RDSSkipped += rdsCount.RDSSkipped

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
		ActedUpon:             totalCount.ActedUpon,
		Skipped:               totalCount.Skipped,
		SkippedAutoScaled:     totalCount.SkippedAutoScaled,
		RDSActedUpon:          totalCount.RDSActedUpon,
		RDSSkipped:            totalCount.RDSSkipped,
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
