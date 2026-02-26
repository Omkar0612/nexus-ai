package github

import (
	"testing"
)

func simConfig() GitHubConfig {
	return GitHubConfig{Simulated: true, Owner: "Omkar0612", Repo: "nexus-ai"}
}

func TestGitHubOpenIssue(t *testing.T) {
	g := New(simConfig())
	issue, err := g.OpenIssue("Test issue", "body", []string{"bug"}, nil)
	if err != nil {
		t.Fatalf("OpenIssue: %v", err)
	}
	if issue.Number == 0 {
		t.Error("expected non-zero issue number")
	}
	if issue.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestGitHubCommentSimulated(t *testing.T) {
	g := New(simConfig())
	if err := g.CommentOnIssue(1, "NEXUS auto-comment"); err != nil {
		t.Fatalf("CommentOnIssue: %v", err)
	}
}

func TestGitHubListOpenIssues(t *testing.T) {
	g := New(simConfig())
	issues, err := g.ListOpenIssues()
	if err != nil {
		t.Fatalf("ListOpenIssues: %v", err)
	}
	if len(issues) == 0 {
		t.Error("expected simulated issues")
	}
}

func TestGitHubSearchCode(t *testing.T) {
	g := New(simConfig())
	result, err := g.SearchCode("CostTracker")
	if err != nil {
		t.Fatalf("SearchCode: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty search result")
	}
}

func TestFormatIssueList(t *testing.T) {
	issues := []Issue{
		{Number: 1, Title: "Bug in loop detector", URL: "https://github.com/test/test/issues/1"},
	}
	output := FormatIssueList(issues)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !containsStr(output, "Bug in loop detector") {
		t.Error("expected issue title in output")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && (s[:len(sub)] == sub || containsStr(s[1:], sub))))
}
