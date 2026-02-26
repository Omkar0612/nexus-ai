package mobile

import (
	"testing"
)

func simBot() *TelegramCompanion {
	return New(BotConfig{Simulated: true, AdminChatID: 12345})
}

func TestTelegramSendSimulated(t *testing.T) {
	bot := simBot()
	if err := bot.Send(12345, "Hello from NEXUS"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(bot.SentLog()) == 0 {
		t.Error("expected message in sent log")
	}
}

func TestTelegramCommandRouting(t *testing.T) {
	bot := simBot()
	bot.RegisterCommand("status", func(chatID int64, args string) string {
		return "🟢 NEXUS is running"
	})

	bot.HandleUpdate(TelegramMessage{
		Message: struct {
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
		}{
			From: struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			}{ID: 0}, // allowed (no allowlist)
			Chat: struct {
				ID int64 `json:"id"`
			}{ID: 12345},
			Text: "/status",
		},
	})

	log := bot.SentLog()
	if len(log) == 0 {
		t.Error("expected a response to /status command")
	}
}

func TestTelegramUnauthorisedUser(t *testing.T) {
	bot := New(BotConfig{
		Simulated:      true,
		AllowedUserIDs: []int64{999},
		AdminChatID:    12345,
	})
	bot.HandleUpdate(TelegramMessage{
		Message: struct {
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
		}{
			From: struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			}{ID: 1234}, // not in allowlist
			Chat: struct {
				ID int64 `json:"id"`
			}{ID: 1234},
			Text: "/status",
		},
	})
	log := bot.SentLog()
	if len(log) == 0 || !containsStr(log[0].Text, "Unauthorised") {
		t.Error("expected unauthorised message")
	}
}

func TestTelegramSendAlert(t *testing.T) {
	bot := simBot()
	if err := bot.SendAlert("🚨 Budget breached!"); err != nil {
		t.Fatalf("SendAlert: %v", err)
	}
	if len(bot.SentLog()) == 0 {
		t.Error("expected alert in sent log")
	}
}

func TestTelegramUnknownCommand(t *testing.T) {
	bot := simBot()
	bot.HandleUpdate(TelegramMessage{
		Message: struct {
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
		}{
			From: struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			}{ID: 0},
			Chat: struct {
				ID int64 `json:"id"`
			}{ID: 99},
			Text: "/unknowncmd",
		},
	})
	if len(bot.SentLog()) == 0 {
		t.Error("expected help text for unknown command")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && (s[:len(sub)] == sub || containsStr(s[1:], sub))))
}
