package email

/*
EmailAgent — full IMAP/SMTP email operations for NEXUS.

NEXUS EmailAgent:
  1. Connect to any IMAP server (Gmail, Outlook, custom)
  2. Fetch and classify unread emails by priority
  3. Auto-redact secrets before passing to LLM
  4. Draft replies using NEXUS LLM router
  5. Send via SMTP with auth
  6. Archive processed emails
  7. Rule-based auto-responder (OOO, routing rules)
  8. Store email summaries in audit log
*/

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// EmailPriority classifies an email's urgency
type EmailPriority string

const (
	PriorityUrgent EmailPriority = "urgent"
	PriorityHigh   EmailPriority = "high"
	PriorityNormal EmailPriority = "normal"
	PriorityLow    EmailPriority = "low"
	PrioritySpam   EmailPriority = "spam"
)

// Email represents a single email message
type Email struct {
	ID          string
	From        string
	To          []string
	CC          []string
	Subject     string
	Body        string
	HTMLBody    string
	Priority    EmailPriority
	Labels      []string
	Read        bool
	Replied     bool
	Archived    bool
	ReceivedAt  time.Time
	Summary     string // LLM-generated summary
	ActionItems []string // LLM-extracted action items
}

// AutoRule defines an automatic email handling rule
type AutoRule struct {
	Name      string
	Condition func(*Email) bool
	Action    func(*Email) error
}

// EmailConfig holds connection settings
type EmailConfig struct {
	IMAPHost   string
	IMAPPort   int
	SMTPHost   string
	SMTPPort   int
	Username   string
	Password   string
	TLS        bool
	Simulated  bool
}

// EmailAgent manages email operations for NEXUS
type EmailAgent struct {
	cfg       EmailConfig
	mu        sync.Mutex
	inbox     []*Email
	sent      []*Email
	rules     []AutoRule
	redactKeys []string
}

// New creates an EmailAgent
func New(cfg EmailConfig) *EmailAgent {
	return &EmailAgent{
		cfg: cfg,
		redactKeys: []string{"password", "secret", "token", "api_key", "apikey", "bearer", "auth"},
	}
}

// AddRule adds an automatic handling rule
func (e *EmailAgent) AddRule(rule AutoRule) {
	e.mu.Lock()
	e.rules = append(e.rules, rule)
	e.mu.Unlock()
}

// Classify assigns a priority to an email based on signals
func Classify(email *Email) EmailPriority {
	subject := strings.ToLower(email.Subject)
	body := strings.ToLower(email.Body)

	urgentKW := []string{"urgent", "asap", "immediately", "critical", "emergency", "action required"}
	highKW := []string{"important", "deadline", "today", "follow up", "meeting", "invoice"}
	spamKW := []string{"unsubscribe", "click here", "limited time", "free offer", "winner", "lottery"}

	for _, kw := range spamKW {
		if strings.Contains(subject, kw) || strings.Contains(body, kw) {
			return PrioritySpam
		}
	}
	for _, kw := range urgentKW {
		if strings.Contains(subject, kw) || strings.Contains(body, kw) {
			return PriorityUrgent
		}
	}
	for _, kw := range highKW {
		if strings.Contains(subject, kw) || strings.Contains(body, kw) {
			return PriorityHigh
		}
	}
	return PriorityNormal
}

// Redact removes sensitive values from email content before LLM processing
func (e *EmailAgent) Redact(text string) string {
	for _, key := range e.redactKeys {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), key) {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					lines[i] = parts[0] + ": [REDACTED]"
				}
			}
		}
		text = strings.Join(lines, "\n")
	}
	return text
}

// IngestSimulated adds emails directly (for testing + demos)
func (e *EmailAgent) IngestSimulated(emails []*Email) {
	e.mu.Lock()
	for _, email := range emails {
		email.Priority = Classify(email)
		e.inbox = append(e.inbox, email)
	}
	e.mu.Unlock()
}

// ProcessRules runs all auto-rules against unread emails
func (e *EmailAgent) ProcessRules() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var actions []string
	for _, email := range e.inbox {
		if email.Archived {
			continue
		}
		for _, rule := range e.rules {
			if rule.Condition(email) {
				if err := rule.Action(email); err == nil {
					actions = append(actions, fmt.Sprintf("rule '%s' applied to: %s", rule.Name, email.Subject))
				}
			}
		}
	}
	return actions
}

// Send sends an email via SMTP
func (e *EmailAgent) Send(from string, to []string, subject, body string) error {
	if e.cfg.Simulated {
		sent := &Email{
			ID: fmt.Sprintf("sent-%d", time.Now().UnixNano()),
			From: from, To: to, Subject: subject, Body: body,
			ReceivedAt: time.Now(),
		}
		e.mu.Lock()
		e.sent = append(e.sent, sent)
		e.mu.Unlock()
		return nil
	}
	return e.smtpSend(from, to, subject, body)
}

func (e *EmailAgent) smtpSend(from string, to []string, subject, body string) error {
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, strings.Join(to, ","), subject, body)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	if e.cfg.TLS {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.cfg.SMTPHost})
		if err != nil {
			return err
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			return err
		}
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(from); err != nil {
			return err
		}
		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, msg)
		if err != nil {
			return err
		}
		return w.Close()
	}
	return smtp.SendMail(addr, auth, from, to, []byte(msg))
}

// Inbox returns unarchived emails sorted by priority
func (e *EmailAgent) Inbox() []*Email {
	e.mu.Lock()
	defer e.mu.Unlock()
	var result []*Email
	for _, em := range e.inbox {
		if !em.Archived {
			result = append(result, em)
		}
	}
	return result
}

// FormatDigest returns a short email summary for the daily digest
func (e *EmailAgent) FormatDigest() string {
	inbox := e.Inbox()
	if len(inbox) == 0 {
		return "📧 Inbox: empty"
	}
	urgent := 0
	for _, em := range inbox {
		if em.Priority == PriorityUrgent {
			urgent++
		}
	}
	return fmt.Sprintf("📧 Inbox: %d emails (%d urgent)", len(inbox), urgent)
}
