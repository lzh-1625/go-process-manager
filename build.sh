export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
# go build -ldflags="-s -w" -tags="slim" -o go_process_manager -trimpath cmd/go_process_manager/*.go
go build -ldflags="-s -w" -o go_process_manager -trimpath cmd/go_process_manager/*.go
