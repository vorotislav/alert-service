PROGRAM_NAME = alert-service

BUILD_VERSION=$(shell git describe --tags)
BUILD_DATE=$(shell date +%FT%T%z)
BUILD_COMMIT=$(shell git rev-parse --short HEAD)

LDFLAGS_AGENT=-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)
LDFLAGS_SERVER=-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)

.PHONY: help dep fmt test

dep: ## Get the dependencies
	go mod download

fmt: ## Format the source files
	gofumpt -l -w .

test: dep ## Run tests
	go test -timeout 5m -race -covermode=atomic -coverprofile=.coverage.out ./... && \
	go tool cover -func=.coverage.out | tail -n1 | awk '{print "Total test coverage: " $$3}'
	@rm .coverage.out

cover: dep ## Run app tests with coverage report
	go test -timeout 5m -race -covermode=atomic -coverprofile=.coverage.out ./... && \
	go tool cover -html=.coverage.out -o .coverage.html
	## Open coverage report in default system browser
	xdg-open .coverage.html
	## Remove coverage report
	sleep 2 && rm -f .coverage.out .coverage.html

lint: ## Lint the source files
	golangci-lint run --timeout 5m

build-mocks: dep
	@mockgen -destination=internal/http/handlers/mocks/mock_repo.go -package=mocks github.com/vorotislav/alert-service/internal/http/handlers Repository

build-clear:
	go build -o ./cmd/server/server ./cmd/server
	go build -o ./cmd/agent/agent ./cmd/agent

build:
	go build -ldflags "${LDFLAGS_SERVER}" -o ./cmd/server/server ./cmd/server
	go build -ldflags "${LDFLAGS_AGENT}" -o ./cmd/agent/agent ./cmd/agent

run-agent:
	go run -ldflags "${LDFLAGS_AGENT}" ./cmd/agent

run-server:
	go run -ldflags "${LDFLAGS_SERVER}" ./cmd/server
