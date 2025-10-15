
rem go get github.com/mattn/go-sqlite3

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -o go-sample-win.exe 
rem -lws2_32 -lsqlite3