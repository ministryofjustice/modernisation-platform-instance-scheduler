cd instance-scheduler
go mod tidy
go mod download
cd ..
make
aws-vault exec core-shared-services-production -- sam validate
aws-vault exec core-shared-services-production -- sam local invoke
