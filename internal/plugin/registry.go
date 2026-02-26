// Package plugin provides the NEXUS plugin registry.
package plugin

import (
	"fmt"
	"sort"
	"strings"
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

// Register adds a plugin to the registry.
// Names are normalised to lowercase so "Echo" and "echo" are the same skill.
// Returns an error on duplicate registration instead of silently overwriting.
func (r *Registry) Register(p Plugin) error {
	key := strings.ToLower(p.Name())
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[key]; exists {
		return fmt.Errorf("plugin: %q already registered", key)
	}
	r.plugins[key] = p
	return nil
}

// Get retrieves a plugin by name (case-insensitive).
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[strings.ToLower(name)]
	return p, ok
}

// List returns all registered plugin names and descriptions, sorted alphabetically.
// Deterministic output makes `nexus skills list` stable across runs.
func (r *Registry) List() []string {
	r.mu.RLock()
	keys := make([]string, 0, len(r.plugins))
	for k := range r.plugins {
		keys = append(keys, k)
	}
	r.mu.RUnlock()

	sort.Strings(keys)

	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		p := r.plugins[k]
		out = append(out, fmt.Sprintf("%-20s %s", p.Name(), p.Description()))
	}
	return out
}

// Execute runs a named plugin with the given input.
func (r *Registry) Execute(name string, input Input) (Output, error) {
	p, ok := r.Get(name) // Get already holds RLock for the lookup
	if !ok {
		return Output{}, fmt.Errorf("plugin: %q not found", strings.ToLower(name))
	}
	return p.Execute(input), nil
}

// Len returns the number of registered plugins.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}
