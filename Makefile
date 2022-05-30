# Copyright 2021 The pacman Authors. All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

HAS_GODOC := $(shell command -v godoc;)
HAS_GOLANGCI := $(shell command -v golangci-lint;)

define PRINT_HELP_PYSCRIPT
import re, sys

for line in sys.stdin:
	match = re.match(r'^([0-9a-zA-Z_-]+):.*?## (.*)$$', line)
	if match:
		target, help = match.groups()
		print("%-20s %s" % (target, help))
endef
export PRINT_HELP_PYSCRIPT

help:  ## prints (this) help message
	@python -c "$$PRINT_HELP_PYSCRIPT" < $(MAKEFILE_LIST)

default: help

lint:  ## lint the code
ifndef HAS_GOLANGCI
	$(error You must install github.com/golangci/golangci-lint)
endif
	@golangci-lint run -v -c .golangci.yml && echo "Lint OK"

test:  ## run tests
	@go test -timeout 30s -short -v -race -cover -coverprofile=coverage.out ./...

coverage:  ## generate coverage report
	@go tool cover -func=coverage.out

doc:  ## start godoc server
ifndef HAS_GODOC
	$(error You must install godoc, run "go get golang.org/x/tools/cmd/godoc")
endif
	@echo "Open http://localhost:6060/pkg/github.com/saucelabs/pacman/ in your browser\n"
	@godoc -http :6060

ci: lint test coverage  ## lint + test + coverage

.PHONY: lint test coverage doc ci help
