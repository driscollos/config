# Copyright 2022 John Driscoll (https://github.com/codebyjdd)
# This code is licensed under the MIT license
# Please see LICENSE.md

.PHONY: test
test:
	@go test ./...

.PHONY: fmt
fmt:
	@gofmt -w .