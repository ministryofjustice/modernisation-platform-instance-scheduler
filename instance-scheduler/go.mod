require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.23.5
	github.com/aws/aws-sdk-go-v2/config v1.25.12
	github.com/aws/aws-sdk-go-v2/credentials v1.16.10
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.140.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.64.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.25.2
	github.com/aws/aws-sdk-go-v2/service/ssm v1.44.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.3
	github.com/aws/smithy-go v1.18.1
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.21
