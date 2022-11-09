package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("Test request", func(t *testing.T) {
		// Accounts mi-platform-development and analytical-platform-data-development cause the main_int_test.go to fail because they appear
		// to lack the InstanceSchedulerAccess role. For now, I have excluded them from the test via the env variable INSTANCE_SCHEDULING_SKIP_ACCOUNTS
		os.Setenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS", "mi-platform-development,analytical-platform-data-development,")

		result, err := handler(context.TODO(), InstanceSchedulingRequest{Action: "Test"})
		if err != nil {
			t.Fatalf("Failed to run lambda's handler: %v", err)
		}
		res := InstanceSchedulingResponse{}
		json.Unmarshal([]byte(result.Body), &res)
		assert.Equal(t, res.Action, "Test", "Response action does not match requested action")
		assert.NotEmpty(t, res.MemberAccountNames, "No member account was found")
		for _, accountName := range res.MemberAccountNames {
			assert.False(t, strings.HasSuffix(accountName, "-production"), fmt.Sprintf("Production account %v was found in the list of member accounts. Production accounts should be skipped.", accountName))
			assert.True(t, strings.HasSuffix(accountName, "-development") ||
				strings.HasSuffix(accountName, "-test") ||
				strings.HasSuffix(accountName, "-preproduction"), fmt.Sprintf("Unexpected suffix in member account %v", accountName))
		}
		assert.NotEmpty(t, res.NonMemberAccountNames, "No member account was found")
		for _, accountName := range res.NonMemberAccountNames {
			assert.False(t, strings.HasSuffix(accountName, "-production"), fmt.Sprintf("Production account %v was found in the list of non-member accounts. Production accounts should be skipped.", accountName))
			if !strings.HasPrefix(accountName, "core-vpc-") {
				assert.False(t, strings.HasSuffix(accountName, "-development"), fmt.Sprintf("Non-member account %v was found with suffix '-development'. Accounts with such suffix are member accounts.", accountName))
				assert.False(t, strings.HasSuffix(accountName, "-test"), fmt.Sprintf("Non-member account %v was found with suffix '-test'. Accounts with such suffix are member accounts.", accountName))
				assert.False(t, strings.HasSuffix(accountName, "-preproduction"), fmt.Sprintf("Non-member account %v was found with suffix '-preproduction'. Accounts with such suffix are member accounts.", accountName))
			}
		}
		assert.Greater(t, res.ActedUpon, 20, "Number of instances acted upon seems too low")
		assert.Less(t, res.ActedUpon, 100, "Number of instances acted upon seems too high")
		assert.Greater(t, res.Skipped, 0, "Number of skipped instances seems too low")
		assert.Less(t, res.Skipped, 50, "Number of skipped instances seems too high")
		assert.Greater(t, res.SkippedAutoScaled, 20, "Number of skipped instances that belong to an Auto Scaling group seems too low")
		assert.Less(t, res.SkippedAutoScaled, 100, "Number of skipped instances that belong to an Auto Scaling group seems too high")
	})
}
