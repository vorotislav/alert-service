PROGRAM_NAME = alert-service

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

build-mocks: dep
	@mockgen -destination=internal/http/handlers/ping/mocks/mock_repo.go -package=mocks github.com/vorotislav/alert-service/internal/http/handlers/ping Repository
	@mockgen -destination=internal/http/handlers/update/mocks/mock_repo.go -package=mocks github.com/vorotislav/alert-service/internal/http/handlers/update Repository
	@mockgen -destination=internal/http/handlers/updates/mocks/mock_repo.go -package=mocks github.com/vorotislav/alert-service/internal/http/handlers/updates Repository
	@mockgen -destination=internal/http/handlers/value/mocks/mock_repo.go -package=mocks github.com/vorotislav/alert-service/internal/http/handlers/value Repository
