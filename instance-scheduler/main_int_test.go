package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Run("Test request", func(t *testing.T) {
		// Docker image test preparation
		t.Log("Running `sam build` to prepare the Dockerfile for Lambda...")
		cmd := exec.Command("sam", "build", "--use-container")
		err := cmd.Run()
		assert.NoError(t, err, "SAM build failed")

		t.Log("Building Docker image from SAM-generated Dockerfile...")
		imageName := "instance-scheduler-test"
		cmd = exec.Command("docker", "build", "-t", imageName, "-f", ".aws-sam/build/instance-scheduler/Dockerfile", ".")
		err = cmd.Run()
		assert.NoError(t, err, "Docker image build failed")

		// Verify the Docker image was created
		t.Log("Verifying Docker image was created...")
		cmd = exec.Command("docker", "images", "-q", imageName)
		imageID, err := cmd.Output()
		assert.NoError(t, err, "Failed to verify Docker image")
		assert.NotEmpty(t, string(imageID), "Docker image should exist")

		// Run the Docker container and check output
		t.Log("Running Docker container and checking output...")
		cmd = exec.Command("docker", "run", "--rm", imageName, "echo", "Hello from Docker")
		output, err := cmd.Output()
		assert.NoError(t, err, "Failed to run Docker container")
		assert.Contains(t, string(output), "Hello from Docker", "Unexpected output from Docker container")

		// Environment variable check
		t.Log("Checking environment variable settings in Docker container...")
		cmd = exec.Command("docker", "run", "--rm", imageName, "printenv", "INSTANCE_SCHEDULING_SKIP_ACCOUNTS")
		envOutput, err := cmd.Output()
		assert.NoError(t, err, "Failed to retrieve environment variable from Docker container")
		expectedEnvValue := "mi-platform-development,analytical-platform-data-development,analytical-platform-development,moj-network-operations-centre-preproduction,opg-lpa-data-store-development,"
		assert.Equal(t, expectedEnvValue, string(envOutput), "Environment variable INSTANCE_SCHEDULING_SKIP_ACCOUNTS does not match expected value")

		// Set the environment variable
		os.Setenv("INSTANCE_SCHEDULING_SKIP_ACCOUNTS", "mi-platform-development,analytical-platform-data-development,analytical-platform-development,moj-network-operations-centre-preproduction,opg-lpa-data-store-development,")

		// Run existing InstanceScheduler test cases
		instanceScheduler := InstanceScheduler{
			LoadDefaultConfig:                        LoadDefaultConfig,
			GetEnv:                                   os.Getenv,
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

		// Validate InstanceSchedulingResponse
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
		assert.Greater(t, res.ActedUpon, 0, fmt.Sprintf("Number of instances acted upon seems too low. %v", addMsg))
		assert.Less(t, res.ActedUpon, 200, fmt.Sprintf("Number of instances acted upon seems too high. %v", addMsg))
		assert.Greater(t, res.Skipped, 0, fmt.Sprintf("Number of skipped instances seems too low. %v", addMsg))
		assert.Less(t, res.Skipped, 200, fmt.Sprintf("Number of skipped instances seems too high. %v", addMsg))
		assert.Greater(t, res.SkippedAutoScaled, 0, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too low. %v", addMsg))
		assert.Less(t, res.SkippedAutoScaled, 200, fmt.Sprintf("Number of skipped instances that belong to an Auto Scaling group seems too high. %v", addMsg))
	})
}
