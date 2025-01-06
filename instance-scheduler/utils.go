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

func getNonProductionAccounts(environments string) map[string]string {
    accounts := make(map[string]string)

    // Fetch the list of in-scope environments from modernisation-platform/environments
    repoOwner := "ministryofjustice"
    repoName := "modernisation-platform"
    branch := "instance-scheduler-skip"
    directory := "environments"
    records, err := FetchDirectory(repoOwner, repoName, branch, directory)
    if err != nil {
        log.Fatalf("getNonProductionAccounts - Failed to fetch directory listing from GitHub: %v", err)
    }

    // Split the records string into a slice of strings
    recordSlice := strings.Split(records, ",")

    // Parse the environments secret into a json object
    var allAccounts map[string]interface{}
    json.Unmarshal([]byte(environments), &allAccounts)

    // Iterate over the fetched records and include environments based on the fetched list
    log.Printf("getNonProductionAccounts - Iterating over the fetched JSON from environments")
    for _, record := range allAccounts {
        if rec, ok := record.(map[string]interface{}); ok {
            for key, val := range rec {
                // Include if the account's name is in the fetched list
                if contains(recordSlice, key) {
                    accounts[key] = val.(string)
                    log.Printf("getNonProductionAccounts - Added account to list:", key)
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

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
