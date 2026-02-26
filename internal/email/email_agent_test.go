package email

import (
	"strings"
	"testing"
)

func TestEmailClassify(t *testing.T) {
	cases := []struct {
		subject string
		body    string
		want    EmailPriority
	}{
		{"URGENT: Server down", "", PriorityUrgent},
		{"Meeting deadline today", "", PriorityHigh},
		{"Hello", "How are you?", PriorityNormal},
		{"Win a lottery prize", "Click here", PrioritySpam},
	}
	for _, c := range cases {
		email := &Email{Subject: c.subject, Body: c.body}
		got := Classify(email)
		if got != c.want {
			t.Errorf("Classify(%q) = %s, want %s", c.subject, got, c.want)
		}
	}
}

func TestEmailRedact(t *testing.T) {
	a := New(EmailConfig{Simulated: true})
	input := "Here is your api_key: super_secret_123\nOther info: normal"
	redacted := a.Redact(input)
	if strings.Contains(redacted, "super_secret_123") {
		t.Error("expected secret to be redacted")
	}
}

func TestEmailSendSimulated(t *testing.T) {
	a := New(EmailConfig{Simulated: true})
	err := a.Send("nexus@example.com", []string{"omkar@example.com"}, "Test", "Hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestEmailIngestAndInbox(t *testing.T) {
	a := New(EmailConfig{Simulated: true})
	a.IngestSimulated([]*Email{
		{ID: "1", Subject: "URGENT: call me", Body: "asap please"},
		{ID: "2", Subject: "Newsletter", Body: "unsubscribe here"},
	})
	inbox := a.Inbox()
	if len(inbox) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(inbox))
	}
	argentFound := false
	for _, em := range inbox {
		if em.Priority == PriorityUrgent {
			argentFound = true
		}
	}
	if !argentFound {
		t.Error("expected at least one urgent email")
	}
}

func TestEmailAutoRule(t *testing.T) {
	a := New(EmailConfig{Simulated: true})
	archived := 0
	a.AddRule(AutoRule{
		Name: "archive-spam",
		Condition: func(em *Email) bool { return em.Priority == PrioritySpam },
		Action:    func(em *Email) error { em.Archived = true; archived++; return nil },
	})
	a.IngestSimulated([]*Email{
		{ID: "1", Subject: "Free lottery", Body: "click here"},
		{ID: "2", Subject: "Real work", Body: "normal content"},
	})
	a.ProcessRules()
	if archived != 1 {
		t.Errorf("expected 1 archived email, got %d", archived)
	}
}
