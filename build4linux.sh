## go env -w CGO_ENABLED=1
## go env -w GOOS=linux  
## go env -w GOARCH=amd64
## set CGO_CFLAGS=-Wno-error=implicit-function-declaration -Wno-error=unused-variable
## go build -o ./bin/go-sample-linux
go env -w CGO_ENABLED=1
go env -w GOOS=linux  
go env -w GOARCH=amd64
go build -o ./go-sample-linux
