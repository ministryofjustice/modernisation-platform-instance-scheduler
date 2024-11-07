require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.32.4
	github.com/aws/aws-sdk-go-v2/config v1.28.1
	github.com/aws/aws-sdk-go-v2/credentials v1.17.42
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.187.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.89.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.55.3
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.3
	github.com/aws/smithy-go v1.22.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.23
