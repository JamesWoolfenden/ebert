package ebert

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
