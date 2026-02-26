package phone

/*
PhoneAgent — make and receive calls + SMS via Twilio.

NEXUS PhoneAgent:
  1. Initiate outbound calls with TTS spoken message
  2. Receive inbound calls with IVR routing
  3. Send SMS alerts (budget breach, HITL approval, drift alerts)
  4. Receive SMS commands ('nexus drift', 'nexus goals')
  5. Store call + SMS transcripts for audit log
  6. Webhook handler for Twilio callbacks (TwiML responses)
  7. Rate limiting to prevent runaway calls
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// CallStatus represents the state of a phone call
type CallStatus string

const (
	CallQueued     CallStatus = "queued"
	CallInitiated  CallStatus = "initiated"
	CallRinging    CallStatus = "ringing"
	CallInProgress CallStatus = "in-progress"
	CallCompleted  CallStatus = "completed"
	CallFailed     CallStatus = "failed"
)

// CallRecord stores a call or SMS event
type CallRecord struct {
	SID        string
	Direction  string // outbound | inbound
	Type       string // call | sms
	From       string
	To         string
	Body       string // SMS body or call summary
	Status     CallStatus
	Duration   int // seconds
	Cost       float64
	CreatedAt  time.Time
	Transcript string
}

// PhoneConfig holds Twilio credentials and settings
type PhoneConfig struct {
	AccountSID  string
	AuthToken   string
	FromNumber  string // your Twilio number
	WebhookURL  string // public URL for Twilio callbacks
	RateLimit   int    // max calls per hour
	Simulated   bool   // true = log only, no real API calls
}

// PhoneAgent manages calls and SMS for NEXUS
type PhoneAgent struct {
	cfg      PhoneConfig
	mu       sync.Mutex
	records  []CallRecord
	callsThisHour int
	hourWindow time.Time
	onInbound  func(CallRecord)
}

// New creates a PhoneAgent
func New(cfg PhoneConfig) *PhoneAgent {
	return &PhoneAgent{
		cfg:        cfg,
		hourWindow: time.Now(),
	}
}

// SetInboundHandler sets the callback for incoming calls/SMS
func (p *PhoneAgent) SetInboundHandler(fn func(CallRecord)) {
	p.onInbound = fn
}

// Call initiates an outbound phone call with a spoken message
func (p *PhoneAgent) Call(to, message string) (*CallRecord, error) {
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}
	rec := &CallRecord{
		SID:       fmt.Sprintf("CA%d", time.Now().UnixNano()),
		Direction: "outbound",
		Type:      "call",
		From:      p.cfg.FromNumber,
		To:        to,
		Body:      message,
		Status:    CallInitiated,
		CreatedAt: time.Now(),
	}
	if p.cfg.Simulated {
		rec.Status = CallCompleted
		rec.Duration = 30
		p.store(rec)
		return rec, nil
	}
	if err := p.twilioCall(to, message, rec); err != nil {
		rec.Status = CallFailed
		p.store(rec)
		return rec, err
	}
	p.store(rec)
	return rec, nil
}

// SMS sends a text message
func (p *PhoneAgent) SMS(to, body string) (*CallRecord, error) {
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}
	rec := &CallRecord{
		SID:       fmt.Sprintf("SM%d", time.Now().UnixNano()),
		Direction: "outbound",
		Type:      "sms",
		From:      p.cfg.FromNumber,
		To:        to,
		Body:      body,
		Status:    CallInitiated,
		CreatedAt: time.Now(),
	}
	if p.cfg.Simulated {
		rec.Status = CallCompleted
		p.store(rec)
		return rec, nil
	}
	if err := p.twilioSMS(to, body, rec); err != nil {
		rec.Status = CallFailed
		p.store(rec)
		return rec, err
	}
	p.store(rec)
	return rec, nil
}

func (p *PhoneAgent) twilioCall(to, message string, rec *CallRecord) error {
	twiml := fmt.Sprintf(`<Response><Say voice="alice">%s</Say></Response>`, message)
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", p.cfg.FromNumber)
	data.Set("Twiml", twiml)
	_, err := p.twilioRequest("Calls", data)
	if err != nil {
		return err
	}
	rec.Status = CallInitiated
	return nil
}

func (p *PhoneAgent) twilioSMS(to, body string, rec *CallRecord) error {
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", p.cfg.FromNumber)
	data.Set("Body", body)
	_, err := p.twilioRequest("Messages", data)
	if err != nil {
		return err
	}
	rec.Status = CallCompleted
	return nil
}

func (p *PhoneAgent) twilioRequest(endpoint string, data url.Values) ([]byte, error) {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/%s.json",
		p.cfg.AccountSID, endpoint)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.cfg.AccountSID, p.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result []byte
	json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("twilio error: %s", resp.Status)
	}
	return result, nil
}

// HandleWebhook processes inbound Twilio webhook callbacks
func (p *PhoneAgent) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	callType := "sms"
	if r.FormValue("CallSid") != "" {
		callType = "call"
	}
	rec := CallRecord{
		SID:       r.FormValue("CallSid") + r.FormValue("SmsSid"),
		Direction: "inbound",
		Type:      callType,
		From:      r.FormValue("From"),
		To:        r.FormValue("To"),
		Body:      r.FormValue("Body"),
		Status:    CallInProgress,
		CreatedAt: time.Now(),
	}
	p.store(&rec)
	if p.onInbound != nil {
		p.onInbound(rec)
	}
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, `<Response><Say voice="alice">NEXUS received your message. Processing now.</Say></Response>`)
}

func (p *PhoneAgent) checkRateLimit() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if time.Since(p.hourWindow) > time.Hour {
		p.callsThisHour = 0
		p.hourWindow = time.Now()
	}
	if p.cfg.RateLimit > 0 && p.callsThisHour >= p.cfg.RateLimit {
		return fmt.Errorf("rate limit reached: max %d calls/hour", p.cfg.RateLimit)
	}
	p.callsThisHour++
	return nil
}

func (p *PhoneAgent) store(rec *CallRecord) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.records = append(p.records, *rec)
	if len(p.records) > 500 {
		p.records = p.records[len(p.records)-500:]
	}
}

// History returns all stored call/SMS records
func (p *PhoneAgent) History() []CallRecord {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]CallRecord, len(p.records))
	copy(result, p.records)
	return result
}
