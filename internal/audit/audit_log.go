package audit

/*
AuditLog — fully queryable agent decision log with rationale.

Security:
  - DB file created with 0600 permissions before sql.Open
  - Entry IDs use crypto/rand (not predictable UnixNano)
*/

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const sqliteTimeFormat = "2006-01-02 15:04:05.000000"

// RiskLevel classifies the risk of an agent action.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// AuditEntry is a single recorded agent decision.
type AuditEntry struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Agent        string            `json:"agent"`
	Action       string            `json:"action"`
	Rationale    string            `json:"rationale"`
	ContextUsed  string            `json:"context_used"`
	Alternatives []string          `json:"alternatives_considered"`
	Outcome      string            `json:"outcome"`
	Risk         RiskLevel         `json:"risk"`
	ApprovedBy   string            `json:"approved_by"`
	DurationMs   int64             `json:"duration_ms"`
	Meta         map[string]string `json:"meta,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// AuditQuery defines filters for querying the audit log.
type AuditQuery struct {
	UserID    string
	Agent     string
	Risk      RiskLevel
	Since     time.Time
	Until     time.Time
	Limit     int
	SearchStr string
}

// Log is the NEXUS audit logging system.
type Log struct {
	db *sql.DB
}

// randomID returns a cryptographically random hex ID with the given prefix.
func randomID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// fallback: should never happen
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}

// Open initialises (or opens) the audit log database.
func Open(dataDir string) (*Log, error) {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".nexus")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "audit.db")
	// Create the file with 0600 before sql.Open so the OS never
	// exposes it world-readable even momentarily.
	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("audit: create db file: %w", err)
	}
	f.Close()
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	l := &Log{db: db}
	return l, l.migrate()
}

func (l *Log) migrate() error {
	_, err := l.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id           TEXT PRIMARY KEY,
			user_id      TEXT NOT NULL,
			agent        TEXT NOT NULL,
			action       TEXT NOT NULL,
			rationale    TEXT DEFAULT '',
			context_used TEXT DEFAULT '',
			alternatives TEXT DEFAULT '[]',
			outcome      TEXT DEFAULT '',
			risk         TEXT DEFAULT 'low',
			approved_by  TEXT DEFAULT 'auto',
			duration_ms  INTEGER DEFAULT 0,
			meta         TEXT DEFAULT '{}',
			created_at   TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now'))
		);
		CREATE INDEX IF NOT EXISTS idx_audit_user  ON audit_log(user_id, created_at);
		CREATE INDEX IF NOT EXISTS idx_audit_risk  ON audit_log(risk);
		CREATE INDEX IF NOT EXISTS idx_audit_agent ON audit_log(agent);
	`)
	return err
}

