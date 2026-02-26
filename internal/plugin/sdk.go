// Package plugin provides the NEXUS plugin SDK.
// Third-party skills implement the Plugin interface and register via Register().
package plugin

import "context"

// Input is passed to a skill's Execute method.
type Input struct {
	Command string            // natural language command
	Args    map[string]string // parsed key-value args
	Context context.Context
}

// Output is returned by a skill.
type Output struct {
	Text  string // human-readable result
	Data  any    // structured data (JSON-serialisable)
	Error error
}

// Plugin is the interface every NEXUS skill must implement.
type Plugin interface {
	// Name returns the unique skill name (e.g. "weather", "translate").
	Name() string
	// Description is shown in `nexus skills list`.
	Description() string
	// Execute runs the skill.
	Execute(input Input) Output
}

// SkillFunc wraps a plain function as a Plugin.
type SkillFunc struct {
	name, desc string
	fn         func(Input) Output
}

// NewSkill creates a Plugin from a function.
func NewSkill(name, description string, fn func(Input) Output) Plugin {
	return &SkillFunc{name: name, desc: description, fn: fn}
}

func (s *SkillFunc) Name() string        { return s.name }
func (s *SkillFunc) Description() string { return s.desc }
func (s *SkillFunc) Execute(in Input) Output { return s.fn(in) }
