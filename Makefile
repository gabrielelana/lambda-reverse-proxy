up:
	@docker compose up -d --build

down:
	@docker compose down --remove-orphans --volumes

test-all: test-unit, test-local

test-unit:
	go test ./...

test-local:
	@./test/local/run.sh

test-aws:
	@echo "Not implemented yet" 1>&2
	@exit 1

.PHONY: start, test-unit, test-local
