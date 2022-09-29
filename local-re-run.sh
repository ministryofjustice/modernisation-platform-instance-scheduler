cd instance-scheduler
go mod tidy
go mod download
cd ..
make
aws-vault exec mod -- sam validate
aws-vault exec mod -- sam local invoke
