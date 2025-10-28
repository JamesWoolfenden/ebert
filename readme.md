# Build executable
go build -o mcp-analyzer main.go

# Run
./mcp-analyzer modelcontextprotocol

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o mcp-analyzer-linux
GOOS=windows GOARCH=amd64 go build -o mcp-analyzer.exe
GOOS=darwin GOARCH=arm64 go build -o mcp-analyzer-mac