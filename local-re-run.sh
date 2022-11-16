cd instance-scheduler
go mod tidy
go mod download
cd ..
make
aws-vault exec core-shared-services-production -- sam validate
aws-vault exec core-shared-services-production -- sam local invoke --event event.json
cd instance-scheduler
aws-vault exec core-shared-services-production -- go test -v
