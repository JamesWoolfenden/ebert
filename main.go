package main

import (
	"ebert/src"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <github-username>")
		fmt.Println("Example: go run main.go modelcontextprotocol")
		fmt.Println("\nOptional: Set GITHUB_TOKEN environment variable for higher rate limits")
		os.Exit(1)
	}

	username := os.Args[1]
	token := os.Getenv("GITHUB_TOKEN")

	fmt.Printf("Analyzing GitHub user: %s\n", username)
	fmt.Println("Fetching data from GitHub API...")

	analyzer := ebert.NewAnalyzer(token)
	analysis, err := analyzer.Analyze(username)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ebert.PrintAnalysis(analysis)

	// Optionally save to JSON
	if len(os.Args) > 2 && os.Args[2] == "--json" {
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("%s_analysis.json", username)
		if err := os.WriteFile(filename, jsonData, 0644); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n💾 Analysis saved to: %s\n", filename)
	}
}
