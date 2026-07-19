# use PowerShell instead of sh on Windows:
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

set quiet := true

# Run tests (short)
[default]
test:
  echo "Running tests"
  go test --short ./...

# Run all tests (including long ones)
test-full:
  echo "Running all tests"
  go test -v ./...

# Run tests with short output and coverage
[windows]
test-cover:
    echo "Running tests with coverage"
    $packages = go list ./... | Where-Object { $_ -notmatch '/db' }; go test --short -v $packages -covermode=count -coverprofile='coverage.out'
    go tool cover -func="coverage.out"

# Run all tests with coverage
test-full-cover:
  echo "Running all tests with coverage"
  go test -v ./... -covermode=count -coverprofile="coverage.out"
  go tool cover -func="coverage.out"

# Switch to main, fetch and update
main:
  git switch main
  git fetch
  git pull

# Lint code
lint:
  golangci-lint run

# Go vet
vet:
  go vet ./...