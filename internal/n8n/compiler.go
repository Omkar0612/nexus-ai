package n8n

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Compiler translates natural language prompts into a valid n8n Directed Acyclic Graph (DAG) JSON.
type Compiler struct {
	llm LLMClient
}

// NewCompiler initializes the n8n DAG Compiler.
func NewCompiler(llm LLMClient) *Compiler {
	return &Compiler{llm: llm}
}

// Compile generates the n8n workflow from a natural language request.
func (c *Compiler) Compile(ctx context.Context, description string) (*Workflow, error) {
	log.Info().Msg("⚙️ Compiling natural language into n8n Workflow DAG...")

	systemPrompt := `You are an expert n8n workflow architect.
Your job is to translate a user's natural language request into a valid n8n Workflow JSON object.
RULES:
1. Output ONLY valid JSON. No markdown, no explanations.
2. The JSON must contain "name", "nodes" (array), and "connections" (object).
3. Use standard n8n node types like "n8n-nodes-base.webhook", "n8n-nodes-base.httpRequest", "n8n-nodes-base.if".
4. Position nodes logically (e.g., [250, 300], [450, 300]).
5. Ensure connections map properly between node names.`

	userPrompt := fmt.Sprintf("Build an n8n workflow for the following requirement: %s", description)

	rawOutput, err := c.llm.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Clean up output in case the LLM ignored rules and wrapped in markdown
	jsonPayload := extractJSON(rawOutput)

	var wf Workflow
	if err := json.Unmarshal([]byte(jsonPayload), &wf); err != nil {
		log.Error().Str("payload", jsonPayload).Err(err).Msg("Failed to parse LLM output into n8n Workflow")
		return nil, fmt.Errorf("invalid n8n JSON generated: %w", err)
	}

	log.Info().
		Str("workflow_name", wf.Name).
		Int("nodes_count", len(wf.Nodes)).
		Msg("✅ Successfully compiled n8n Workflow DAG!")

	return &wf, nil
}

// extractJSON isolates a JSON block from text if wrapped in markdown code fences.
func extractJSON(input string) string {
	re := regexp.MustCompile("(?s)```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(input)
}
