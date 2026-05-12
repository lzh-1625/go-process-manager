SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
@REM go build -ldflags="-s -w" -tags="slim" -o gpm -trimpath .\cmd\go_process_manager\.
go build -ldflags="-s -w" -o gpm -trimpath .\cmd\go_process_manager\.
