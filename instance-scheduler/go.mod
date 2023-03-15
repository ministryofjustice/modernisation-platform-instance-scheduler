require (
	github.com/aws/aws-lambda-go v1.38.0
	github.com/aws/aws-sdk-go-v2 v1.17.6
	github.com/aws/aws-sdk-go-v2/config v1.18.17
	github.com/aws/aws-sdk-go-v2/credentials v1.13.17
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.89.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.18.7
	github.com/aws/aws-sdk-go-v2/service/ssm v1.35.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.6
	github.com/aws/smithy-go v1.13.5
	github.com/stretchr/testify v1.8.2
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module instance-scheduler

go 1.16
