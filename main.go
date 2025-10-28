package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GitHub API structures
type GitHubUser struct {
	Login           string    `json:"login"`
	Name            string    `json:"name"`
	Company         string    `json:"company"`
	Blog            string    `json:"blog"`
	Email           string    `json:"email"`
	Bio             string    `json:"bio"`
	PublicRepos     int       `json:"public_repos"`
	Followers       int       `json:"followers"`
	Following       int       `json:"following"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	AvatarURL       string    `json:"avatar_url"`
	HTMLURL         string    `json:"html_url"`
	TwitterUsername string    `json:"twitter_username"`
}

type GitHubRepo struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Language        string    `json:"language"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	Archived        bool      `json:"archived"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	Topics          []string  `json:"topics"`
	HasPages        bool      `json:"has_pages"`
	HTMLURL         string    `json:"html_url"`
}

type GitHubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Payload   struct {
		Commits []interface{} `json:"commits"`
	} `json:"payload"`
}

// Analysis structures
type RiskScores struct {
	Identity    float64 `json:"identity"`
	Activity    float64 `json:"activity"`
	Quality     float64 `json:"quality"`
	Maintenance float64 `json:"maintenance"`
	Community   float64 `json:"community"`
}

type Metrics struct {
	AccountAgeDays  int `json:"account_age_days"`
	Repos           int `json:"repos"`
	Stars           int `json:"stars"`
	Forks           int `json:"forks"`
	Followers       int `json:"followers"`
	RecentCommits   int `json:"recent_commits"`
	RecentlyUpdated int `json:"recently_updated"`
	Archived        int `json:"archived"`
	NPMPackages     int `json:"npm_packages"`
	PythonPackages  int `json:"python_packages"`
}

type Analysis struct {
	User         GitHubUser `json:"user"`
	Scores       RiskScores `json:"scores"`
	OverallScore float64    `json:"overall_score"`
	RiskLevel    string     `json:"risk_level"`
	Metrics      Metrics    `json:"metrics"`
	RedFlags     []string   `json:"red_flags"`
	Warnings     []string   `json:"warnings"`
	Positives    []string   `json:"positives"`
	Timestamp    time.Time  `json:"timestamp"`
}

// GitHubClient handles API requests
type GitHubClient struct {
	BaseURL string
	Token   string // Optional: GitHub token for higher rate limits
}

func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		BaseURL: "https://api.github.com",
		Token:   token,
	}
}

func (c *GitHubClient) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *GitHubClient) GetUser(username string) (*GitHubUser, error) {
	data, err := c.get(fmt.Sprintf("%s/users/%s", c.BaseURL, username))
	if err != nil {
		return nil, err
	}

	var user GitHubUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *GitHubClient) GetRepos(username string) ([]GitHubRepo, error) {
	data, err := c.get(fmt.Sprintf("%s/users/%s/repos?per_page=100&sort=updated", c.BaseURL, username))
	if err != nil {
		return nil, err
	}

	var repos []GitHubRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (c *GitHubClient) GetEvents(username string) ([]GitHubEvent, error) {
	data, err := c.get(fmt.Sprintf("%s/users/%s/events/public?per_page=100", c.BaseURL, username))
	if err != nil {
		return nil, err
	}

	var events []GitHubEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// Analyzer performs the security analysis
type Analyzer struct {
	client *GitHubClient
}

func NewAnalyzer(token string) *Analyzer {
	return &Analyzer{
		client: NewGitHubClient(token),
	}
}

func (a *Analyzer) Analyze(username string) (*Analysis, error) {
	// Fetch data from GitHub
	user, err := a.client.GetUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	repos, err := a.client.GetRepos(username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}

	events, err := a.client.GetEvents(username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	return a.performAnalysis(user, repos, events), nil
}

func (a *Analyzer) performAnalysis(user *GitHubUser, repos []GitHubRepo, events []GitHubEvent) *Analysis {
	now := time.Now()
	accountAge := int(now.Sub(user.CreatedAt).Hours() / 24)

	// Calculate metrics
	metrics := a.calculateMetrics(user, repos, events, accountAge, now)

	// Calculate risk scores
	scores := RiskScores{
		Identity:    a.calculateIdentityScore(user, accountAge),
		Activity:    a.calculateActivityScore(metrics, accountAge, len(repos)),
		Quality:     a.calculateQualityScore(repos, metrics),
		Maintenance: a.calculateMaintenanceScore(metrics, len(repos)),
		Community:   a.calculateCommunityScore(metrics),
	}

	overallScore := (scores.Identity + scores.Activity + scores.Quality + scores.Maintenance + scores.Community) / 5

	// Determine risk level
	riskLevel := "low"
	if overallScore >= 60 {
		riskLevel = "high"
	} else if overallScore >= 30 {
		riskLevel = "medium"
	}

	// Generate flags
	redFlags, warnings, positives := a.generateFlags(user, metrics, accountAge, len(repos))

	return &Analysis{
		User:         *user,
		Scores:       scores,
		OverallScore: overallScore,
		RiskLevel:    riskLevel,
		Metrics:      metrics,
		RedFlags:     redFlags,
		Warnings:     warnings,
		Positives:    positives,
		Timestamp:    now,
	}
}

func (a *Analyzer) calculateMetrics(user *GitHubUser, repos []GitHubRepo, events []GitHubEvent, accountAge int, now time.Time) Metrics {
	metrics := Metrics{
		AccountAgeDays: accountAge,
		Repos:          len(repos),
		Followers:      user.Followers,
	}

	// Analyze repos
	for _, repo := range repos {
		metrics.Stars += repo.StargazersCount
		metrics.Forks += repo.ForksCount

		if repo.Archived {
			metrics.Archived++
		}

		daysSinceUpdate := now.Sub(repo.UpdatedAt).Hours() / 24
		if daysSinceUpdate <= 30 {
			metrics.RecentlyUpdated++
		}

		if repo.Language == "JavaScript" || repo.Language == "TypeScript" {
			metrics.NPMPackages++
		} else if repo.Language == "Python" {
			metrics.PythonPackages++
		}
	}

	// Analyze events (last 90 days)
	for _, event := range events {
		daysSinceEvent := now.Sub(event.CreatedAt).Hours() / 24
		if daysSinceEvent <= 90 && event.Type == "PushEvent" {
			metrics.RecentCommits += len(event.Payload.Commits)
		}
	}

	return metrics
}

func (a *Analyzer) calculateIdentityScore(user *GitHubUser, accountAge int) float64 {
	score := 50.0

	if accountAge > 730 {
		score -= 20
	} else if accountAge > 365 {
		score -= 10
	} else if accountAge < 90 {
		score += 30
	}

	if user.Company != "" {
		score -= 10
	}
	if user.Email != "" {
		score -= 5
	}
	if user.Blog != "" {
		score -= 5
	}
	if user.TwitterUsername != "" {
		score -= 5
	}
	if user.Name == "" {
		score += 10
	}

	return clamp(score, 0, 100)
}

func (a *Analyzer) calculateActivityScore(metrics Metrics, accountAge, totalRepos int) float64 {
	score := 50.0
	commitsPerMonth := float64(metrics.RecentCommits) / 3.0

	if commitsPerMonth > 20 {
		score -= 20
	} else if commitsPerMonth > 10 {
		score -= 10
	} else if commitsPerMonth < 2 {
		score += 20
	}

	if metrics.RecentlyUpdated == 0 && totalRepos > 0 {
		score += 30
	} else if float64(metrics.RecentlyUpdated) > float64(totalRepos)*0.3 {
		score -= 10
	}

	return clamp(score, 0, 100)
}

func (a *Analyzer) calculateQualityScore(repos []GitHubRepo, metrics Metrics) float64 {
	score := 50.0

	if len(repos) == 0 {
		return score
	}

	avgStars := float64(metrics.Stars) / float64(len(repos))

	if avgStars > 50 {
		score -= 20
	} else if avgStars > 10 {
		score -= 10
	} else if avgStars < 1 {
		score += 10
	}

	if metrics.Stars > 0 && float64(metrics.Forks) > float64(metrics.Stars)*0.1 {
		score -= 10
	}

	return clamp(score, 0, 100)
}

func (a *Analyzer) calculateMaintenanceScore(metrics Metrics, totalRepos int) float64 {
	score := 50.0

	if totalRepos == 0 {
		return score
	}

	archivedRatio := float64(metrics.Archived) / float64(totalRepos)
	activeRatio := float64(metrics.RecentlyUpdated) / float64(totalRepos)

	if archivedRatio > 0.5 {
		score += 30
	} else if archivedRatio > 0.3 {
		score += 15
	}

	if activeRatio > 0.5 {
		score -= 20
	} else if activeRatio > 0.3 {
		score -= 10
	} else if activeRatio == 0 && totalRepos > 0 {
		score += 20
	}

	return clamp(score, 0, 100)
}

func (a *Analyzer) calculateCommunityScore(metrics Metrics) float64 {
	score := 50.0

	if metrics.Followers > 500 {
		score -= 25
	} else if metrics.Followers > 100 {
		score -= 15
	} else if metrics.Followers > 50 {
		score -= 10
	} else if metrics.Followers < 10 {
		score += 15
	}

	if metrics.Stars > 1000 {
		score -= 15
	} else if metrics.Stars > 100 {
		score -= 10
	} else if metrics.Stars < 10 && metrics.Repos > 5 {
		score += 10
	}

	return clamp(score, 0, 100)
}

func (a *Analyzer) generateFlags(user *GitHubUser, metrics Metrics, accountAge, totalRepos int) ([]string, []string, []string) {
	var redFlags, warnings, positives []string

	// Account age
	if accountAge < 180 {
		redFlags = append(redFlags, fmt.Sprintf("Account only %d months old - limited history", accountAge/30))
	} else if accountAge > 365 {
		positives = append(positives, fmt.Sprintf("Established account (%d years)", accountAge/365))
	}

	// Followers
	if metrics.Followers < 10 {
		warnings = append(warnings, "Low follower count - limited community validation")
	} else if metrics.Followers > 100 {
		positives = append(positives, fmt.Sprintf("Strong community following (%d followers)", metrics.Followers))
	}

	// Activity
	if metrics.RecentCommits < 10 {
		warnings = append(warnings, "Low recent activity (last 90 days)")
	} else if metrics.RecentCommits > 50 {
		positives = append(positives, fmt.Sprintf("Active contributor (%d commits in 90 days)", metrics.RecentCommits))
	}

	// Archived repos
	if totalRepos > 0 && float64(metrics.Archived)/float64(totalRepos) > 0.3 {
		redFlags = append(redFlags, fmt.Sprintf("High proportion of archived repos (%d/%d)", metrics.Archived, totalRepos))
	}

	// Contact info
	if user.Company == "" && user.Blog == "" && user.Email == "" {
		warnings = append(warnings, "No verifiable contact information or affiliation")
	} else {
		if user.Company != "" {
			positives = append(positives, fmt.Sprintf("Affiliated with: %s", user.Company))
		}
		if user.Blog != "" {
			positives = append(positives, "Has published website/blog")
		}
	}

	// Maintenance
	if metrics.RecentlyUpdated == 0 && totalRepos > 0 {
		redFlags = append(redFlags, "No repositories updated in last 30 days")
	}

	// Engagement
	if metrics.Stars < 10 && totalRepos > 5 {
		warnings = append(warnings, "Low community engagement (stars/repos ratio)")
	}

	return redFlags, warnings, positives
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// CLI output functions
func printAnalysis(analysis *Analysis) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  MCP SERVER SECURITY ANALYZER")
	fmt.Println(strings.Repeat("=", 80))

	// User info
	fmt.Printf("\n👤 User: %s (@%s)\n", analysis.User.Name, analysis.User.Login)
	if analysis.User.Bio != "" {
		fmt.Printf("   Bio: %s\n", analysis.User.Bio)
	}
	fmt.Printf("   Profile: %s\n", analysis.User.HTMLURL)

	// Overall risk
	fmt.Printf("\n🛡️  OVERALL RISK ASSESSMENT: %s\n", strings.ToUpper(analysis.RiskLevel))
	fmt.Printf("   Risk Score: %.1f/100 (lower is better)\n", analysis.OverallScore)

	// Key metrics
	fmt.Println("\n📊 KEY METRICS")
	fmt.Printf("   Account Age:        %dy %dm\n", analysis.Metrics.AccountAgeDays/365, (analysis.Metrics.AccountAgeDays%365)/30)
	fmt.Printf("   Repositories:       %d\n", analysis.Metrics.Repos)
	fmt.Printf("   Total Stars:        %d\n", analysis.Metrics.Stars)
	fmt.Printf("   Followers:          %d\n", analysis.Metrics.Followers)
	fmt.Printf("   Recent Commits:     %d (90 days)\n", analysis.Metrics.RecentCommits)
	fmt.Printf("   Recently Updated:   %d repos (30 days)\n", analysis.Metrics.RecentlyUpdated)
	fmt.Printf("   Archived:           %d repos\n", analysis.Metrics.Archived)

	// Detailed scores
	fmt.Println("\n📈 DETAILED RISK SCORES")
	fmt.Printf("   Identity:           %.1f/100\n", analysis.Scores.Identity)
	fmt.Printf("   Activity:           %.1f/100\n", analysis.Scores.Activity)
	fmt.Printf("   Quality:            %.1f/100\n", analysis.Scores.Quality)
	fmt.Printf("   Maintenance:        %.1f/100\n", analysis.Scores.Maintenance)
	fmt.Printf("   Community:          %.1f/100\n", analysis.Scores.Community)

	// Flags
	if len(analysis.RedFlags) > 0 {
		fmt.Println("\n🚨 RED FLAGS")
		for _, flag := range analysis.RedFlags {
			fmt.Printf("   • %s\n", flag)
		}
	}

	if len(analysis.Warnings) > 0 {
		fmt.Println("\n⚠️  WARNINGS")
		for _, warning := range analysis.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}

	if len(analysis.Positives) > 0 {
		fmt.Println("\n✅ POSITIVE SIGNALS")
		for _, positive := range analysis.Positives {
			fmt.Printf("   • %s\n", positive)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

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

	analyzer := NewAnalyzer(token)
	analysis, err := analyzer.Analyze(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printAnalysis(analysis)

	// Optionally save to JSON
	if len(os.Args) > 2 && os.Args[2] == "--json" {
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("%s_analysis.json", username)
		if err := os.WriteFile(filename, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n💾 Analysis saved to: %s\n", filename)
	}
}
