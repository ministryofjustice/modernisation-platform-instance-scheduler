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

[Standards Link]: https://operations-engineering-reports.cloud-platform.service.justice.gov.uk/public-report/modernisation-platform-instance-scheduler "Repo standards badge."
[Standards Icon]: https://img.shields.io/endpoint?labelColor=231f20&color=005ea5&style=for-the-badge&label=MoJ%20Compliant&url=https%3A%2F%2Foperations-engineering-reports.cloud-platform.service.justice.gov.uk%2Fapi%2Fv1%2Fcompliant_public_repositories%2Fendpoint%2Fmodernisation-platform-instance-scheduler&logo=data:image/png;base64,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
[Format Code Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/format-code.yml?labelColor=231f20&style=for-the-badge&label=Formate%20Code
[Format Code Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/format-code.yml
[Scorecards Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/scorecards.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Scorecards
[Scorecards Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/scorecards.yml
[SCA Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/code-scanning.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Secure%20Code%20Analysis
[SCA Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/code-scanning.yml
[Build test push Icon]: https://img.shields.io/github/actions/workflow/status/ministryofjustice/modernisation-platform-instance-scheduler/code-scanning.yml?branch=main&labelColor=231f20&style=for-the-badge&label=Build%20Test%20Push
[Build test push Link]: https://github.com/ministryofjustice/modernisation-platform-instance-scheduler/actions/workflows/build-test-push.yml