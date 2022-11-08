require (
	github.com/aws/aws-lambda-go v1.34.1
	github.com/aws/aws-sdk-go-v2 v1.17.1
	github.com/aws/aws-sdk-go-v2/config v1.17.10
	github.com/aws/aws-sdk-go-v2/credentials v1.12.23
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.65.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.16.4
	github.com/aws/aws-sdk-go-v2/service/ssm v1.32.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.17.1
	github.com/aws/smithy-go v1.13.4
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.16
