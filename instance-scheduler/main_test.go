package main

import (
	"github.com/aws/aws-lambda-go/events"
	"os"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Testing the lambda function", func(t *testing.T) {
		os.Setenv("INSTANCE_SCHEDULING_ACTION", "Test")

		_, err := handler(events.APIGatewayProxyRequest{})
		if err != nil {
			t.Fatal("[ERROR]:", err)
		}
	})
}
