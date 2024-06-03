package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func mockLoadDefaultConfig() (aws.Config, error) {
	return *aws.NewConfig(), errors.New("Mock Error!")
}

func TestHandlerUnit(t *testing.T) {
	t.Run("returns 400 error when `InstanceSchedulingRequest.Action` is invalid", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{loadDefaultConfig: mockLoadDefaultConfig}
		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "Invalid Action! 😱"})
		assert.Equal(t, response.StatusCode, 400)
		assert.NotNil(t, err)
	})

	t.Run("returns 500 error when default config cannot load", func(t *testing.T) {
		instanceScheduler := InstanceScheduler{loadDefaultConfig: mockLoadDefaultConfig}
		response, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "test"})
		assert.Equal(t, response.StatusCode, 500)
		assert.NotNil(t, err)
	})
}
