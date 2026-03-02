package ui2api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// NetworkRequest represents a captured HTTP request from the browser's Network tab.
type NetworkRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// NetworkResponse represents the payload returned by the server.
type NetworkResponse struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// HARLog is a simplified representation of the browser's HTTP Archive.
type HARLog struct {
	Entries []HAREntry `json:"entries"`
}

// HAREntry represents a single network call pair.
type HAREntry struct {
	Request  NetworkRequest  `json:"request"`
	Response NetworkResponse `json:"response"`
}

// Vault abstractly represents the NEXUS AES-256 Vault where we securely store
// extracted session cookies and bearer tokens.
type Vault interface {
	StoreSecret(ctx context.Context, key, value string) error
}

// Interceptor processes raw browser traffic to extract undocumented APIs.
type Interceptor struct {
	vault Vault
}

func NewInterceptor(vault Vault) *Interceptor {
	return &Interceptor{vault: vault}
}

// ProcessTraffic analyzes the headless browser's network log to identify API routes,
// strips out static assets, and extracts the authentication tokens.
func (i *Interceptor) ProcessTraffic(ctx context.Context, appName string, rawHAR []byte) (*HARLog, error) {
	log.Info().Str("app", appName).Msg("Processing intercepted browser traffic...")

	var fullLog HARLog
	if err := json.Unmarshal(rawHAR, &fullLog); err != nil {
		return nil, fmt.Errorf("failed to parse HAR data: %w", err)
	}

	var filteredLog HARLog

	for _, entry := range fullLog.Entries {
		url := entry.Request.URL
		// Filter out static assets and telemetry
		if strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".css") || strings.HasSuffix(url, ".js") {
			continue
		}
		if strings.Contains(url, "google-analytics.com") || strings.Contains(url, "telemetry") {
			continue
		}

		// Look for Bearer tokens or Session Cookies
		authHeader := entry.Request.Headers["Authorization"]
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			secretKey := fmt.Sprintf("%s_API_TOKEN", strings.ToUpper(appName))

			log.Info().Str("app", appName).Msg("Discovered undocumented Bearer Token! Securing in Vault.")
			if err := i.vault.StoreSecret(ctx, secretKey, token); err != nil {
				return nil, fmt.Errorf("failed to secure token in vault: %w", err)
			}

			// Redact the token from the log we send to the LLM to prevent leaking secrets to Groq/OpenAI
			entry.Request.Headers["Authorization"] = "Bearer [REDACTED_SECURED_IN_VAULT]"
		}

		cookieHeader := entry.Request.Headers["Cookie"]
		if cookieHeader != "" {
			secretKey := fmt.Sprintf("%s_SESSION_COOKIE", strings.ToUpper(appName))
			log.Info().Str("app", appName).Msg("Discovered Session Cookie! Securing in Vault.")

			if err := i.vault.StoreSecret(ctx, secretKey, cookieHeader); err != nil {
				return nil, fmt.Errorf("failed to secure cookie in vault: %w", err)
			}

			entry.Request.Headers["Cookie"] = "[REDACTED_SECURED_IN_VAULT]"
		}

		// Only keep requests that return JSON (the actual hidden APIs)
		if strings.Contains(entry.Response.Body, "{") || strings.Contains(entry.Response.Body, "[") {
			filteredLog.Entries = append(filteredLog.Entries, entry)
		}
	}

	log.Info().Int("endpoints_discovered", len(filteredLog.Entries)).Msg("Traffic processing complete.")
	return &filteredLog, nil
}
