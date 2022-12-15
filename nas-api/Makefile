GO_VER              ?= go

lint: golangci-lint

.PHONY: ensure-mod lint install-tools test run-local golangci-lint tools

install-tools:
	@echo ">  Installing tools..."
	(cd /tmp && go get -u github.com/golangci/golangci-lint/cmd/golangci-lint)

golangci-lint:
	@echo "==> Checking source code with golangci-lint..."
	@golangci-lint run --out-format code-climate | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'

test:
	@echo ">  Running tests..."
	go test -cover -race -v ./...

run-local: ensure-mod
	cd cmd/nas && go run main.go

ensure-mod:
	go mod download
