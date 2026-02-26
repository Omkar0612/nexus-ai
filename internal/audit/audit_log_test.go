package audit

import (
	"testing"
	"time"
)

func TestAuditLogRecordAndQuery(t *testing.T) {
	l, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer l.Close()

	entry := AuditEntry{
		UserID:       "user1",
		Agent:        "drift",
		Action:       "scan memory for stalled tasks",
		Rationale:    "User has been inactive for 3 hours",
		ContextUsed:  "last 100 episodic memories",
		Alternatives: []string{"skip scan", "partial scan"},
		Outcome:      "found 2 stalled tasks",
		Risk:         RiskLow,
		ApprovedBy:   "auto",
		DurationMs:   142,
	}
	if err := l.Record(entry); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := l.Query(AuditQuery{UserID: "user1", Limit: 10})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Agent != "drift" {
		t.Errorf("unexpected agent: %s", entries[0].Agent)
	}
	if len(entries[0].Alternatives) != 2 {
		t.Errorf("expected 2 alternatives, got %d", len(entries[0].Alternatives))
	}
}

func TestAuditQueryByRisk(t *testing.T) {
	l, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer l.Close()

	_ = l.Record(AuditEntry{UserID: "u1", Agent: "a", Action: "read", Risk: RiskLow, ApprovedBy: "auto"})
	_ = l.Record(AuditEntry{UserID: "u1", Agent: "b", Action: "delete file", Risk: RiskHigh, ApprovedBy: "auto"})

	high, _ := l.Query(AuditQuery{UserID: "u1", Risk: RiskHigh})
	if len(high) != 1 {
		t.Errorf("expected 1 high-risk entry, got %d", len(high))
	}
}

func TestAuditQuerySince(t *testing.T) {
	l, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer l.Close()

	_ = l.Record(AuditEntry{UserID: "u1", Agent: "a", Action: "old action", Risk: RiskLow, ApprovedBy: "auto"})
	since := time.Now()
	_ = l.Record(AuditEntry{UserID: "u1", Agent: "b", Action: "new action", Risk: RiskLow, ApprovedBy: "auto"})

	recent, _ := l.Query(AuditQuery{UserID: "u1", Since: since})
	if len(recent) != 1 {
		t.Errorf("expected 1 recent entry, got %d", len(recent))
	}
}

func TestClassifyRisk(t *testing.T) {
	cases := []struct {
		action string
		want   RiskLevel
	}{
		{"delete user file", RiskHigh},
		{"send telegram message", RiskHigh},
		{"update config", RiskMedium},
		{"read memory", RiskLow},
		{"search web", RiskLow},
	}
	for _, c := range cases {
		got := ClassifyRisk(c.action)
		if got != c.want {
			t.Errorf("ClassifyRisk(%q) = %s, want %s", c.action, got, c.want)
		}
	}
}

func TestExportJSON(t *testing.T) {
	l, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer l.Close()

	_ = l.Record(AuditEntry{UserID: "u1", Agent: "vault", Action: "store secret", Risk: RiskMedium, ApprovedBy: "auto"})
	data, err := l.ExportJSON(AuditQuery{UserID: "u1"})
	if err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON export")
	}
}
