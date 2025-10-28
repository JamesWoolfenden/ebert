# Basic usage
go run main.go modelcontextprotocol

# With JSON export
go run main.go modelcontextprotocol --json

# With GitHub token for higher rate limits (60/hour → 5000/hour)
export GITHUB_TOKEN=your_token_here
go run main.go username