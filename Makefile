.PHONY: all generate test check staticcheck vulncheck deadcode fmt build

all: generate test check fmt build

generate:
	cd ./internal/webui/ui && npm i && npm run build
	cp -r ./internal/webui/ui/dist/* ./internal/webui/static/

test:
	go test -cover --race ./...

check: staticcheck vulncheck deadcode

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

build:
	go build ./...