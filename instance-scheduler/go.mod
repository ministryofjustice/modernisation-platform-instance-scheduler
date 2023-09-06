require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.36
	github.com/aws/aws-sdk-go-v2/credentials v1.13.35
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.113.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.53.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.21.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.37.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.21.5
	github.com/aws/smithy-go v1.14.2
	github.com/stretchr/testify v1.8.4
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.16
