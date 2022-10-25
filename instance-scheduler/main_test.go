package main

import (
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Testing the lambda function", func(t *testing.T) {
		_, err := handler(events.APIGatewayProxyRequest{})
		if err != nil {
			t.Error("[ERROR]:", err)
		} else {
			t.Log("[INFO]:", "Test has run successfully!")
		}
	})
}
