# Ministry of Justice Instance Scheduler

[![repo standards badge](https://img.shields.io/badge/dynamic/json?color=blue&style=for-the-badge&logo=github&label=MoJ%20Compliant&query=%24.result&url=https%3A%2F%2Foperations-engineering-reports.cloud-platform.service.justice.gov.uk%2Fapi%2Fv1%2Fcompliant_public_repositories%2Fmodernisation-platform-instance-scheduler)](https://operations-engineering-reports.cloud-platform.service.justice.gov.uk/public-github-repositories.html#modernisation-platform-instance-scheduler "Link to report")

A Go lambda function for stopping and starting instance, rds resources and autoscaling groups. The function is used by the [Ministry of Justice Modernisation Platform](https://github.com/ministryofjustice/modernisation-platform) and can be re-used in any environment with minimal changes.

## Requirements

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

- Environment variable **INSTANCE_SCHEDULING_SKIP_ACCOUNTS**: A comma-separated list of account names to be skipped from instance scheduling. For example:
  `export INSTANCE_SCHEDULING_SKIP_ACCOUNTS="xhibit-portal-development,another-development,"`. As can be observed in this example, every account name must be suffixed with a leading comma, hence the last comma in the list.

## References

1. [User Guide](https://user-guide.modernisation-platform.service.justice.gov.uk/concepts/environments/instance-scheduling.html)
2. [How the original Go SAM project was created](sam-init.md)
3. [AWS Serverless Application Model](https://aws.amazon.com/serverless/sam/)
4. [AWS Serverless Application Repository](https://aws.amazon.com/serverless/serverlessrepo/)
5. [Terraform module for deployment of the lambda](https://github.com/ministryofjustice/modernisation-platform-terraform-lambda-function)
