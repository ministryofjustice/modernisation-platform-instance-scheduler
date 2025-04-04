package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Run("Test request", func(t *testing.T) {

		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:                        LoadDefaultConfig,
			CreateSSMClient:                          CreateSSMClient,
			GetParameter:                             getParameter,
			CreateSecretManagerClient:                CreateSecretManagerClient,
			GetSecret:                                getSecret,
			GetEc2ClientForMemberAccount:             getEc2ClientForMemberAccount,
			GetRDSClientForMemberAccount:             getRDSClientForMemberAccount,
			StopStartTestInstancesInMemberAccount:    stopStartTestInstancesInMemberAccount,
			StopStartTestRDSInstancesInMemberAccount: StopStartTestRDSInstancesInMemberAccount,
		}
		result, err := instanceScheduler.handler(InstanceSchedulingRequest{Action: "Test"})
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

		addMsg := "Please manually check the Instance Scheduler logs and verify this is reasonable. If it is reasonable, modify and adjust this test accordingly."
		assert.Greater(t, res.ActedUpon, 0, fmt.Sprintf("Number of instances acted upon seems too low. %v", addMsg))
		assert.Less(t, res.ActedUpon, 200, fmt.Sprintf("Number of instances acted upon seems too high. %v", addMsg))
		assert.Greater(t, res.Skipped, 0, fmt.Sprintf("Number of skipped instances seems too low. %v", addMsg))
		assert.Less(t, res.Skipped, 200, fmt.Sprintf("Number of skipped instances seems too high. %v", addMsg))
		assert.Greater(t, res.SkippedAutoScaled, 0, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too low. %v", addMsg))
		assert.Less(t, res.SkippedAutoScaled, 200, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too high. %v", addMsg))
	})
}
