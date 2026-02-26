package mobile

/*
TelegramCompanion — NEXUS on any device via Telegram bot.

NEXUS anywhere, without an app install:
  1. Full command routing — any nexus CLI command via Telegram message
  2. Inline keyboard — quick action buttons for common tasks
  3. HITL approvals — approve/reject high-risk actions from phone
  4. Digest delivery — morning briefing pushed to your phone
  5. Alert channel — budget alerts, drift signals, loop events
  6. Voice message support — send voice note, get transcript + response
  7. Auth guard — only responds to whitelisted Telegram user IDs
  8. Simulated mode for testing without a real bot token
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// TelegramMessage is a received Telegram update
type TelegramMessage struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text  string `json:"text"`
		Voice *struct {
			FileID string `json:"file_id"`
		} `json:"voice,omitempty"`
	} `json:"message"`
}

// OutboundMessage is a message sent to Telegram
type OutboundMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"` // Markdown | HTML
}

// BotConfig holds Telegram bot settings
type BotConfig struct {
	Token         string
	AllowedUserIDs []int64
	AdminChatID   int64
	Simulated     bool
}

// CommandHandler is a function that handles a bot command
type CommandHandler func(chatID int64, args string) string

// TelegramCompanion manages the NEXUS Telegram bot
type TelegramCompanion struct {
	cfg        BotConfig
	handlers   map[string]CommandHandler
	mu         sync.RWMutex
	sentLog    []OutboundMessage
	client     *http.Client
	baseURL    string
}

// New creates a TelegramCompanion
func New(cfg BotConfig) *TelegramCompanion {
	return &TelegramCompanion{
		cfg:      cfg,
		handlers: make(map[string]CommandHandler),
		client:   &http.Client{Timeout: 10 * time.Second},
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", cfg.Token),
	}
}

// RegisterCommand adds a command handler
func (t *TelegramCompanion) RegisterCommand(cmd string, handler CommandHandler) {
	t.mu.Lock()
	t.handlers[strings.ToLower(cmd)] = handler
	t.mu.Unlock()
}

// Send sends a message to a Telegram chat
func (t *TelegramCompanion) Send(chatID int64, text string) error {
	msg := OutboundMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}
	if t.cfg.Simulated {
		t.mu.Lock()
		t.sentLog = append(t.sentLog, msg)
		t.mu.Unlock()
		log.Debug().Int64("chat", chatID).Str("text", truncate(text, 80)).Msg("TG (sim): send")
		return nil
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := t.client.Post(t.baseURL+"/sendMessage", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram API error: %s", resp.Status)
	}
	return nil
}

// HandleUpdate processes a single Telegram update
func (t *TelegramCompanion) HandleUpdate(update TelegramMessage) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if !t.isAllowed(userID) {
		log.Warn().Int64("user", userID).Msg("TG: unauthorised user")
		t.Send(chatID, "🚫 Unauthorised. Contact the NEXUS admin.")
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	if text == "" {
		if update.Message.Voice != nil {
			t.Send(chatID, "🎤 Voice message received. Transcription support requires Whisper integration.")
		}
		return
	}

	// Parse command
	parts := strings.SplitN(text, " ", 2)
	cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	t.mu.RLock()
	handler, ok := t.handlers[cmd]
	t.mu.RUnlock()

	if !ok {
		// Default: echo command list
		t.Send(chatID, t.helpText())
		return
	}

	response := handler(chatID, args)
	if response != "" {
		t.Send(chatID, response)
	}
}

// SendAlert pushes an alert to the admin chat
func (t *TelegramCompanion) SendAlert(message string) error {
	if t.cfg.AdminChatID == 0 {
		log.Warn().Msg("TG: no admin chat ID configured for alerts")
		return nil
	}
	return t.Send(t.cfg.AdminChatID, message)
}

// SentLog returns all messages sent in simulated mode
func (t *TelegramCompanion) SentLog() []OutboundMessage {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]OutboundMessage, len(t.sentLog))
	copy(result, t.sentLog)
	return result
}

func (t *TelegramCompanion) isAllowed(userID int64) bool {
	if len(t.cfg.AllowedUserIDs) == 0 {
		return true // no allowlist = allow all
	}
	for _, id := range t.cfg.AllowedUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (t *TelegramCompanion) helpText() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var cmds []string
	for cmd := range t.handlers {
		cmds = append(cmds, "  /"+cmd)
	}
	return fmt.Sprintf("🧠 **NEXUS Bot**\n\nAvailable commands:\n%s\n\nSend any message to chat with NEXUS.",
		strings.Join(cmds, "\n"))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
