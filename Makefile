.PHONY: server server-dev client client-dev start-pg stop-pg test test-cover test-cover-svg test-cover-html install-deps gen-proto gen-proto-auth-v1
.SILENT:

lint:
	@$(CURDIR)/bin/golangci-lint run -c .golangci.yaml --path-prefix . --fix

test:
	@go test --cover --coverprofile=coverage.out ./...

test-cover:
	@go test --coverprofile=coverage.out ./... > /dev/null
	@go tool cover -func=coverage.out | grep total | grep -oE '[0-9]+(\.[0-9]+)?%'

test-cover-svg:
	@go test --coverprofile=coverage.out ./... > /dev/null
	@$(CURDIR)/bin/go-cover-treemap -coverprofile coverage.out > coverage.svg
	@xdg-open ./coverage.svg

test-cover-html:
	@go test --coverprofile=coverage.out ./... > /dev/null
	@go tool cover -html="coverage.out"

doc:
	@$(CURDIR)/bin/godoc

install-deps:
	@GOBIN=$(CURDIR)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@GOBIN=$(CURDIR)/bin go install github.com/nikolaydubina/go-cover-treemap@latest
	@GOBIN=$(CURDIR)/bin go install -v golang.org/x/tools/cmd/godoc@latest
	@go mod tidy