package digest

import (
	"testing"
)

func TestDigestBuildEmpty(t *testing.T) {
	d := New("Omkar", "Asia/Dubai")
	report := d.Build()
	if report.TotalItems != 0 {
		t.Errorf("expected 0 items, got %d", report.TotalItems)
	}
	output := report.FormatTelegram()
	if output == "" {
		t.Error("expected non-empty output even with zero items")
	}
}

func TestDigestBuildWithFeeds(t *testing.T) {
	d := New("Omkar", "Asia/Dubai")
	d.AddFeed("drift", DriftFeed([]string{"Launch page stalled for 3 days", "Follow up with designer not done"}))
	d.AddFeed("cost", CostFeed(0.85, 1.00))
	d.AddFeed("audit", HighRiskAuditFeed([]string{"delete old backups"}))
	d.AddFeed("goals", GoalsFeed([]string{"Ship NEXUS v2"}))

	report := d.Build()
	if report.TotalItems < 4 {
		t.Errorf("expected at least 4 items, got %d", report.TotalItems)
	}
	if len(report.HighPriority) == 0 {
		t.Error("expected high priority items from audit feed")
	}
}

func TestDigestFormatTelegram(t *testing.T) {
	d := New("Omkar", "UTC")
	d.AddFeed("drift", DriftFeed([]string{"Task overdue"}))
	report := d.Build()
	output := report.FormatTelegram()
	if len(output) < 20 {
		t.Error("digest output too short")
	}
}

func TestDigestToJSON(t *testing.T) {
	d := New("Omkar", "UTC")
	d.AddFeed("cost", CostFeed(0.10, 1.00))
	report := d.Build()
	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if len(data) < 10 {
		t.Error("JSON output too short")
	}
}

func TestDigestCostFeedHighAlert(t *testing.T) {
	feed := CostFeed(0.95, 1.00)
	items := feed()
	if len(items) == 0 {
		t.Fatal("expected items from cost feed")
	}
	if items[0].Priority != 3 {
		t.Errorf("expected priority 3 for 95%% spend, got %d", items[0].Priority)
	}
}
