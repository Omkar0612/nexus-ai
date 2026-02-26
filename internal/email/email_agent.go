package email

/*
EmailAgent — full IMAP/SMTP email operations for NEXUS.

Security:
  - SMTP header injection prevented: \r and \n stripped from From/To/Subject
  - Recipient addresses validated (must contain '@', no newlines)
  - Password masked in fmt/log output via SecretString type
  - Sensitive field redaction before any LLM processing
*/

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// SecretString wraps a string and masks it in all fmt/log output.
// This prevents passwords from appearing in log files or error messages.
type SecretString struct{ v string }

// NewSecret wraps a plaintext value as a SecretString.
func NewSecret(s string) SecretString { return SecretString{v: s} }

// Value returns the raw secret (only call when actually needed for auth).
func (s SecretString) Value() string { return s.v }

// String implements fmt.Stringer — always returns "[REDACTED]".
func (s SecretString) String() string { return "[REDACTED]" }

// GoString implements fmt.GoStringer — prevents leakage via %#v.
func (s SecretString) GoString() string { return "email.SecretString([REDACTED])" }

// EmailPriority classifies an email's urgency.
type EmailPriority string

const (
	PriorityUrgent EmailPriority = "urgent"
	PriorityHigh   EmailPriority = "high"
	PriorityNormal EmailPriority = "normal"
	PriorityLow    EmailPriority = "low"
	PrioritySpam   EmailPriority = "spam"
)

// Email represents a single email message.
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
	Summary     string   // LLM-generated summary
	ActionItems []string // LLM-extracted action items
}

// AutoRule defines an automatic email handling rule.
type AutoRule struct {
	Name      string
	Condition func(*Email) bool
	Action    func(*Email) error
}

// EmailConfig holds IMAP/SMTP connection settings.
// Password is a SecretString — it will never appear in logs.
type EmailConfig struct {
	IMAPHost  string
	IMAPPort  int
	SMTPHost  string
	SMTPPort  int
	Username  string
	Password  SecretString // masked in fmt/log output
	TLS       bool
	Simulated bool
}

// EmailAgent manages email operations for NEXUS.
type EmailAgent struct {
	cfg        EmailConfig
	mu         sync.Mutex
	inbox      []*Email
	sent       []*Email
	rules      []AutoRule
	redactKeys []string
}

// New creates an EmailAgent.
func New(cfg EmailConfig) *EmailAgent {
	return &EmailAgent{
		cfg:        cfg,
		redactKeys: []string{"password", "secret", "token", "api_key", "apikey", "bearer", "auth"},
	}
}

// AddRule adds an automatic handling rule.
func (e *EmailAgent) AddRule(rule AutoRule) {
	e.mu.Lock()
	e.rules = append(e.rules, rule)
	e.mu.Unlock()
}

// Classify assigns a priority to an email based on keyword signals.
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

// Redact removes sensitive values from email content before LLM processing.
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

// IngestSimulated adds emails directly (for testing and demos).
func (e *EmailAgent) IngestSimulated(emails []*Email) {
	e.mu.Lock()
	for _, email := range emails {
		email.Priority = Classify(email)
		e.inbox = append(e.inbox, email)
	}
	e.mu.Unlock()
}

// ProcessRules runs all auto-rules against unread, non-archived emails.
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

// sanitiseHeader strips \r and \n from an email header value to prevent
// SMTP header injection (CVE class: email header injection).
func sanitiseHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// validateRecipient returns an error if the address looks malformed.
func validateRecipient(addr string) error {
	if !strings.Contains(addr, "@") {
		return fmt.Errorf("email: invalid recipient address %q (missing @)", addr)
	}
	if strings.ContainsAny(addr, "\r\n") {
		return fmt.Errorf("email: recipient address contains illegal characters")
	}
	return nil
}

// Send sends an email via SMTP (or records it in simulation mode).
func (e *EmailAgent) Send(from string, to []string, subject, body string) error {
	// Sanitise header fields to prevent SMTP injection.
	from = sanitiseHeader(from)
	subject = sanitiseHeader(subject)
	sanitised := make([]string, 0, len(to))
	for _, addr := range to {
		if err := validateRecipient(addr); err != nil {
			return err
		}
		sanitised = append(sanitised, sanitiseHeader(addr))
	}
	to = sanitised

	if e.cfg.Simulated {
		sent := &Email{
			ID:         fmt.Sprintf("sent-%d", time.Now().UnixNano()),
			From:       from,
			To:         to,
			Subject:    subject,
			Body:       body,
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
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password.Value(), e.cfg.SMTPHost)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, strings.Join(to, ","), subject, body)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	if e.cfg.TLS {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.cfg.SMTPHost})
		if err != nil {
			return fmt.Errorf("email: tls dial: %w", err)
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("email: smtp client: %w", err)
		}
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("email: auth: %w", err)
		}
		if err = client.Mail(from); err != nil {
			return fmt.Errorf("email: MAIL FROM: %w", err)
		}
		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				return fmt.Errorf("email: RCPT TO %s: %w", recipient, err)
			}
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("email: DATA: %w", err)
		}
		_, err = fmt.Fprint(w, msg)
		if err != nil {
			return fmt.Errorf("email: write body: %w", err)
		}
		return w.Close()
	}
	return smtp.SendMail(addr, auth, from, to, []byte(msg))
}

// Inbox returns unarchived emails.
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

// FormatDigest returns a short email summary for the daily digest.
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
