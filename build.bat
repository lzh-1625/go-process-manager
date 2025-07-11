SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
@REM go build -ldflags="-s -w" -tags="slim" -o go_process_manager -trimpath cmd/go_process_manager/main.go
go build -ldflags="-s -w" -o go_process_manager -trimpath cmd/go_process_manager/main.go