package phone

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func simConfig() PhoneConfig {
	return PhoneConfig{Simulated: true, FromNumber: "+10000000000", RateLimit: 10}
}

func TestPhoneCallSimulated(t *testing.T) {
	p := New(simConfig())
	rec, err := p.Call("+971500000000", "Hello from NEXUS")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if rec.Status != CallCompleted {
		t.Errorf("expected completed, got %s", rec.Status)
	}
}

func TestPhoneSMSSimulated(t *testing.T) {
	p := New(simConfig())
	rec, err := p.SMS("+971500000000", "NEXUS drift alert")
	if err != nil {
		t.Fatalf("SMS: %v", err)
	}
	if rec.Type != "sms" {
		t.Errorf("expected sms type, got %s", rec.Type)
	}
}

func TestPhoneRateLimit(t *testing.T) {
	pl := New(PhoneConfig{Simulated: true, FromNumber: "+1000", RateLimit: 2})
	_, _ = pl.Call("+1", "msg1")
	_, _ = pl.Call("+1", "msg2")
	_, err := pl.Call("+1", "msg3")
	if err == nil {
		t.Error("expected rate limit error on 3rd call")
	}
}

func TestPhoneWebhook(t *testing.T) {
	p := New(simConfig())
	var received *CallRecord
	p.SetInboundHandler(func(rec CallRecord) { received = &rec })

	body := strings.NewReader("SmsSid=SM123&From=%2B971500000000&To=%2B1000&Body=nexus+status")
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	p.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if received == nil {
		t.Error("expected inbound handler to fire")
	}
}

func TestPhoneHistory(t *testing.T) {
	p := New(simConfig())
	p.SMS("+1", "test")
	p.Call("+1", "hello")
	history := p.History()
	if len(history) < 2 {
		t.Errorf("expected 2 records, got %d", len(history))
	}
}
