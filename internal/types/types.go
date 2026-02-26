package types

import "time"

// Message is a single conversation turn
type Message struct {
	Role      string    `json:"role"`      // user, assistant, system
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	AgentUsed string    `json:"agent_used,omitempty"`
	LatencyMs int64     `json:"latency_ms,omitempty"`
}

// AgentResult is the output of any NEXUS agent execution
type AgentResult struct {
	Content   string            `json:"content"`
	Agent     string            `json:"agent"`
	Model     string            `json:"model"`
	LatencyMs int64             `json:"latency_ms"`
	TokensIn  int               `json:"tokens_in"`
	TokensOut int               `json:"tokens_out"`
	Meta      map[string]string `json:"meta,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Config is the top-level NEXUS configuration
type Config struct {
	UserID   string         `toml:"user_id"   mapstructure:"user_id"`
	DataDir  string         `toml:"data_dir"  mapstructure:"data_dir"`
	LogLevel string         `toml:"log_level" mapstructure:"log_level"`
	LLM      LLMConfig      `toml:"llm"       mapstructure:"llm"`
	Memory   MemoryConfig   `toml:"memory"    mapstructure:"memory"`
	Gateway  GatewayConfig  `toml:"gateway"   mapstructure:"gateway"`
	Agents   AgentsConfig   `toml:"agents"    mapstructure:"agents"`
}

// LLMConfig holds LLM provider settings
type LLMConfig struct {
	Provider   string `toml:"provider"   mapstructure:"provider"`
	Model      string `toml:"model"      mapstructure:"model"`
	APIKey     string `toml:"api_key"    mapstructure:"api_key"`
	BaseURL    string `toml:"base_url"   mapstructure:"base_url"`
	MaxTokens  int    `toml:"max_tokens" mapstructure:"max_tokens"`
	TimeoutSec int    `toml:"timeout_sec" mapstructure:"timeout_sec"`
	Fallback   string `toml:"fallback"   mapstructure:"fallback"`
}

// MemoryConfig holds memory storage settings
type MemoryConfig struct {
	MaxEpisodic int    `toml:"max_episodic" mapstructure:"max_episodic"`
	MaxSemantic int    `toml:"max_semantic" mapstructure:"max_semantic"`
	Encrypt     bool   `toml:"encrypt"     mapstructure:"encrypt"`
	Path        string `toml:"path"        mapstructure:"path"`
}

// GatewayConfig holds API gateway settings
type GatewayConfig struct {
	Port    int    `toml:"port"    mapstructure:"port"`
	Host    string `toml:"host"    mapstructure:"host"`
	APIKey  string `toml:"api_key" mapstructure:"api_key"`
}

// AgentsConfig holds agent-level settings
type AgentsConfig struct {
	TimeoutSec       int  `toml:"timeout_sec"        mapstructure:"timeout_sec"`
	HumanInLoopRisk  bool `toml:"human_in_loop_high_risk" mapstructure:"human_in_loop_high_risk"`
}
