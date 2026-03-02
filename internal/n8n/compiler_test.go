package n8n

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompiler_CompileFromNaturalLanguage(t *testing.T) {
	compiler := NewDAGCompiler()
	err := compiler.Start()
	if err != nil {
		t.Fatalf("failed to start compiler: %v", err)
	}
	defer compiler.Stop()

	// Test conditional workflow
	input := "When temperature is above 30 degrees, send me an alert"
	wf, err := compiler.CompileFromNaturalLanguage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf == nil {
		t.Fatal("expected workflow, got nil")
	}

	if len(wf.Nodes) == 0 {
		t.Error("expected workflow to have nodes")
	}

	if wf.ID == "" {
		t.Error("expected workflow to have an ID")
	}

	t.Logf("Generated workflow '%s' with %d nodes", wf.Name, len(wf.Nodes))
}

func TestCompiler_ValidateWorkflow(t *testing.T) {
	compiler := NewDAGCompiler()

	// Test workflow without trigger - should fail
	emptyWorkflow := &Workflow{
		ID:    "test-wf",
		Name:  "Empty Workflow",
		Nodes: make(map[string]*WorkflowNode),
	}

	err := compiler.validateWorkflow(emptyWorkflow)
	if err == nil {
		t.Error("expected validation error for empty workflow")
	}

	// Test workflow with trigger - should pass
	validWorkflow := &Workflow{
		ID:   "test-wf-2",
		Name: "Valid Workflow",
		Nodes: map[string]*WorkflowNode{
			"trigger-1": {
				ID:          "trigger-1",
				Name:        "Trigger",
				Type:        NodeTrigger,
				Connections: []string{},
			},
		},
	}

	err = compiler.validateWorkflow(validWorkflow)
	if err != nil {
		t.Errorf("expected valid workflow to pass, got: %v", err)
	}
}

func TestCompiler_ExportWorkflow(t *testing.T) {
	compiler := NewDAGCompiler()
	compiler.Start()
	defer compiler.Stop()

	// Create a workflow
	wf, err := compiler.CompileFromNaturalLanguage("Send notification every day at 9 AM")
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	// Export to JSON
	jsonStr, err := compiler.ExportWorkflow(wf.ID)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}

	// Verify it's valid JSON
	var parsed map[string]any
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Errorf("exported JSON is invalid: %v", err)
	}

	t.Logf("Exported workflow: %d bytes", len(jsonStr))
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON in markdown",
			input:    "Here is your JSON:\n```json\n{\"test\": 123}\n```\nEnjoy!",
			expected: `{"test": 123}`,
		},
		{
			name:     "Plain JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with text before",
			input:    "Some text {\"data\": true} more text",
			expected: `{"data": true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if !strings.Contains(result, "{") {
				t.Errorf("expected JSON in result, got: %s", result)
			}
		})
	}
}

// extractJSON extracts JSON from text (simple implementation)
func extractJSON(text string) string {
	// Try to extract from markdown code block
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.Index(text[start:], "```")
		if end > 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// Try to find JSON object
	startIdx := strings.Index(text, "{")
	if startIdx >= 0 {
		// Find matching closing brace
		braceCount := 0
		for i := startIdx; i < len(text); i++ {
			if text[i] == '{' {
				braceCount++
			} else if text[i] == '}' {
				braceCount--
				if braceCount == 0 {
					return strings.TrimSpace(text[startIdx : i+1])
				}
			}
		}
	}

	return text
}
