require (
	github.com/aws/aws-lambda-go v1.54.0
	github.com/aws/aws-sdk-go-v2 v1.43.8
	github.com/aws/aws-sdk-go-v2/config v1.32.39
	github.com/aws/aws-sdk-go-v2/credentials v1.19.38
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.323.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.124.5
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.44.8
	github.com/aws/aws-sdk-go-v2/service/ssm v1.73.8
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.8
	github.com/aws/smithy-go v1.27.10
	github.com/stretchr/testify v1.12.1
	github.com/tidwall/gjson v1.19.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.40 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.39 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.8 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.24
