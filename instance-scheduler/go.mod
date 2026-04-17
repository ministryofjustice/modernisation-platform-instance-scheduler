require (
	github.com/aws/aws-lambda-go v1.54.0
	github.com/aws/aws-sdk-go-v2 v1.41.5
	github.com/aws/aws-sdk-go-v2/config v1.32.15
	github.com/aws/aws-sdk-go-v2/credentials v1.19.14
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.297.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.118.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.5
	github.com/aws/aws-sdk-go-v2/service/ssm v1.68.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.10
	github.com/aws/smithy-go v1.25.0
	github.com/stretchr/testify v1.11.1
	github.com/tidwall/gjson v1.18.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.19 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.24
