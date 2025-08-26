# Copyright 2022 John Driscoll (https://github.com/codebyjdd)
# This code is licensed under the MIT license
# Please see LICENSE.md

.PHONY: test
test:
	@go test ./...

.PHONY: mocks
mocks:
	@go get go.uber.org/mock/mockgen
	@go get golang.org/x/tools/internal/gocommand
	@go get golang.org/x/tools/internal/imports
	@cd internal; go generate ./...
	@go mod tidy

.PHONY: fmt
fmt:
	@gofmt -w .