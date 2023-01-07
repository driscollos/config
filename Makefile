.PHONY: test
test:
	@go test ./...

.PHONY: fmt
fmt:
	@gofmt -w .