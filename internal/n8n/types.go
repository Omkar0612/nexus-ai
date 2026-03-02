package n8n

// Workflow represents a standard n8n workflow JSON structure.
type Workflow struct {
	Name        string                               `json:"name"`
	Nodes       []Node                               `json:"nodes"`
	Connections map[string]map[string][]ConnectionTarget `json:"connections"`
	Settings    map[string]interface{}               `json:"settings,omitempty"`
}

// Node represents a single step or integration in an n8n workflow.
type Node struct {
	Parameters  map[string]interface{} `json:"parameters"`
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	TypeVersion float64                `json:"typeVersion"`
	Position    []float64              `json:"position"`
}

// ConnectionTarget represents the destination of an n8n node connection.
type ConnectionTarget struct {
	Node  string `json:"node"`
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// LLMClient represents an interface for text generation.
type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
