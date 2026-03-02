package agents

/*
PersonaEngine — switch between different AI work modes.

Users need different AI behavior for different contexts:
  nexus persona use work       # formal, full tool access, code-heavy
  nexus persona use creative   # casual, brainstorming mode
  nexus persona use client     # professional, no internal data
  nexus persona use focus      # minimal, 200-word cap
  nexus persona use research   # deep, well-cited, academic

No other open-source agent supports switchable personas.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Persona defines a named AI work mode
type Persona struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	SystemPrompt  string            `json:"system_prompt"`
	LLMPreference string            `json:"llm_preference"`
	ResponseStyle string            `json:"response_style"`
	AllowedAgents []string          `json:"allowed_agents"`
	BlockedAgents []string          `json:"blocked_agents"`
	MemoryScope   string            `json:"memory_scope"`
	Emoji         string            `json:"emoji"`
	IsDefault     bool              `json:"is_default"`
	Custom        map[string]string `json:"custom"`
}

// PersonaEngine manages persona loading and switching
type PersonaEngine struct {
	personas map[string]*Persona
	active   *Persona
	path     string
}

// NewPersonaEngine creates and loads the persona engine
func NewPersonaEngine(configDir string) (*PersonaEngine, error) {
	pe := &PersonaEngine{
		personas: make(map[string]*Persona),
		path:     filepath.Join(configDir, "personas.json"),
	}
	if err := pe.load(); err != nil {
		pe.createDefaults()
	}
	for _, p := range pe.personas {
		if p.IsDefault {
			pe.active = p
			break
		}
	}
	if pe.active == nil {
		pe.active = pe.personas["default"]
	}
	return pe, nil
}

func (pe *PersonaEngine) createDefaults() {
	defaults := []*Persona{
		{Name: "default", Description: "Balanced assistant",
			SystemPrompt:  "You are NEXUS, a helpful AI. Be accurate, concise, and practical.",
			ResponseStyle: "normal", MemoryScope: "full", Emoji: "🧠", IsDefault: true},
		{Name: "work", Description: "Focused professional mode",
			SystemPrompt:  "Be formal, precise, and efficient. Prefer structured outputs. Always include code when relevant.",
			LLMPreference: "anthropic", ResponseStyle: "detailed", MemoryScope: "business", Emoji: "💼"},
		{Name: "creative", Description: "Brainstorming and ideation mode",
			SystemPrompt:  "Think outside the box. Suggest unexpected angles. Challenge assumptions. Be playful but insightful.",
			LLMPreference: "groq", ResponseStyle: "normal", MemoryScope: "full", Emoji: "🎨"},
		{Name: "client", Description: "Client-facing mode — no internal data",
			SystemPrompt:  "Be formal, concise, and positive. Never share internal data or personal notes.",
			ResponseStyle: "brief", MemoryScope: "none",
			BlockedAgents: []string{"system", "file", "memory"}, Emoji: "🤝"},
		{Name: "focus", Description: "Deep work — minimal, no distractions",
			SystemPrompt:  "Give extremely brief, action-oriented responses. No pleasantries. Just answer what was asked.",
			ResponseStyle: "brief", MemoryScope: "full",
			Custom: map[string]string{"max_response_length": "200"}, Emoji: "🎯"},
		{Name: "research", Description: "Deep research — thorough and cited",
			SystemPrompt:  "Be thorough, cite sources, consider multiple perspectives. Quality over speed.",
			LLMPreference: "anthropic", ResponseStyle: "detailed", MemoryScope: "full",
			AllowedAgents: []string{"research", "browser", "vision", "data", "file"}, Emoji: "🔬"},
	}
	for _, p := range defaults {
		pe.personas[p.Name] = p
	}
	_ = pe.save()
}

// Switch changes the active persona
func (pe *PersonaEngine) Switch(name string) error {
	p, ok := pe.personas[name]
	if !ok {
		return fmt.Errorf("persona '%s' not found. Available: %v", name, pe.ListNames())
	}
	pe.active = p
	return nil
}

// Active returns the current persona
func (pe *PersonaEngine) Active() *Persona { return pe.active }

// ListNames returns all persona names
func (pe *PersonaEngine) ListNames() []string {
	names := make([]string, 0, len(pe.personas))
	for name := range pe.personas {
		names = append(names, name)
	}
	return names
}

func (pe *PersonaEngine) load() error {
	data, err := os.ReadFile(pe.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &pe.personas)
}

func (pe *PersonaEngine) save() error {
	data, err := json.MarshalIndent(pe.personas, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(pe.path), 0700)
	return os.WriteFile(pe.path, data, 0600)
}
