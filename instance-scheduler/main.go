package main

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	LoadDefaultConfig                        func() (aws.Config, error)
	CreateSSMClient                          func(aws.Config) ISSMGetParameter
	GetParameter                             func(client ISSMGetParameter, parameterName string) string
	CreateSecretManagerClient                func(cfg aws.Config) ISecretManagerGetSecretValue
	GetSecret                                func(client ISecretManagerGetSecretValue, secretId string) string
	GetEc2ClientForMemberAccount             func(cfg aws.Config, accountName string, accountId string) IEC2InstancesAPI
	GetRDSClientForMemberAccount             func(cfg aws.Config, accountName string, accountId string) IRDSInstancesAPI
	StopStartTestInstancesInMemberAccount    func(client IEC2InstancesAPI, action string) *InstanceCount
	StopStartTestRDSInstancesInMemberAccount func(RDSClient IRDSInstancesAPI, action string) *RDSInstanceCount
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

	cfg, err := instanceScheduler.LoadDefaultConfig()
	if err != nil {
		body, _ := json.Marshal(instanceSchedulingResponse)
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: 500,
		}, err
	}

	// skipAccounts := instanceScheduler.GetEnv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS")
	// log.Printf("INSTANCE_SCHEDULING_SKIP_ACCOUNTS=%v\n", skipAccounts)

	ssmClient := instanceScheduler.CreateSSMClient(cfg)
	secretId := instanceScheduler.GetParameter(ssmClient, "environment_management_arn")

	secretsManagerClient := instanceScheduler.CreateSecretManagerClient(cfg)
	environments := instanceScheduler.GetSecret(secretsManagerClient, secretId)

	accounts := getNonProductionAccounts(environments)
	for accName, accId := range accounts {
		ec2Client := instanceScheduler.GetEc2ClientForMemberAccount(cfg, accName, accId)
		rdsClient := instanceScheduler.GetRDSClientForMemberAccount(cfg, accName, accId)

		if ec2Client == nil || rdsClient == nil {
			instanceSchedulingResponse.NonMemberAccountNames = append(instanceSchedulingResponse.NonMemberAccountNames, accName)
		} else {
			instanceSchedulingResponse.MemberAccountNames = append(instanceSchedulingResponse.MemberAccountNames, accName)
			log.Printf("INFO: Instance scheduling for member account: accountName=%v\n", accName)

			count := instanceScheduler.StopStartTestInstancesInMemberAccount(ec2Client, action)
			instanceSchedulingResponse.ActedUpon += count.actedUpon
			instanceSchedulingResponse.Skipped += count.skipped
			instanceSchedulingResponse.SkippedAutoScaled += count.skippedAutoScaled

			rdsCount := instanceScheduler.StopStartTestRDSInstancesInMemberAccount(rdsClient, action)
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
	InstanceScheduler := InstanceScheduler{
		LoadDefaultConfig:                        LoadDefaultConfig,
		CreateSSMClient:                          CreateSSMClient,
		GetParameter:                             getParameter,
		CreateSecretManagerClient:                CreateSecretManagerClient,
		GetSecret:                                getSecret,
		GetEc2ClientForMemberAccount:             getEc2ClientForMemberAccount,
		GetRDSClientForMemberAccount:             getRDSClientForMemberAccount,
		StopStartTestInstancesInMemberAccount:    stopStartTestInstancesInMemberAccount,
		StopStartTestRDSInstancesInMemberAccount: StopStartTestRDSInstancesInMemberAccount,
	}
	lambda.Start(InstanceScheduler.handler)
}
