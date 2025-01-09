package main

import (
	"encoding/json"
	//"errors"
	//"log"
	"strings"
	"fmt"
    "net/http"
    "net/url"
    "io"
    "github.com/tidwall/gjson"

)

// Additional functions that parse json data from the environments directory obtail the full list of in-scope non-prod environments.

// JSONFileContent represents the structure of the JSON content
type JSONFileContent map[string]interface{}

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

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    var content JSONFileContent
    err = json.Unmarshal(body, &content)
	  if err != nil {
		  return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	  }

    return content, nil
}

// fetches the environments JSON data from GitHub
func fetchGitHubData(baseURL, repoOwner, repoName, branch, directory string) ([]byte, error) {

    u, err := url.Parse(path.Join(baseURL, repoOwner, repoName, "contents", directory))
    if err != nil {
        return nil, fmt.Errorf("failed to parse URL: %w", err)
    }

    query := u.Query()
    query.Set("ref", branch)
    u.RawQuery = query.Encode()

    // Print the constructed URL for debugging
    fmt.Println("Constructed URL:", u.String())

    // Create a new HTTP request
    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }

    // Set the User-Agent header (GitHub API requires a User-Agent header)
    req.Header.Set("User-Agent", "Go-http-client")

    // Perform the HTTP request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch directory listing: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    return body, nil
}

// processGitHubData processes the JSON data and returns a slice of GitHubFile
func processGitHubData(body []byte) ([]GitHubFile, error) {
    var files []GitHubFile
    err = json.Unmarshal(body, &files)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
    }
    return files, nil
}


// Helper function to check if instance_scheduler_skip exists and is true
func hasInstanceSchedulerSkip(content gjson.Result) bool {
    skip := content.Get("instance_scheduler_skip")
    return skip.Exists() && skip.String() == "true"
}

// extractNames finds all "name" elements in the "environments" array, excluding those with instance_scheduler_skip or production
func extractNames(content JSONFileContent, envName string) []string {

    var names []string
    jsonData, err := json.Marshal(content)

    if err != nil {
        fmt.Println("Failed to marshal JSON content:", err)
        return names
    }

    environments := gjson.GetBytes(jsonData, "environments")

    environments.ForEach(func(_, env gjson.Result) bool {
        name := env.Get("name").String()
        if name == "" {
            return true // continue
        }

        if hasInstanceSchedulerSkip(env) {
            fmt.Println("extractNames - Skipping due to instance_scheduler_skip:", envName+"."+name)
            return true // continue
        }

        if name == "production" {
            fmt.Println("extractNames - Skipping due to production:", envName+"."+name)
            return true // continue
        }

        fmt.Println("extractNames - Found name:", envName+"."+name)
        names = append(names, name)
        return true // continue
    })

    return names
}