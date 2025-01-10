package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"fmt"

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
    baseURL := "https://api.github.com/repos"
    repoOwner := "ministryofjustice"
    repoName := "modernisation-platform"
    branch := "main"
    directory := "environments"

    // Step 1: Fetch the JSON data from GitHub
    body, err := fetchGitHubData(baseURL, repoOwner, repoName, branch, directory)
    if err != nil {
        log.Fatalf("getNonProductionAccounts - Failed to fetch directory listing from GitHub: %v", err)
    }

    // Step 2: Process the JSON data
    files, err := processGitHubData(body)
    if err != nil {
        log.Fatalf("getNonProductionAccounts - Failed to process GitHub data: %v", err)
    }

	// Step 3: Iterate through returned files, check the JSON of each file and obtain a list of accounts to be inlcuded by the scheduler
    var result []string

    for _, file := range files {
        // Only process JSON files
        if file.Type == "file" && strings.HasSuffix(file.Name, ".json") {
            fmt.Println("**** Processing file:", file.Name)
            rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", repoOwner, repoName, branch, file.Path)
			// The extracted json is held in the content
			content, err := FetchJSON(rawURL)
            if err != nil {
                fmt.Println("Error fetching", rawURL, ":", err)
                continue
            }
            if accountType, ok := content["account-type"]; ok {
                // Check whether the account is of type "member". We want to exclude all accounts types that are not member.
                if accountType == "member" {
                    fileNameWithoutExt := strings.TrimSuffix(file.Name, ".json")
					fmt.Println("Account is of type member:", fileNameWithoutExt)
					// This returns a list of accounts for each environment that filters out 1) Production accounts, and 2) Those accounts with the instance_scheduler_skip flag.
                    names := extractNames(content, fileNameWithoutExt)
					// Avoids returning an empty list as there may be member environments that have no accounts to be included in the scheduler.
					if len(names) == 0 {
                        fmt.Println("No names extracted, skipping file:", file.Name)
                        continue
                    }
					// Adds the environment-name.account-name to the list.
                    for _, name := range names {
                        finalName := fmt.Sprintf("%s-%s", fileNameWithoutExt, name)
                        result = append(result, finalName)
                    }
                }
            }
        }
    }

    // Split the records string into a slice of strings
    recordSlice := strings.Split(strings.Join(result, ","), ",")

    // Parse the environments secret into a json object
    var allAccounts map[string]interface{}
    json.Unmarshal([]byte(environments), &allAccounts)

    // This checks the secret of account names & numbers against those from "result" above to get definative list of numbers to be included in the scheduler run.
    log.Printf("getNonProductionAccounts - Iterating over the fetched JSON from environments")
    for _, record := range allAccounts {
        if rec, ok := record.(map[string]interface{}); ok {
            for key, val := range rec {
                // Include if the account's name is in the fetched list
                if contains(recordSlice, key) {
                    accounts[key] = val.(string)
                    fmt.Println("getNonProductionAccounts - Added account to list:", key)
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


