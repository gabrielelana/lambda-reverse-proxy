.PHONY: help start test-unit test-local

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

up:								## Starts local development environment
	@docker compose up -d --build

down:							## Stops local development environment, see `up`
	@docker compose down --remove-orphans --volumes

test: test-unit test-local		## Runs all tests

test-unit:						## Runs unit tests
	@echo "=============== UNIT"
	@go test ./...

test-local:						## Runs e2e tests targeting a local environment
	@echo "=============== LOCAL"
	@./test/local/run.sh

test-aws:						## Runs e2e tests targeting AWS environment
	@echo "=============== AWS"
	@echo "Not implemented yet" 1>&2
	@exit 1
