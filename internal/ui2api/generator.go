package ui2api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// LLMClient represents an interface to Groq/OpenAI to synthesize the code.
type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Generator takes the filtered HAR logs and uses an LLM to write the Go (TinyGo)
// code needed for the Auto-Forge to compile it into a Wasm agent.
type Generator struct {
	llm LLMClient
}

func NewGenerator(llm LLMClient) *Generator {
	return &Generator{llm: llm}
}

// SynthesizeAgent analyzes the undocumented endpoints and writes the Wasm module.
func (g *Generator) SynthesizeAgent(ctx context.Context, appName string, traffic *HARLog) (string, error) {
	log.Info().Str("app", appName).Msg("Synthesizing native tool code from intercepted API traffic...")

	trafficBytes, _ := json.MarshalIndent(traffic.Entries, "", "  ")

	systemPrompt := `You are the NEXUS UI-to-API Engine. 
The user has provided a JSON array of intercepted HTTP traffic from an undocumented application.
Your job is to reverse-engineer these endpoints and write a valid Go script.
This Go script will be compiled to a WebAssembly (Wasi) agent.

RULES:
1. Do not use external libraries. Use only standard library "net/http", "encoding/json", etc.
2. The agent must export a single function: //export run
3. The auth tokens are stored in the NEXUS AES-256 Vault. The agent should retrieve them using the host ABI.
4. Output ONLY the raw Go source code. No markdown formatting, no explanations.`

	userPrompt := fmt.Sprintf(`App Name: %s
Intercepted Traffic:
%s

Write the complete main.go for the new agent.`, appName, string(trafficBytes))

	sourceCode, err := g.llm.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM failed to synthesize agent code: %w", err)
	}

	log.Info().Str("app", appName).Msg("Go code successfully synthesized!")
	return sourceCode, nil
}
