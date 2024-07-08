GOMARKDOC := gomarkdoc

# hacked gomarkdoc to resolve https://github.com/princjef/gomarkdoc/issues/113
# GOMARKDOC := /Users/ryanclark/rclark/gomarkdoc/cmd/gomarkdoc/gomarkdoc_hack

.PHONY: init doc test-standard test-structured test

init:
	go mod tidy
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@v1.1.0
	go install gotest.tools/gotestsum@v1.12.0
	go install github.com/axw/gocov/gocov@v1.1.0
	go install github.com/matm/gocov-html/cmd/gocov-html@v1.4.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	@if [ ! -e .git/hooks/pre-commit ]; then \
		chmod 755 .githooks/pre-commit; \
		ln -s $(PWD)/.githooks/pre-commit .git/hooks/pre-commit; \
	fi

doc:
	@$(GOMARKDOC) \
		--output standard.md  \
		.

	@$(GOMARKDOC) \
	  --tags structuredlogs \
		--output structured.md  \
		--template-file package=templates/package.md \
		.

test-standard:
	@gotestsum --format testname -- -coverprofile=coverage.out ./...
	@gocov convert coverage.out | gocov-html > coverage-standard.html

test-structured:
	@gotestsum --format testname -- -tags=structuredlogs -coverprofile=coverage.out ./...
	@gocov convert coverage.out | gocov-html > coverage-structured.html

test: test-standard test-structured
