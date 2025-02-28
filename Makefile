.PHONY: all generate test check staticcheck vulncheck deadcode fmt build

all: generate test check fmt build

generate:
	cd ./internal/webui/ui && npm i && npm run build
	cp -r ./internal/webui/ui/dist/* ./internal/webui/static/

test:
	go test -cover --race ./...

check: staticcheck vulncheck deadcode

staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

fmt:
	go tool mvdan.cc/gofumpt -w ./

build:
	go build ./...