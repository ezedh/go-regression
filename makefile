# Get coverage
tests:
	@go test ./... -race -coverprofile=/tmp/coverage.out -covermode=atomic
	@go tool cover -func=/tmp/coverage.out