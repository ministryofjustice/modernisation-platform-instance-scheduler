package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

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

func CreateSSMClient(config aws.Config) ISSMGetParameter {
	return ssm.NewFromConfig(config)
}

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

func CreateSecretManagerClient(config aws.Config) ISecretManagerGetSecretValue {
	return secretsmanager.NewFromConfig(config)
}

func getNonProductionAccounts(environments string, skipAccountNames string) map[string]string {
	accounts := make(map[string]string)
    log.Printf("Skip Accounts: %s", skipAccountNames)
	var allAccounts map[string]interface{}
	json.Unmarshal([]byte(environments), &allAccounts)

	for _, record := range allAccounts {
		if rec, ok := record.(map[string]interface{}); ok {
			for key, val := range rec {
				// Skip if the account's name ends with "-production", for example: performance-hub-production will be skipped
				if !strings.HasSuffix(key, "-production") && (len(skipAccountNames) < 1 || !strings.Contains(skipAccountNames, key)) {
					log.Printf("INFO: Account Name: %s", val.(string))
					accounts[key] = val.(string)
				}
			}
		}
	}
	return accounts
}

func parseAction(action string) (string, error) {
	log.Printf("Action=%v\n", action)
	actionAsLower := strings.ToLower(action)

	switch actionAsLower {
	case "test", "start", "stop":
		return actionAsLower, nil
	}
	return "", errors.New("ERROR: Invalid Action. Must be one of 'start' 'stop' 'test'")
}

func LoadDefaultConfig() (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
}
