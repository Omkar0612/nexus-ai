package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Deployer manages n8n workflow deployment and execution.
type Deployer struct {
	n8nURL     string
	apiKey     string
	httpClient *http.Client
}

// NewDeployer creates a new n8n deployer.
func NewDeployer(n8nURL, apiKey string) *Deployer {
	return &Deployer{
		n8nURL: strings.TrimSuffix(n8nURL, "/"),
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DeployWorkflow creates or updates a workflow in n8n.
func (d *Deployer) DeployWorkflow(ctx context.Context, workflow *Workflow) (string, error) {
	log.Info().Str("workflow", workflow.Name).Msg("Deploying workflow to n8n")

	payload, err := json.Marshal(workflow)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.n8nURL+"/api/v1/workflows", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", d.apiKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to deploy workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("n8n API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	workflowID := result["id"].(string)
	log.Info().Str("workflow_id", workflowID).Msg("✅ Workflow deployed to n8n")
	return workflowID, nil
}

// ActivateWorkflow enables a workflow in n8n.
func (d *Deployer) ActivateWorkflow(ctx context.Context, workflowID string) error {
	log.Info().Str("workflow_id", workflowID).Msg("Activating workflow")

	payload := map[string]bool{"active": true}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "PATCH", d.n8nURL+"/api/v1/workflows/"+workflowID, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", d.apiKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to activate workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to activate workflow: status %d", resp.StatusCode)
	}

	log.Info().Str("workflow_id", workflowID).Msg("✅ Workflow activated")
	return nil
}

// ExecuteWorkflow triggers a workflow execution.
func (d *Deployer) ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (map[string]interface{}, error) {
	log.Info().Str("workflow_id", workflowID).Msg("Executing workflow")

	payload, _ := json.Marshal(input)

	req, err := http.NewRequestWithContext(ctx, "POST", d.n8nURL+"/api/v1/workflows/"+workflowID+"/execute", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", d.apiKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	log.Info().Str("workflow_id", workflowID).Msg("✅ Workflow executed")
	return result, nil
}
