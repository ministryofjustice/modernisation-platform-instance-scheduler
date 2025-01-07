package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// Mock JSON content for testing
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