package github

/*
GitHubAgent — autonomous GitHub operations for NEXUS.

Security:
  - Token stored as SecretString — never appears in logs or fmt output
  - API error responses capped at 512 bytes (no internal detail leakage)
  - SearchCode query URL-escaped to prevent query-string injection
  - Response bodies limited to 4 MB (DoS protection)
*/

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SecretString wraps a string and masks it in all fmt/log output.
type SecretString struct{ v string }

// NewSecret wraps a plaintext value as a SecretString.
func NewSecret(s string) SecretString { return SecretString{v: s} }

// Value returns the raw token (only call when building HTTP headers).
func (s SecretString) Value() string { return s.v }

// String implements fmt.Stringer — always returns "[REDACTED]".
func (s SecretString) String() string { return "[REDACTED]" }

// GoString implements fmt.GoStringer — prevents leakage via %#v.
func (s SecretString) GoString() string { return "github.SecretString([REDACTED])" }

// maxResponseBytes is the maximum number of bytes read from any GitHub API response.
const maxResponseBytes = 4 * 1024 * 1024 // 4 MB

// GitHubConfig holds GitHub API credentials.
// Token is a SecretString — it will never appear in log files.
type GitHubConfig struct {
	Token     SecretString
	Owner     string
	Repo      string
	BaseURL   string // default: https://api.github.com
	Simulated bool
}

// Issue represents a GitHub issue.
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	Labels    []string  `json:"labels"`
	Assignees []string  `json:"assignees"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"html_url"`
}

// PullRequest represents a GitHub PR.
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Head   string `json:"head"`
	Base   string `json:"base"`
	URL    string `json:"html_url"`
}

// GitHubAgent performs autonomous GitHub operations.
type GitHubAgent struct {
	cfg    GitHubConfig
	client *http.Client
}

// New creates a GitHubAgent.
func New(cfg GitHubConfig) *GitHubAgent {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.github.com"
	}
	return &GitHubAgent{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// OpenIssue creates a new GitHub issue.
func (g *GitHubAgent) OpenIssue(title, body string, labels, assignees []string) (*Issue, error) {
	if g.cfg.Simulated {
		return &Issue{
			Number: 999, Title: title, Body: body,
			Labels: labels, Assignees: assignees,
			URL: fmt.Sprintf("https://github.com/%s/%s/issues/999", g.cfg.Owner, g.cfg.Repo),
		}, nil
	}
	payload := map[string]interface{}{
		"title":     title,
		"body":      body,
		"labels":    labels,
		"assignees": assignees,
	}
	var issue Issue
	if err := g.post(fmt.Sprintf("/repos/%s/%s/issues", g.cfg.Owner, g.cfg.Repo), payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// CommentOnIssue adds a comment to an issue or PR.
func (g *GitHubAgent) CommentOnIssue(number int, body string) error {
	if g.cfg.Simulated {
		return nil
	}
	payload := map[string]string{"body": body}
	return g.post(fmt.Sprintf("/repos/%s/%s/issues/%d/comments", g.cfg.Owner, g.cfg.Repo, number), payload, nil)
}

// CreateBranch creates a new branch from a base SHA.
func (g *GitHubAgent) CreateBranch(name, baseSHA string) error {
	if g.cfg.Simulated {
		return nil
	}
	payload := map[string]string{
		"ref": "refs/heads/" + name,
		"sha": baseSHA,
	}
	return g.post(fmt.Sprintf("/repos/%s/%s/git/refs", g.cfg.Owner, g.cfg.Repo), payload, nil)
}

// ListOpenIssues returns open issues for the configured repo.
func (g *GitHubAgent) ListOpenIssues() ([]Issue, error) {
	if g.cfg.Simulated {
		return []Issue{
			{Number: 1, Title: "Fix CI", State: "open", URL: "https://github.com/test/test/issues/1"},
			{Number: 2, Title: "Add tests", State: "open", URL: "https://github.com/test/test/issues/2"},
		}, nil
	}
	var issues []Issue
	if err := g.get(fmt.Sprintf("/repos/%s/%s/issues?state=open", g.cfg.Owner, g.cfg.Repo), &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// SearchCode searches for code across the repo.
// The query is URL-escaped to prevent query-string injection via special characters.
func (g *GitHubAgent) SearchCode(query string) (string, error) {
	if g.cfg.Simulated {
		return fmt.Sprintf("[simulated] code search for: %s", query), nil
	}
	var result map[string]interface{}
	// url.QueryEscape prevents '&', '#', '+' etc. from injecting extra params.
	path := fmt.Sprintf("/search/code?q=%s+repo:%s/%s",
		url.QueryEscape(query), g.cfg.Owner, g.cfg.Repo)
	if err := g.get(path, &result); err != nil {
		return "", err
	}
	count, _ := result["total_count"].(float64)
	return fmt.Sprintf("%d results for '%s'", int(count), query), nil
}

// FormatIssueList returns a formatted issue list string.
func FormatIssueList(issues []Issue) string {
	if len(issues) == 0 {
		return "✅ No open issues."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 **Open Issues** (%d)\n\n", len(issues)))
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("#%d %s\n   %s\n", issue.Number, issue.Title, issue.URL))
	}
	return sb.String()
}

func (g *GitHubAgent) post(path string, payload interface{}, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, g.cfg.BaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	g.setHeaders(req)
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("github: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		// Cap at 512 bytes — avoids leaking internal GitHub error details
		// and prevents a crafted response from allocating huge buffers.
		_, _ = io.ReadAll(io.LimitReader(resp.Body, 512)) // drain
		return fmt.Errorf("github API error %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(out)
	}
	return nil
}

func (g *GitHubAgent) get(path string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, g.cfg.BaseURL+path, nil)
	if err != nil {
		return err
	}
	g.setHeaders(req)
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("github: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		_, _ = io.ReadAll(io.LimitReader(resp.Body, 512)) // drain
		return fmt.Errorf("github API error %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(out)
}

func (g *GitHubAgent) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+g.cfg.Token.Value())
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NEXUS-GitHubAgent/1.7")
}
