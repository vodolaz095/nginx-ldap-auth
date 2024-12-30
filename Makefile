export app=nginx-ldap-auth
export majorVersion=0
export minorVersion=1

export arch=$(shell uname)-$(shell uname -m)
export gittip=$(shell git log --format='%h' -n 1)
export subver=$(shell hostname)_on_$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
export patchVersion=$(shell git log --format='%h' | wc -l)
export ver=$(majorVersion).$(minorVersion).$(patchVersion).$(gittip)-$(arch)

include make/*.mk

tools:
	which go
	which govulncheck

# https://go.dev/blog/govulncheck
# install it by `go install golang.org/x/vuln/cmd/govulncheck@latest`
vuln:
	which govulncheck
	govulncheck ./...

deps:
	go mod download
	go mod verify
	go mod tidy

build:
# https://www.reddit.com/r/golang/comments/10te58n/error_loading_shared_library_libresolvso2_no_such/
	CGO_ENABLED=0 go build -ldflags "-X main.Subversion=$(subver) -X main.Version=$(ver)" -o build/$(app) main.go
#	upx build/eda


start:
	go run main.go ./contrib/config.yaml

.PHONY: build
