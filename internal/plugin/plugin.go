// Package plugin provides the NEXUS Plugin SDK.
// Plugins are Go shared objects (.so) OR Python scripts placed in ~/.nexus/plugins/.
// Each plugin registers as a Skill: a named function callable by the agent bus.
// Security: plugins run in a subprocess with no vault access unless explicitly granted.
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Skill is a callable unit exposed by a plugin.
type Skill struct {
	Name        string   // e.g. "weather"
	Description string   // shown to LLM for routing decisions
	Version     string   // semver
	Author      string
	Perms       []string // declared permissions: "network", "filesystem", "vault"
}

// Result is the output of a skill invocation.
type Result struct {
	Output  string
	Error   string
	Latency time.Duration
}

// Plugin wraps a loaded skill provider.
type Plugin struct {
	Skill
	path       string // absolute path to the plugin file
	lang       Language
}

// Language is the implementation language of a plugin.
type Language string

const (
	LangPython Language = "python"
	LangBinary Language = "binary" // compiled Go .so or standalone binary
)

// Registry manages all loaded plugins.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin // keyed by Skill.Name
	dir     string             // plugin directory
}

// NewRegistry creates a registry pointing at pluginDir.
// pluginDir is created if it doesn't exist.
func NewRegistry(pluginDir string) (*Registry, error) {
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		return nil, fmt.Errorf("plugin: mkdir %s: %w", pluginDir, err)
	}
	return &Registry{plugins: make(map[string]*Plugin), dir: pluginDir}, nil
}

// Discover scans the plugin directory, reads manifests, and registers plugins.
// A plugin is a directory containing nexus-plugin.json + either plugin.py or plugin (binary).
func (r *Registry) Discover() ([]Skill, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("plugin: read dir: %w", err)
	}
	var found []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skill, lang, err := readManifest(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			continue // silently skip malformed plugins
		}
		p := &Plugin{Skill: skill, path: filepath.Join(r.dir, entry.Name()), lang: lang}
		r.mu.Lock()
		r.plugins[skill.Name] = p
		r.mu.Unlock()
		found = append(found, skill)
	}
	return found, nil
}

// Invoke runs a plugin by name with the given JSON-encoded input.
// The plugin is invoked as a subprocess; stdout is captured as the result.
func (r *Registry) Invoke(ctx context.Context, name, inputJSON string) (*Result, error) {
	r.mu.RLock()
	p, ok := r.plugins[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin: %q not found", name)
	}
	start := time.Now()
	var cmd *exec.Cmd
	switch p.lang {
	case LangPython:
		cmd = exec.CommandContext(ctx, "python3", filepath.Join(p.path, "plugin.py"))
	case LangBinary:
		cmd = exec.CommandContext(ctx, filepath.Join(p.path, "plugin"))
	default:
		return nil, fmt.Errorf("plugin: unknown language %s", p.lang)
	}
	cmd.Stdin = strings.NewReader(inputJSON)
	out, err := cmd.Output()
	latency := time.Since(start)
	if err != nil {
		var exitErr *exec.ExitError
		if ok := isExitError(err, &exitErr); ok {
			return &Result{Error: string(exitErr.Stderr), Latency: latency}, nil
		}
		return nil, fmt.Errorf("plugin[%s]: exec: %w", name, err)
	}
	return &Result{Output: strings.TrimSpace(string(out)), Latency: latency}, nil
}

// List returns all registered skills.
func (r *Registry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skills := make([]Skill, 0, len(r.plugins))
	for _, p := range r.plugins {
		skills = append(skills, p.Skill)
	}
	return skills
}

// Unload removes a plugin from the registry by name.
func (r *Registry) Unload(name string) {
	r.mu.Lock()
	delete(r.plugins, name)
	r.mu.Unlock()
}

// manifest is the JSON structure of nexus-plugin.json.
type manifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Language    string   `json:"language"` // "python" | "binary"
	Perms       []string `json:"permissions"`
}

func readManifest(dir string) (Skill, Language, error) {
	data, err := os.ReadFile(filepath.Join(dir, "nexus-plugin.json"))
	if err != nil {
		return Skill{}, "", err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Skill{}, "", err
	}
	if m.Name == "" {
		return Skill{}, "", fmt.Errorf("plugin: manifest missing name in %s", dir)
	}
	lang := Language(m.Language)
	if lang == "" {
		// Auto-detect
		if _, err := os.Stat(filepath.Join(dir, "plugin.py")); err == nil {
			lang = LangPython
		} else {
			lang = LangBinary
		}
	}
	skill := Skill{
		Name:        m.Name,
		Description: m.Description,
		Version:     m.Version,
		Author:      m.Author,
		Perms:       m.Perms,
	}
	return skill, lang, nil
}

// isExitError is a helper that avoids importing errors package in tests.
func isExitError(err error, out **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*out = e
		return true
	}
	return false
}
