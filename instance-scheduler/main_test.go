package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerUnit(t *testing.T) {
	t.Run("returns 400 error when `InstanceSchedulingRequest.Action` is invalid", func(t *testing.T) {
		response, err := handler(InstanceSchedulingRequest{Action: "Invalid Action! 😱"})
		assert.Equal(t, response.StatusCode, 400)
		assert.NotNil(t, err)
	})
}
