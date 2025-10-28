package ebert

import "time"

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

type RiskScores struct {
	Identity    float64 `json:"identity"`
	Activity    float64 `json:"activity"`
	Quality     float64 `json:"quality"`
	Maintenance float64 `json:"maintenance"`
	Community   float64 `json:"community"`
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
