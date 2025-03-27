# Ministry of Justice Instance Scheduler

[![Standards Icon]][Standards Link] [![Format Code Icon]][Format Code Link] [![Scorecards Icon]][Scorecards Link] [![SCA Icon]][SCA Link] [![Build test push Icon]][Build test push Link]

A Go lambda function for stopping and starting instance, rds resources and autoscaling groups. The function is used by the [Ministry of Justice Modernisation Platform](https://github.com/ministryofjustice/modernisation-platform) and can be re-used in any environment with minimal changes.

The list of MP accounts to be included by the scheduler, whether production or for testing purposes, is determined by the following criteria based on information obtained from the github/ministryofjustice/modernisation-platform/environments json files.

- the account-type is "member"
- the account is not production
- the field "instance_scheduler_skip": ["true"] has NOT been added to the "environment" list in the json.

If the account does not meet any of the above criteria then it is excluded.

## Requirements

Testing changes to the go source code of the module can be done by creating a Pull Request with a new branch containing the changes. The log output of the github workflow build-test-push.yml will show the results. 

For development & running the scheduler locally:

- [AWS Vault](https://github.com/99designs/aws-vault) with a profile configured to access the core-shared-services account
- [Docker installed](https://www.docker.com/community-edition)
- [Golang](https://golang.org)
- SAM CLI - [Install the SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)

To install the SAM CLI on macOS:

```
brew tap aws/tap
brew install aws-sam-cli
```

## Local development

Install dependencies & build the target

    cd modernisation-platform-instance-scheduler
    make

> :warning: Code changes require a rebuild before you're able to run them. Use `local-re-run.sh` as a utility script when re-running local code changes.

Validate SAM template

    aws-vault exec core-shared-services-production -- sam validate

Invoke Function

    aws-vault exec core-shared-services-production -- sam local invoke --event event.json

Test Function in the Cloud

    aws-vault exec core-shared-services-production -- sam sync --stack-name instance-scheduler --watch

Deploy on sprinkler-development using the local `samconfig.toml` and preventing prompts and failure when the stack is unchanged. The following command requires the `instance-scheduler-ecr-repo` ECR repository to be present in order to succeed.

    ACCOUNT_ID=$(aws-vault exec sprinkler-development -- aws sts get-caller-identity --query Account --output text)
    aws-vault exec sprinkler-development -- sam deploy --no-confirm-changeset --no-fail-on-empty-changeset --region eu-west-2 --image-repository $ACCOUNT_ID.dkr.ecr.eu-west-2.amazonaws.com/instance-scheduler-ecr-repo

Module initialisation. The following commands were used in order to generate the required `go.mod` and `go.sum` files prior to the first run of the tests.

    cd instance-scheduler
    go mod init github.com/ministryofjustice/modernisation-platform-instance-scheduler
    go mod tidy
    go mod download

Run Tests

    cd instance-scheduler
    aws-vault exec core-shared-services-production -- go test -v .

## Configuration

Note that setting a local environment variable **INSTANCE_SCHEDULING_SKIP_ACCOUNTS** is no longer required and it is not used.

## References

1. [User Guide](https://user-guide.modernisation-platform.service.justice.gov.uk/concepts/environments/instance-scheduling.html)
2. [How the original Go SAM project was created](sam-init.md)
3. [AWS Serverless Application Model](https://aws.amazon.com/serverless/sam/)
4. [AWS Serverless Application Repository](https://aws.amazon.com/serverless/serverlessrepo/)
5. [Terraform module for deployment of the lambda](https://github.com/ministryofjustice/modernisation-platform-terraform-lambda-function)

[Standards Link]: https://github-community.cloud-platform.service.justice.gov.uk/repository-standards/modernisation-platform-instance-scheduler "Repo standards badge."
[Standards Icon]: https://github-community.cloud-platform.service.justice.gov.uk/repository-standards/api/modernisation-platform-instance-scheduler/badge
[Format Code Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/format-code.yml?labelColor=231f20&style=for-the-badge&label=Formate%20Code
[Format Code Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/format-code.yml
[Scorecards Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/scorecards.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Scorecards
[Scorecards Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/scorecards.yml
[SCA Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/code-scanning.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Secure%20Code%20Analysis
[SCA Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/code-scanning.yml
[Build test push Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/code-scanning.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Build%20Test%20Push
[Build test push Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/build-test-push.yml
