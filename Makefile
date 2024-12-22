# https://go.dev/blog/govulncheck
# install it by `go install golang.org/x/vuln/cmd/govulncheck@latest`
vuln:
	which govulncheck
	govulncheck ./...

deps:
	go mod download
	go mod verify
	go mod tidy

start:
	go run main.go ./contrib/config.yaml
