REM Linux 去执行
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
SET "GOFLAGS=-buildvcs=false"
go build