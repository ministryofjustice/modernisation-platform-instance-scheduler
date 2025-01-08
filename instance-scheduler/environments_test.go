package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// Unit test for hasInstanceSchedulerSkip
func TestHasInstanceSchedulerSkip(t *testing.T) {
    // Define test cases
    testCases := []struct {
        name     string
        content  JSONFileContent
        expected bool
    }{
        {
            name: "Skip is true",
            content: JSONFileContent{
                "instance_scheduler_skip": []interface{}{"true"},
            },
            expected: true,
        },
        {
            name: "Skip is false",
            content: JSONFileContent{
                "instance_scheduler_skip": []interface{}{"false"},
            },
            expected: false,
        },
        {
            name: "Skip is missing",
            content: JSONFileContent{
                "some_other_key": "some_value",
            },
            expected: false,
        },
        {
            name: "Skip is empty",
            content: JSONFileContent{
                "instance_scheduler_skip": []interface{}{},
            },
            expected: false,
        },
        {
            name: "Skip is not a string",
            content: JSONFileContent{
                "instance_scheduler_skip": []interface{}{123},
            },
            expected: false,
        },
    }

    // Run test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := hasInstanceSchedulerSkip(tc.content)
            assert.Equal(t, tc.expected, result)
        })
    }
}


// Mock JSON content for testing extractNames
var mockJSONContent = JSONFileContent{
    "environments": []interface{}{
        map[string]interface{}{
            "name": "development",
            "instance_scheduler_skip": []interface{}{"false"},
        },
        map[string]interface{}{
            "name": "test",
            "instance_scheduler_skip": []interface{}{"true"},
        },
        map[string]interface{}{
            "name": "preproduction",
            "instance_scheduler_skip": []interface{}{"false"},
        },
        map[string]interface{}{
            "name": "production",
            "instance_scheduler_skip": []interface{}{"false"},
        },
    },
}

func TestExtractNames(t *testing.T) {
    envName := "env"
    expectedNames := []string{"development", "preproduction"}

    // Call the extractNames function
    names := extractNames(mockJSONContent, envName)

    // Assert that the returned names match the expected names
    assert.Equal(t, expectedNames, names, "The extracted names should match the expected names")
}