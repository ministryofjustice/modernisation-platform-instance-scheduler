package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Test request", func(t *testing.T) {
		result, err := handler(context.TODO(), InstanceSchedulingRequest{Action: "Test"})
		if err != nil {
			t.Fatalf("Failed to run lambda's handler: %v", err)
		}
		res := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(result.Body), &res)
		assert.NotEmpty(t, res.MemberAccountNames, "No member account was found")
		for _, accountName := range res.MemberAccountNames {
			assert.False(t, strings.HasSuffix(accountName, "-production"), fmt.Sprintf("Production account %v was found in the list of member accounts. Production accounts should be skipped.", accountName))
		}
	})
}
