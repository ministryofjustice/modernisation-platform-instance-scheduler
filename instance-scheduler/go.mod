require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.30.5
	github.com/aws/aws-sdk-go-v2/config v1.27.32
	github.com/aws/aws-sdk-go-v2/credentials v1.17.31
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.177.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.82.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.32.6
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.6
	github.com/aws/smithy-go v1.20.4
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.6 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.21
