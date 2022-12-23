require (
	github.com/aws/aws-lambda-go v1.36.0
	github.com/aws/aws-sdk-go-v2 v1.17.3
	github.com/aws/aws-sdk-go-v2/config v1.18.4
	github.com/aws/aws-sdk-go-v2/credentials v1.13.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.76.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.16.11
	github.com/aws/aws-sdk-go-v2/service/ssm v1.33.4
	github.com/aws/aws-sdk-go-v2/service/sts v1.17.7
	github.com/aws/smithy-go v1.13.5
	github.com/stretchr/testify v1.8.1
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.16
