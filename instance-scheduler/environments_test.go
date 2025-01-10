// This file contains unit tests for each of the functions in environments.go.
// These functions are written using go's built-in testing package, and the testify library.

package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/tidwall/gjson"
)

// Unit test for FetchJSON
func TestFetchJSON(t *testing.T) {
    // Create a mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Write a mock response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"key1": "value1", "key2": "value2"}`))
    }))
    defer server.Close()

    // Call FetchJSON with the mock server URL
    url := server.URL
    content, err := FetchJSON(url)
    assert.NoError(t, err)
    assert.NotNil(t, content)

    // Validate the content of the JSON response
    assert.Equal(t, "value1", content["key1"])
    assert.Equal(t, "value2", content["key2"])
}

// Unit test for fetchGitHubData
func TestFetchGitHubData(t *testing.T) {
    // Define dummy values for the parameters
    repoOwner := "dummyOwner"
    repoName := "dummyRepo"
    branch := "dummyBranch"
    directory := "dummyDirectory"
    expectedURL := "/dummyOwner/dummyRepo/contents/dummyDirectory?ref=dummyBranch"

    // Define the expected JSON response
    expectedJSON := `[{"name": "file1.json", "path": "path/to/file1.json", "type": "file"}]`

    // Create a mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check the request URL
        assert.Equal(t, expectedURL, r.URL.String())
        // Write a mock response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(expectedJSON))
    }))
    defer server.Close()

    // Override the base URL to point to the mock server
    baseURL := server.URL

    // Call fetchGitHubData with the mock server URL
    body, err := fetchGitHubData(baseURL, repoOwner, repoName, branch, directory)
    assert.NoError(t, err)
    assert.NotNil(t, body)

    // Check that the response contains valid JSON
    var jsonResponse []map[string]interface{}
    err = json.Unmarshal(body, &jsonResponse)
    assert.NoError(t, err)
    assert.NotEmpty(t, jsonResponse)

    // Ensure the JSON received back is the same as the expected JSON
    receivedJSON, err := json.Marshal(jsonResponse)
    assert.NoError(t, err)
    assert.JSONEq(t, expectedJSON, string(receivedJSON))

    // Optionally, print the response for manual inspection
    t.Logf("Response: %s", string(body))
}


// Unit test for processGitHubData
func TestProcessGitHubData(t *testing.T) {
    // Define a mock JSON response
    mockJSON := []byte(`[{"name": "file1.json", "path": "path/to/file1.json", "type": "file"}, {"name": "file2.json", "path": "path/to/file2.json", "type": "file"}]`)

    // Call processGitHubData with the mock JSON
    files, err := processGitHubData(mockJSON)
    assert.NoError(t, err)
    assert.NotNil(t, files)
    assert.Len(t, files, 2)

    // Validate the content of the files
    assert.Equal(t, "file1.json", files[0].Name)
    assert.Equal(t, "path/to/file1.json", files[0].Path)
    assert.Equal(t, "file", files[0].Type)

    assert.Equal(t, "file2.json", files[1].Name)
    assert.Equal(t, "path/to/file2.json", files[1].Path)
    assert.Equal(t, "file", files[1].Type)
}


// Unit test for hasInstanceSchedulerSkip
func TestHasInstanceSchedulerSkip(t *testing.T) {
    testCases := []struct {
        name     string
        json     string
        expected bool
    }{
        {
            name:     "Skip is true",
            json:     `{"instance_scheduler_skip": "true"}`,
            expected: true,
        },
        {
            name:     "Skip is false",
            json:     `{"instance_scheduler_skip": "false"}`,
            expected: false,
        },
        {
            name:     "Skip is missing",
            json:     `{}`,
            expected: false,
        },
        {
            name:     "True is missing",
            json:     `{"instance_scheduler_skip": 123}`,
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := hasInstanceSchedulerSkip(gjson.Parse(tc.json))
            assert.Equal(t, tc.expected, result)
        })
    }
}


// Unit test for extractNames
func TestExtractNames(t *testing.T) {

    mockJSONContent := JSONFileContent{
        "environments": []interface{}{
            map[string]interface{}{
                "name": "development",
                "instance_scheduler_skip": []interface{}{"true"},
            },
            map[string]interface{}{
                "name": "test",
                "instance_scheduler_skip": []interface{}{"false"},
            },
            map[string]interface{}{
                "name": "preproduction",
            },
            map[string]interface{}{
                "name": "production",
                "instance_scheduler_skip": []interface{}{"false"},
            },
        },
    }

    envName := "env"
    expectedNames := []string{"test", "preproduction"}

    // Call the extractNames function
    names := extractNames(mockJSONContent, envName)

    // Assert that the returned names match the expected names
    assert.Equal(t, expectedNames, names, "The extracted names should match the expected names")
}