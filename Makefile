up:
	@docker compose up -d --build

down:
	@docker compose down --remove-orphans --volumes

test: test-unit test-local

test-unit:
	@echo "=============== UNIT"
	@go test ./...

test-local:
	@echo "=============== LOCAL"
	@./test/local/run.sh

test-aws:
	@echo "Not implemented yet" 1>&2
	@exit 1

.PHONY: start, test-unit, test-local