// Record writes an audit entry.
func (l *Log) Record(entry AuditEntry) error {
	if entry.ID == "" {
		entry.ID = randomID("aud")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	altsJSON, _ := json.Marshal(entry.Alternatives)
	metaJSON, _ := json.Marshal(entry.Meta)
	createdAtStr := entry.CreatedAt.UTC().Format(sqliteTimeFormat)
	_, err := l.db.Exec(
		`INSERT INTO audit_log
		 (id,user_id,agent,action,rationale,context_used,alternatives,outcome,risk,approved_by,duration_ms,meta,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		entry.ID, entry.UserID, entry.Agent, entry.Action,
		entry.Rationale, entry.ContextUsed, string(altsJSON),
		entry.Outcome, string(entry.Risk), entry.ApprovedBy,
		entry.DurationMs, string(metaJSON), createdAtStr,
	)
	return err
}

// Query returns audit entries matching the given filters.
func (l *Log) Query(q AuditQuery) ([]AuditEntry, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if q.UserID != "" {
		where = append(where, "user_id = ?")
		args = append(args, q.UserID)
	}
	if q.Agent != "" {
		where = append(where, "agent = ?")
		args = append(args, q.Agent)
	}
	if q.Risk != "" {
		where = append(where, "risk = ?")
		args = append(args, string(q.Risk))
	}
	if !q.Since.IsZero() {
		where = append(where, "created_at >= ?")
		args = append(args, q.Since.UTC().Format(sqliteTimeFormat))
	}
	if !q.Until.IsZero() {
		where = append(where, "created_at <= ?")
		args = append(args, q.Until.UTC().Format(sqliteTimeFormat))
	}
	if q.SearchStr != "" {
		where = append(where, "(action LIKE ? OR rationale LIKE ?)")
		args = append(args, "%"+q.SearchStr+"%", "%"+q.SearchStr+"%")
	}
	limit := 50
	if q.Limit > 0 {
		limit = q.Limit
	}
	query := fmt.Sprintf(
		`SELECT id,user_id,agent,action,rationale,context_used,alternatives,outcome,risk,approved_by,duration_ms,meta,created_at
		 FROM audit_log WHERE %s ORDER BY created_at DESC LIMIT %d`,
		strings.Join(where, " AND "), limit,
	)
	rows, err := l.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var altsJSON, metaJSON, risk, createdAtStr string
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Agent, &e.Action, &e.Rationale,
			&e.ContextUsed, &altsJSON, &e.Outcome, &risk,
			&e.ApprovedBy, &e.DurationMs, &metaJSON, &createdAtStr,
		); err != nil {
			return nil, err
		}
		e.Risk = RiskLevel(risk)
		_ = json.Unmarshal([]byte(altsJSON), &e.Alternatives)
		_ = json.Unmarshal([]byte(metaJSON), &e.Meta)
		if t, err := time.ParseInLocation(sqliteTimeFormat, createdAtStr, time.UTC); err == nil {
			e.CreatedAt = t
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// FormatReport renders audit entries as a human-readable report.
func FormatReport(entries []AuditEntry) string {
	if len(entries) == 0 {
		return "📋 No audit entries found for this query."
	}
	riskIcons := map[RiskLevel]string{
		RiskLow:    "🟢",
		RiskMedium: "🟡",
		RiskHigh:   "🔴",
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 **NEXUS Audit Log** (%d entries)\n\n", len(entries)))
	for _, e := range entries {
		icon := riskIcons[e.Risk]
		if icon == "" {
			icon = "⚪"
		}
		sb.WriteString(fmt.Sprintf("%s [%s] **%s** → %s\n", icon, e.Agent, e.Action, e.Outcome))
		if e.Rationale != "" {
			sb.WriteString(fmt.Sprintf("   💭 Why: %s\n", e.Rationale))
		}
		if len(e.Alternatives) > 0 {
			sb.WriteString(fmt.Sprintf("   🔀 Considered: %s\n", strings.Join(e.Alternatives, ", ")))
		}
		if e.ApprovedBy != "auto" && e.ApprovedBy != "" {
			sb.WriteString(fmt.Sprintf("   ✅ Approved by: %s\n", e.ApprovedBy))
		}
		sb.WriteString(fmt.Sprintf("   🕐 %s (%dms)\n\n", e.CreatedAt.Format("Jan 2 15:04:05"), e.DurationMs))
	}
	return sb.String()
}

// ExportJSON returns all entries as a JSON byte slice (for compliance export).
func (l *Log) ExportJSON(q AuditQuery) ([]byte, error) {
	entries, err := l.Query(q)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(entries, "", "  ")
}

// ClassifyRisk auto-classifies an action's risk level.
func ClassifyRisk(action string) RiskLevel {
	action = strings.ToLower(action)
	highRisk := []string{"delete", "remove", "drop", "send", "post", "pay", "transfer", "deploy", "execute", "run", "purchase"}
	mediumRisk := []string{"update", "modify", "write", "create", "upload", "download", "notify"}
	for _, kw := range highRisk {
		if strings.Contains(action, kw) {
			return RiskHigh
		}
	}
	for _, kw := range mediumRisk {
		if strings.Contains(action, kw) {
			return RiskMedium
		}
	}
	return RiskLow
}

// Close shuts down the audit log.
func (l *Log) Close() error { return l.db.Close() }
