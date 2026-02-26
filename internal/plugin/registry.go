package plugin

import (
	"fmt"
	"sync"
)

// Registry holds all registered plugins.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]Plugin)}
}

// Register adds a plugin to the registry. Panics on duplicate names.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin: %q already registered", p.Name())
	}
	r.plugins[p.Name()] = p
	return nil
}

// Get retrieves a plugin by name.
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

// List returns all registered plugin names and descriptions.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.plugins))
	for _, p := range r.plugins {
		out = append(out, fmt.Sprintf("%-20s %s", p.Name(), p.Description()))
	}
	return out
}

// Execute runs a named plugin with the given input.
func (r *Registry) Execute(name string, input Input) (Output, error) {
	p, ok := r.Get(name)
	if !ok {
		return Output{}, fmt.Errorf("plugin: %q not found", name)
	}
	return p.Execute(input), nil
}
