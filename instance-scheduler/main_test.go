
package main

import (
	"context"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Testing the lambda function", func(t *testing.T) {
		_, err := handler(context.TODO(), InstanceSchedulingRequest{Action: "Test"})
		if err != nil {
			t.Fatal("Everything should be ok")
		}
	})
}
