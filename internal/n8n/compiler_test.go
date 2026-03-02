package n8n

import (
	"context"
	"testing"
)

type mockLLM struct {
	jsonResponse string
}

func (m *mockLLM) Generate(ctx context.Context, sys, prompt string) (string, error) {
	return m.jsonResponse, nil
}

func TestCompiler_Compile(t *testing.T) {
	mockResponse := `
	{
		"name": "ERP Sync to Meta Ads",
		"nodes": [
			{
				"parameters": {},
				"name": "Webhook",
				"type": "n8n-nodes-base.webhook",
				"typeVersion": 1,
				"position": [250, 300]
			},
			{
				"parameters": {"url": "https://graph.facebook.com"},
				"name": "HTTP Request",
				"type": "n8n-nodes-base.httpRequest",
				"typeVersion": 1,
				"position": [450, 300]
			}
		],
		"connections": {
			"Webhook": {
				"main": [
					[
						{"node": "HTTP Request", "type": "main", "index": 0}
					]
				]
			}
		}
	}
	`

	llm := &mockLLM{jsonResponse: mockResponse}
	compiler := NewCompiler(llm)

	wf, err := compiler.Compile(context.Background(), "Create a workflow that triggers on a webhook and hits a URL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Name != "ERP Sync to Meta Ads" {
		t.Errorf("expected workflow name 'ERP Sync to Meta Ads', got '%s'", wf.Name)
	}

	if len(wf.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(wf.Nodes))
	}

	if wf.Nodes[1].Name != "HTTP Request" {
		t.Errorf("expected second node to be 'HTTP Request', got '%s'", wf.Nodes[1].Name)
	}
}

func TestExtractJSON(t *testing.T) {
	input := "Here is your JSON:\n```json\n{\"test\": 123}\n```\nEnjoy!"
	expected := `{"test": 123}`

	result := extractJSON(input)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
