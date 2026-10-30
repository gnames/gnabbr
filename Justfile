# Install just: https://github.com/casey/just

# Default recipe - run tests
default: test

# Run all tests (skip integration tests requiring database)
test:
    go test -count=1 ./...

# Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run all tests with race detector
test-race:
    go test -count=1 -race ./...

# Clean build artifacts
clean:
    rm -rf coverage.out coverage.html
    go clean
    @echo "✅ Cleaned build artifacts"

# Format all Go code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Tidy dependencies
tidy:
    go mod tidy

# Verify project builds and all tests pass
verify: fmt tidy test
    @echo "✅ Verification complete: code formatted, dependencies tidied, tests passing"
