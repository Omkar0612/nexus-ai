package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Deployer handles pushing compiled workflows directly to a live n8n instance.
type Deployer struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// NewDeployer initializes a client to communicate with the n8n REST API.
func NewDeployer(baseURL, apiKey string) *Deployer {
	return &Deployer{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Deploy pushes a Workflow struct to the configured n8n instance.
func (d *Deployer) Deploy(ctx context.Context, wf *Workflow) error {
	log.Info().Str("target", d.BaseURL).Msg("🚀 Deploying workflow to n8n instance...")

	payload, err := json.Marshal(wf)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/workflows", d.BaseURL), bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", d.APIKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("n8n deployment request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("n8n API returned error status: %d", resp.StatusCode)
	}

	log.Info().Str("workflow", wf.Name).Msg("✅ Workflow successfully deployed to n8n!")
	return nil
}
