package n8n

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// NodeType represents different n8n node types
type NodeType string

const (
	NodeTrigger    NodeType = "trigger"
	NodeAction     NodeType = "action"
	NodeCondition  NodeType = "condition"
	NodeLoop       NodeType = "loop"
	NodeTransform  NodeType = "transform"
	NodeMerge      NodeType = "merge"
)

// WorkflowNode represents a single node in the DAG
type WorkflowNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        NodeType               `json:"type"`
	Operation   string                 `json:"operation"`
	Parameters  map[string]any         `json:"parameters"`
	Position    [2]int                 `json:"position"`
	Connections []string               `json:"connections"`
}

// Workflow represents a complete n8n workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       map[string]*WorkflowNode `json:"nodes"`
	CreatedAt   time.Time              `json:"created_at"`
	Active      bool                   `json:"active"`
}

// ExecutionContext holds workflow execution state
type ExecutionContext struct {
	WorkflowID   string                 `json:"workflow_id"`
	ExecutionID  string                 `json:"execution_id"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Status       string                 `json:"status"`
	Data         map[string]any         `json:"data"`
	Error        string                 `json:"error,omitempty"`
	NodeResults  map[string]any         `json:"node_results"`
}

// DAGCompiler converts natural language to n8n workflows
type DAGCompiler struct {
	mu           sync.RWMutex
	workflows    map[string]*Workflow
	executions   map[string]*ExecutionContext
	ctx          context.Context
	cancel       context.CancelFunc
	nodeRegistry map[string]NodeHandler
}

// NodeHandler executes a specific node type
type NodeHandler func(ctx *ExecutionContext, node *WorkflowNode) (any, error)

// NewDAGCompiler creates a new workflow compiler
func NewDAGCompiler() *DAGCompiler {
	ctx, cancel := context.WithCancel(context.Background())

	compiler := &DAGCompiler{
		workflows:    make(map[string]*Workflow),
		executions:   make(map[string]*ExecutionContext),
		ctx:          ctx,
		cancel:       cancel,
		nodeRegistry: make(map[string]NodeHandler),
	}

	// Register default node handlers
	compiler.registerDefaultHandlers()

	return compiler
}

// Start initializes the compiler
func (dc *DAGCompiler) Start() error {
	log.Info().Msg("Starting n8n DAG compiler")
	return nil
}

// Stop gracefully shuts down the compiler
func (dc *DAGCompiler) Stop() error {
	log.Info().Msg("Stopping n8n DAG compiler")
	dc.cancel()
	return nil
}

// CompileFromNaturalLanguage converts text to workflow
func (dc *DAGCompiler) CompileFromNaturalLanguage(input string) (*Workflow, error) {
	log.Info().Str("input", input).Msg("Compiling natural language to workflow")

	// Parse natural language
	tokens := dc.tokenize(input)

	// Extract intent and parameters
	intent, params := dc.extractIntent(tokens)

	// Build workflow based on intent
	workflow := dc.buildWorkflow(intent, params)

	// Validate workflow
	if err := dc.validateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Store workflow
	dc.mu.Lock()
	dc.workflows[workflow.ID] = workflow
	dc.mu.Unlock()

	log.Info().
		Str("workflow_id", workflow.ID).
		Str("name", workflow.Name).
		Int("nodes", len(workflow.Nodes)).
		Msg("Workflow compiled successfully")

	return workflow, nil
}

// tokenize splits input into semantic tokens
func (dc *DAGCompiler) tokenize(input string) []string {
	// Simple tokenization - can be enhanced with NLP
	input = strings.ToLower(input)
	tokens := strings.Fields(input)
	return tokens
}

// extractIntent determines the workflow type and parameters
func (dc *DAGCompiler) extractIntent(tokens []string) (string, map[string]any) {
	params := make(map[string]any)

	// Detect common workflow patterns
	if containsAny(tokens, []string{"when", "if", "whenever"}) {
		return "conditional", params
	}

	if containsAny(tokens, []string{"every", "daily", "hourly", "schedule"}) {
		return "scheduled", params
	}

	if containsAny(tokens, []string{"send", "email", "notify", "message"}) {
		return "notification", params
	}

	if containsAny(tokens, []string{"fetch", "get", "retrieve", "download"}) {
		return "data_fetch", params
	}

	if containsAny(tokens, []string{"process", "transform", "convert", "parse"}) {
		return "data_transform", params
	}

	return "generic", params
}

// buildWorkflow creates a workflow from intent
func (dc *DAGCompiler) buildWorkflow(intent string, params map[string]any) *Workflow {
	workflowID := fmt.Sprintf("wf-%d", time.Now().UnixNano())

	workflow := &Workflow{
		ID:          workflowID,
		Name:        fmt.Sprintf("Workflow-%s", intent),
		Description: fmt.Sprintf("Auto-generated %s workflow", intent),
		Nodes:       make(map[string]*WorkflowNode),
		CreatedAt:   time.Now(),
		Active:      true,
	}

	// Build nodes based on intent
	switch intent {
	case "conditional":
		dc.buildConditionalWorkflow(workflow, params)
	case "scheduled":
		dc.buildScheduledWorkflow(workflow, params)
	case "notification":
		dc.buildNotificationWorkflow(workflow, params)
	case "data_fetch":
		dc.buildDataFetchWorkflow(workflow, params)
	case "data_transform":
		dc.buildDataTransformWorkflow(workflow, params)
	default:
		dc.buildGenericWorkflow(workflow, params)
	}

	return workflow
}

// buildConditionalWorkflow creates if-then-else workflow
func (dc *DAGCompiler) buildConditionalWorkflow(wf *Workflow, params map[string]any) {
	// Trigger node
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Manual Trigger",
		Type:        NodeTrigger,
		Operation:   "manual",
		Parameters:  params,
		Position:    [2]int{100, 100},
		Connections: []string{"condition-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	// Condition node
	condition := &WorkflowNode{
		ID:          "condition-1",
		Name:        "IF Condition",
		Type:        NodeCondition,
		Operation:   "evaluate",
		Parameters:  map[string]any{"expression": "{{ $json.value > 0 }}"},
		Position:    [2]int{300, 100},
		Connections: []string{"action-true", "action-false"},
	}
	wf.Nodes[condition.ID] = condition

	// True branch
	actionTrue := &WorkflowNode{
		ID:          "action-true",
		Name:        "Execute True",
		Type:        NodeAction,
		Operation:   "execute",
		Parameters:  map[string]any{"branch": "true"},
		Position:    [2]int{500, 50},
		Connections: []string{},
	}
	wf.Nodes[actionTrue.ID] = actionTrue

	// False branch
	actionFalse := &WorkflowNode{
		ID:          "action-false",
		Name:        "Execute False",
		Type:        NodeAction,
		Operation:   "execute",
		Parameters:  map[string]any{"branch": "false"},
		Position:    [2]int{500, 150},
		Connections: []string{},
	}
	wf.Nodes[actionFalse.ID] = actionFalse
}

// buildScheduledWorkflow creates time-based workflow
func (dc *DAGCompiler) buildScheduledWorkflow(wf *Workflow, params map[string]any) {
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Schedule Trigger",
		Type:        NodeTrigger,
		Operation:   "cron",
		Parameters:  map[string]any{"cron": "0 9 * * *"}, // Daily at 9 AM
		Position:    [2]int{100, 100},
		Connections: []string{"action-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	action := &WorkflowNode{
		ID:          "action-1",
		Name:        "Execute Task",
		Type:        NodeAction,
		Operation:   "execute",
		Parameters:  params,
		Position:    [2]int{300, 100},
		Connections: []string{},
	}
	wf.Nodes[action.ID] = action
}

// buildNotificationWorkflow creates messaging workflow
func (dc *DAGCompiler) buildNotificationWorkflow(wf *Workflow, params map[string]any) {
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Webhook Trigger",
		Type:        NodeTrigger,
		Operation:   "webhook",
		Parameters:  params,
		Position:    [2]int{100, 100},
		Connections: []string{"send-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	send := &WorkflowNode{
		ID:          "send-1",
		Name:        "Send Notification",
		Type:        NodeAction,
		Operation:   "send_message",
		Parameters:  map[string]any{"channel": "email"},
		Position:    [2]int{300, 100},
		Connections: []string{},
	}
	wf.Nodes[send.ID] = send
}

// buildDataFetchWorkflow creates data retrieval workflow
func (dc *DAGCompiler) buildDataFetchWorkflow(wf *Workflow, params map[string]any) {
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Manual Trigger",
		Type:        NodeTrigger,
		Operation:   "manual",
		Parameters:  params,
		Position:    [2]int{100, 100},
		Connections: []string{"fetch-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	fetch := &WorkflowNode{
		ID:          "fetch-1",
		Name:        "Fetch Data",
		Type:        NodeAction,
		Operation:   "http_request",
		Parameters:  map[string]any{"method": "GET"},
		Position:    [2]int{300, 100},
		Connections: []string{"transform-1"},
	}
	wf.Nodes[fetch.ID] = fetch

	transform := &WorkflowNode{
		ID:          "transform-1",
		Name:        "Transform Data",
		Type:        NodeTransform,
		Operation:   "json_parse",
		Parameters:  map[string]any{},
		Position:    [2]int{500, 100},
		Connections: []string{},
	}
	wf.Nodes[transform.ID] = transform
}

// buildDataTransformWorkflow creates data processing workflow
func (dc *DAGCompiler) buildDataTransformWorkflow(wf *Workflow, params map[string]any) {
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Data Input",
		Type:        NodeTrigger,
		Operation:   "manual",
		Parameters:  params,
		Position:    [2]int{100, 100},
		Connections: []string{"transform-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	transform := &WorkflowNode{
		ID:          "transform-1",
		Name:        "Process Data",
		Type:        NodeTransform,
		Operation:   "transform",
		Parameters:  params,
		Position:    [2]int{300, 100},
		Connections: []string{"output-1"},
	}
	wf.Nodes[transform.ID] = transform

	output := &WorkflowNode{
		ID:          "output-1",
		Name:        "Output Result",
		Type:        NodeAction,
		Operation:   "output",
		Parameters:  map[string]any{},
		Position:    [2]int{500, 100},
		Connections: []string{},
	}
	wf.Nodes[output.ID] = output
}

// buildGenericWorkflow creates default workflow
func (dc *DAGCompiler) buildGenericWorkflow(wf *Workflow, params map[string]any) {
	trigger := &WorkflowNode{
		ID:          "trigger-1",
		Name:        "Start",
		Type:        NodeTrigger,
		Operation:   "manual",
		Parameters:  params,
		Position:    [2]int{100, 100},
		Connections: []string{"action-1"},
	}
	wf.Nodes[trigger.ID] = trigger

	action := &WorkflowNode{
		ID:          "action-1",
		Name:        "Execute",
		Type:        NodeAction,
		Operation:   "execute",
		Parameters:  params,
		Position:    [2]int{300, 100},
		Connections: []string{},
	}
	wf.Nodes[action.ID] = action
}

// validateWorkflow checks workflow structure
func (dc *DAGCompiler) validateWorkflow(wf *Workflow) error {
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow has no nodes")
	}

	// Check for trigger node
	hasTrigger := false
	for _, node := range wf.Nodes {
		if node.Type == NodeTrigger {
			hasTrigger = true
			break
		}
	}

	if !hasTrigger {
		return fmt.Errorf("workflow must have at least one trigger node")
	}

	// Check for circular dependencies
	if dc.hasCircularDependency(wf) {
		return fmt.Errorf("workflow has circular dependencies")
	}

	return nil
}

// hasCircularDependency detects cycles in workflow
func (dc *DAGCompiler) hasCircularDependency(wf *Workflow) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(nodeID string) bool
	dfs = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		node, exists := wf.Nodes[nodeID]
		if !exists {
			return false
		}

		for _, connID := range node.Connections {
			if !visited[connID] {
				if dfs(connID) {
					return true
				}
			} else if recStack[connID] {
				return true // Cycle detected
			}
		}

		recStack[nodeID] = false
		return false
	}

	for nodeID := range wf.Nodes {
		if !visited[nodeID] {
			if dfs(nodeID) {
				return true
			}
		}
	}

	return false
}

// ExecuteWorkflow runs a compiled workflow
func (dc *DAGCompiler) ExecuteWorkflow(workflowID string, input map[string]any) (*ExecutionContext, error) {
	dc.mu.RLock()
	workflow, exists := dc.workflows[workflowID]
	dc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	execCtx := &ExecutionContext{
		WorkflowID:  workflowID,
		ExecutionID: fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		StartTime:   time.Now(),
		Status:      "running",
		Data:        input,
		NodeResults: make(map[string]any),
	}

	dc.mu.Lock()
	dc.executions[execCtx.ExecutionID] = execCtx
	dc.mu.Unlock()

	log.Info().
		Str("workflow_id", workflowID).
		Str("execution_id", execCtx.ExecutionID).
		Msg("Starting workflow execution")

	// Execute workflow
	if err := dc.executeDAG(execCtx, workflow); err != nil {
		execCtx.Status = "failed"
		execCtx.Error = err.Error()
		execCtx.EndTime = time.Now()
		return execCtx, err
	}

	execCtx.Status = "completed"
	execCtx.EndTime = time.Now()

	log.Info().
		Str("execution_id", execCtx.ExecutionID).
		Dur("duration", execCtx.EndTime.Sub(execCtx.StartTime)).
		Msg("Workflow execution completed")

	return execCtx, nil
}

// executeDAG runs nodes in topological order
func (dc *DAGCompiler) executeDAG(execCtx *ExecutionContext, workflow *Workflow) error {
	// Find trigger nodes
	triggerNodes := make([]*WorkflowNode, 0)
	for _, node := range workflow.Nodes {
		if node.Type == NodeTrigger {
			triggerNodes = append(triggerNodes, node)
		}
	}

	if len(triggerNodes) == 0 {
		return fmt.Errorf("no trigger nodes found")
	}

	// Execute from each trigger
	for _, trigger := range triggerNodes {
		if err := dc.executeNode(execCtx, workflow, trigger); err != nil {
			return err
		}
	}

	return nil
}

// executeNode runs a single node and its connections
func (dc *DAGCompiler) executeNode(execCtx *ExecutionContext, workflow *Workflow, node *WorkflowNode) error {
	log.Debug().
		Str("node_id", node.ID).
		Str("node_name", node.Name).
		Msg("Executing node")

	// Get handler for node operation
	handler := dc.getNodeHandler(string(node.Type), node.Operation)
	if handler == nil {
		return fmt.Errorf("no handler for node type %s operation %s", node.Type, node.Operation)
	}

	// Execute node
	result, err := handler(execCtx, node)
	if err != nil {
		return fmt.Errorf("node %s execution failed: %w", node.ID, err)
	}

	// Store result
	execCtx.NodeResults[node.ID] = result

	// Execute connected nodes
	for _, connID := range node.Connections {
		connNode, exists := workflow.Nodes[connID]
		if !exists {
			continue
		}

		if err := dc.executeNode(execCtx, workflow, connNode); err != nil {
			return err
		}
	}

	return nil
}

// registerDefaultHandlers sets up built-in node handlers
func (dc *DAGCompiler) registerDefaultHandlers() {
	dc.RegisterNodeHandler("trigger", "manual", func(ctx *ExecutionContext, node *WorkflowNode) (any, error) {
		return ctx.Data, nil
	})

	dc.RegisterNodeHandler("action", "execute", func(ctx *ExecutionContext, node *WorkflowNode) (any, error) {
		return map[string]any{"status": "executed", "params": node.Parameters}, nil
	})

	dc.RegisterNodeHandler("condition", "evaluate", func(ctx *ExecutionContext, node *WorkflowNode) (any, error) {
		// Simple condition evaluation
		return map[string]any{"result": true}, nil
	})

	dc.RegisterNodeHandler("transform", "transform", func(ctx *ExecutionContext, node *WorkflowNode) (any, error) {
		return map[string]any{"transformed": true, "data": ctx.Data}, nil
	})
}

// RegisterNodeHandler adds a custom node handler
func (dc *DAGCompiler) RegisterNodeHandler(nodeType, operation string, handler NodeHandler) {
	key := fmt.Sprintf("%s:%s", nodeType, operation)
	dc.nodeRegistry[key] = handler
}

// getNodeHandler retrieves a node handler
func (dc *DAGCompiler) getNodeHandler(nodeType, operation string) NodeHandler {
	key := fmt.Sprintf("%s:%s", nodeType, operation)
	return dc.nodeRegistry[key]
}

// GetWorkflow retrieves a workflow by ID
func (dc *DAGCompiler) GetWorkflow(workflowID string) (*Workflow, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	workflow, exists := dc.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return workflow, nil
}

// ExportWorkflow converts workflow to n8n JSON format
func (dc *DAGCompiler) ExportWorkflow(workflowID string) (string, error) {
	workflow, err := dc.GetWorkflow(workflowID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow: %w", err)
	}

	return string(data), nil
}

// Helper function
func containsAny(tokens []string, keywords []string) bool {
	for _, token := range tokens {
		for _, keyword := range keywords {
			if token == keyword {
				return true
			}
		}
	}
	return false
}
