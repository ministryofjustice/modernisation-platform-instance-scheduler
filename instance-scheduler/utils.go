package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"fmt"
    "net/http"
    "net/url"

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

	// Fetch the list of in-scope environments from modernisation-platform/environments
	repoOwner := "ministryofjustice"
	repoName := "modernisation-platform"
	branch := "instance-scheduler-skip"
	directory := "environments"
	records, err := FetchDirectory(repoOwner, repoName, branch, directory)
	if err != nil {
		log.Fatalf("Failed to fetch directory listing from GitHub: %v", err)
	}

    // Parse the environments secret into a json object
    var allAccounts map[string]interface{}
    json.Unmarshal([]byte(environments), &allAccounts)

    // Iterate over the fetched records and include environments based on the fetched list
    for _, record := range allAccounts {
        if rec, ok := record.(map[string]interface{}); ok {
            for key, val := range rec {
                // Include if the account's name is in the fetched list and does not end with "-production"
                if !strings.HasSuffix(key, "-production") && contains(records, key) {
                    accounts[key] = val.(string)
					fmt.Println("Added account:", val)
					
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


// Additional functions that parse json data from the environments directory obtail the full list of in-scope non-prod environments.

// GitHubFile represents a single file in a GitHub directory as returned by the GitHub API
type GitHubFile struct {
    Name string `json:"name"`
    Path string `json:"path"`
    Type string `json:"type"` // "file" or "dir"
}

// JSONFileContent represents the structure of each JSON file's content
type JSONFileContent map[string]interface{}

// FetchJSON fetches the JSON content from a given URL
func FetchJSON(rawURL string) (JSONFileContent, error) {
    resp, err := http.Get(rawURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch JSON content: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    var content JSONFileContent
    if err := json.Unmarshal(body, &content); err != nil {
        return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
    }

    return content, nil
}

// FetchDirectory fetches the list of files in a GitHub directory using the GitHub API
func FetchDirectory(repoOwner, repoName, branch, directory string) (string, error) {
    baseURL := "https://api.github.com/repos"
    u, err := url.Parse(baseURL)
    if err != nil {
        return "", fmt.Errorf("failed to parse base URL: %w", err)
    }

    u.Path = strings.Join([]string{u.Path, repoOwner, repoName, "contents", directory}, "/")
    query := u.Query()
    query.Set("ref", branch)
    u.RawQuery = query.Encode()

    // Print the constructed URL for debugging
    fmt.Println("Constructed URL:", u.String())

    // Create a new HTTP request
    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        return "", fmt.Errorf("failed to create HTTP request: %w", err)
    }

    // Set the User-Agent header (GitHub API requires a User-Agent header)
    req.Header.Set("User-Agent", "Go-http-client")

    // Perform the HTTP request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to fetch directory listing: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response body: %w", err)
    }

    var files []GitHubFile
    if err := json.Unmarshal(body, &files); err != nil {
        return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
    }

    var result []string
    for _, file := range files {
        // Only process JSON files
        if file.Type == "file" && strings.HasSuffix(file.Name, ".json") {
            rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", repoOwner, repoName, branch, file.Path)
            content, err := FetchJSON(rawURL)
            if err != nil {
                fmt.Println("Error fetching", rawURL, ":", err)
                continue
            }

            if accountType, ok := content["account-type"]; ok && accountType == "member" {
                fileNameWithoutExt := strings.TrimSuffix(file.Name, ".json")
                names := extractNames(content, fileNameWithoutExt)
                for _, name := range names {
                    finalName := fmt.Sprintf("%s-%s", fileNameWithoutExt, name)
                    result = append(result, finalName)
                    fmt.Println("Processed file:", file.Name, "Result:", finalName)
                }
            }
        }
    }

    finalResult := strings.Join(result, ",")
    fmt.Println("Final comma-delimited list:", finalResult)
    return finalResult, nil
}

// extractNames recursively finds all "name" elements in the JSON content
func extractNames(content JSONFileContent, envName string) []string {
    var names []string
    for key, value := range content {
        if key == "name" {
            if nameStr, ok := value.(string); ok && nameStr != "production" {
                // Check for instance_scheduler_skip
                if skip, ok := content["instance_scheduler_skip"]; ok {
                    if skipArray, ok := skip.([]interface{}); ok {
                        for _, skipValue := range skipArray {
                            if skipStr, ok := skipValue.(string); ok && skipStr == "true" {
                                fmt.Println("Skipping environment due to instance_scheduler_skip: " + envName + "." + nameStr)
                                continue
                            }
                        }
                    }
                }
                names = append(names, nameStr)
            }
        } else if nestedContent, ok := value.(map[string]interface{}); ok {
            nestedNames := extractNames(nestedContent, envName)
            names = append(names, nestedNames...)
        } else if nestedArray, ok := value.([]interface{}); ok {
            for _, item := range nestedArray {
                if itemMap, ok := item.(map[string]interface{}); ok {
                    nestedNames := extractNames(itemMap, envName)
                    names = append(names, nestedNames...)
                }
            }
        }
    }
    return names
}

