SHELL := /bin/bash
.SILENT:
.DEFAULT_GOAL := help

# Global vars
export SYS_GO=$(shell which go)
export SYS_GOFMT=$(shell which gofmt)
export SYS_GOLANGCI_LINT=$(shell which golangci-lint)

.PHONY: run
## Run as go run cmd/main.go
run: cmd/main.go
	$(SYS_GO) run cmd/main.go

.PHONY: tidy
## Install all requirements
tidy: go.mod
	$(SYS_GO) mod tidy

.PHONY: fmt
## Run go fmt
fmt:
	$(SYS_GOFMT) -s -w .

.PHONY: vet
## Run go vet ./...
vet:
	$(SYS_GO) vet ./...

.PHONY: test
## Run tests without caching
test:
	@$(SYS_GO) clean -testcache && $(SYS_GO) test ./...

.PHONY: lint
## Run golangci-lint
lint:
	@$(SYS_GOLANGCI_LINT) -v run --out-format=colored-line-number

.PHONY: help
## Show this help message
help:
	@echo "$$(tput bold)Available rules:$$(tput sgr0)"
	@echo
	@sed -n -e "/^## / { \
		h; \
		s/.*//; \
		:doc" \
		-e "H; \
		n; \
		s/^## //; \
		t doc" \
		-e "s/:.*//; \
		G; \
		s/\\n## /---/; \
		s/\\n/ /g; \
		p; \
	}" ${MAKEFILE_LIST} \
	| LC_ALL='C' sort --ignore-case \
	| awk -F '---' \
		-v ncol=$$(tput cols) \
		-v indent=19 \
		-v col_on="$$(tput setaf 6)" \
		-v col_off="$$(tput sgr0)" \
	'{ \
		printf "%s%*s%s ", col_on, -indent, $$1, col_off; \
		n = split($$2, words, " "); \
		line_length = ncol - indent; \
		for (i = 1; i <= n; i++) { \
			line_length -= length(words[i]) + 1; \
			if (line_length <= 0) { \
				line_length = ncol - indent - length(words[i]) - 1; \
				printf "\n%*s ", -indent, " "; \
			} \
			printf "%s ", words[i]; \
		} \
		printf "\n"; \
	}' \
	| more $(shell test $(shell uname) == Darwin && echo '--no-init --raw-control-chars')

