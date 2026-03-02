package ui2api

import (
	"context"
	"strings"
	"testing"
)

type mockVault struct {
	store map[string]string
}

func (m *mockVault) StoreSecret(ctx context.Context, key, value string) error {
	m.store[key] = value
	return nil
}

func TestInterceptor_ProcessTraffic(t *testing.T) {
	rawHAR := []byte(`{
		"entries": [
			{
				"request": {
					"url": "https://erp.company.com/api/v1/invoices",
					"method": "GET",
					"headers": {
						"Authorization": "Bearer super_secret_jwt_123"
					}
				},
				"response": {
					"status": 200,
					"body": "{\"data\": [\"inv_1\", \"inv_2\"]}"
				}
			},
			{
				"request": {
					"url": "https://erp.company.com/logo.png",
					"method": "GET"
				},
				"response": {
					"status": 200,
					"body": "image_bytes"
				}
			}
		]
	}`)

	vault := &mockVault{store: make(map[string]string)}
	interceptor := NewInterceptor(vault)

	filtered, err := interceptor.ProcessTraffic(context.Background(), "FirstBit", rawHAR)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Should filter out the .png request
	if len(filtered.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(filtered.Entries))
	}

	// 2. Should extract and redact the Bearer token
	redactedHeader := filtered.Entries[0].Request.Headers["Authorization"]
	if !strings.Contains(redactedHeader, "REDACTED") {
		t.Errorf("expected header to be redacted, got %s", redactedHeader)
	}

	// 3. Should store the token in the vault safely
	if vault.store["FIRSTBIT_API_TOKEN"] != "super_secret_jwt_123" {
		t.Errorf("expected token to be stored in vault, got %s", vault.store["FIRSTBIT_API_TOKEN"])
	}
}
