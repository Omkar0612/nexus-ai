package browser

import (
	"testing"
)

func TestBrowserIsAllowedValid(t *testing.T) {
	b := New(DefaultConfig())
	ok, reason := b.IsAllowed("https://github.com/Omkar0612/nexus-ai")
	if !ok {
		t.Errorf("expected allowed, got blocked: %s", reason)
	}
}

func TestBrowserBlockedLocalhost(t *testing.T) {
	b := New(DefaultConfig())
	ok, reason := b.IsAllowed("http://localhost:8080/admin")
	if ok {
		t.Error("expected localhost to be blocked")
	}
	if reason == "" {
		t.Error("expected non-empty block reason")
	}
}

func TestBrowserVisitLimit(t *testing.T) {
	b := New(DefaultConfig()) // MaxVisits = 3
	url := "https://example.com/page"
	for i := 0; i < 3; i++ {
		b.RecordVisit(url)
	}
	ok, _ := b.IsAllowed(url)
	if ok {
		t.Error("expected URL to be blocked after 3 visits")
	}
}

func TestBrowserPlanTask(t *testing.T) {
	b := New(DefaultConfig())
	actions := b.PlanTask("go to https://github.com and extract the text")
	if len(actions) == 0 {
		t.Error("expected at least 1 action")
	}
	hasNavigate := false
	for _, a := range actions {
		if a.Type == "navigate" {
			hasNavigate = true
		}
	}
	if !hasNavigate {
		t.Error("expected navigate action when URL is in task")
	}
}

func TestBrowserRunBlocked(t *testing.T) {
	b := New(DefaultConfig())
	result := b.Run("test", []BrowseAction{
		{Type: "navigate", Target: "http://localhost:9999/secret"},
	})
	if result.Success {
		t.Error("expected blocked result for localhost navigation")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestExtractLinks(t *testing.T) {
	html := `<a href="https://github.com">GitHub</a> <a href="https://golang.org">Go</a>`
	links := ExtractLinks(html)
	if len(links) < 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}
