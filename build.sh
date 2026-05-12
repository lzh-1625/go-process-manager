export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
# go build -ldflags="-s -w" -tags="slim" -o gpm -trimpath cmd/go_process_manager/*.go
go build -ldflags="-s -w" -o gpm -trimpath cmd/go_process_manager/*.go
