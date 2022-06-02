help:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint: ## run code linters
	golangci-lint run

test: ## run tests
	go test -race ./...

test-coverage: ## run unit tests with test coverage
	go test ./... -coverprofile .testCoverage -covermode=atomic -coverpkg=./...

install-linters: ## install all linters
	go install github.com/fraugster/flint@v0.1.1
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.46.2
