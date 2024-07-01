require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.30.1
	github.com/aws/aws-sdk-go-v2/config v1.27.22
	github.com/aws/aws-sdk-go-v2/credentials v1.17.22
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.167.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.81.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.32.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.0
	github.com/aws/smithy-go v1.20.3
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.21
