require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.13
	github.com/aws/aws-sdk-go-v2/credentials v1.17.66
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.211.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.94.3
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.35.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.58.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.18
	github.com/aws/smithy-go v1.22.3
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/gjson v1.18.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.23
