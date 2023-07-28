require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.19.1
	github.com/aws/aws-sdk-go-v2/config v1.18.30
	github.com/aws/aws-sdk-go-v2/credentials v1.13.29
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.103.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.19.12
	github.com/aws/aws-sdk-go-v2/service/ssm v1.36.7
	github.com/aws/aws-sdk-go-v2/service/sts v1.20.1
	github.com/aws/smithy-go v1.13.5
	github.com/stretchr/testify v1.8.4
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.16
