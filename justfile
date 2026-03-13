# x-mcp justfile

# Build the MCP server
build:
    go build -o bin/x-mcp .

# Build the CLI
build-cli:
    go build -o bin/x-cli ./cmd/cli/

# Install the MCP server to $GOPATH/bin
install:
    go install .

# Install the CLI to $GOPATH/bin
install-cli:
    go install ./cmd/cli/

# Run the MCP server
run:
    go run .

# Run tests
test:
    go test ./...
