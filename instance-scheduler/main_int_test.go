package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Run("Test request", func(t *testing.T) {
		// Accounts mi-platform-development and analytical-platform-data-development cause the main_int_test.go to fail because they are non-member accounts
		// lacking the InstanceSchedulerAccess role, but they have the '-development' suffix typically present in member accounts.
		os.Setenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS", "mi-platform-development,analytical-platform-data-development,analytical-platform-development,moj-network-operations-centre-preproduction,opg-lpa-data-store-development,")

		result, err := handler(InstanceSchedulingRequest{Action: "Test"})
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
		assert.NotEmpty(t, res.NonMemberAccountNames, "No non-member account was found")
		for _, accountName := range res.NonMemberAccountNames {
			assert.False(t, strings.HasSuffix(accountName, "-production"), fmt.Sprintf("Production account %v was found in the list of non-member accounts. Production accounts should be skipped.", accountName))
			if !strings.HasPrefix(accountName, "core-vpc-") {
				assert.False(t, strings.HasSuffix(accountName, "-development"), fmt.Sprintf("Non-member account %v was found with suffix '-development'. Accounts with such suffix are member accounts.", accountName))
				assert.False(t, strings.HasSuffix(accountName, "-test"), fmt.Sprintf("Non-member account %v was found with suffix '-test'. Accounts with such suffix are member accounts.", accountName))
				assert.False(t, strings.HasSuffix(accountName, "-preproduction"), fmt.Sprintf("Non-member account %v was found with suffix '-preproduction'. Accounts with such suffix are member accounts.", accountName))
			}
		}
		addMsg := "Please manually check the Instance Scheduler logs and verify this is reasonable. If it is reasonable, modify and adjust this test accordingly."
		assert.Greater(t, res.ActedUpon, 20, fmt.Sprintf("Number of instances acted upon seems too low. %v", addMsg))
		assert.Less(t, res.ActedUpon, 100, fmt.Sprintf("Number of instances acted upon seems too high. %v", addMsg))
		assert.Greater(t, res.Skipped, 0, fmt.Sprintf("Number of skipped instances seems too low. %v", addMsg))
		assert.Less(t, res.Skipped, 80, fmt.Sprintf("Number of skipped instances seems too high. %v", addMsg))
		assert.Greater(t, res.SkippedAutoScaled, 20, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too low. %v", addMsg))
		assert.Less(t, res.SkippedAutoScaled, 100, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too high. %v", addMsg))
	})
}
