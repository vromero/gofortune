SOURCE_FILES?=./...
TEST_PATTERN?=.
TEST_OPTIONS?=

export PATH := ./bin:$(PATH)
export GO111MODULE := on

# Install all the build and lint dependencies
setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go mod download
.PHONY: setup

# Run all the tests
test:
	go test $(TEST_OPTIONS) -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

# gofmt and goimports all go files
fmt:
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done
	# find . -name '*.md' -not -wholename './vendor/*' | xargs prettier --write
.PHONY: fmt

# Run all the linters
lint:
	# TODO: fix tests and lll issues
	./bin/golangci-lint run --tests=false --enable-all --disable=lll ./...
	# find . -name '*.md' -not -wholename './vendor/*' | xargs prettier -l
.PHONY: lint

# Run all the tests and code checks
ci: build test # lint
.PHONY: ci

# Build a beta version of goreleaser
build:
	go build
.PHONY: build

# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=node_modules \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow' .
.PHONY: todo


.DEFAULT_GOAL := build
