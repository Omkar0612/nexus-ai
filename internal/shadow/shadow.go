package shadow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ExecutionMode defines how shadow execution behaves
type ExecutionMode int

const (
	ModePassive ExecutionMode = iota // Observe only, no side effects
	ModeActive                        // Full execution with rollback capability
	ModeABTest                        // Split traffic between strategies
)

// Strategy represents an agent execution strategy
type Strategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Config      map[string]any         `json:"config"`
	Enabled     bool                   `json:"enabled"`
	SuccessRate float64                `json:"success_rate"`
	AvgLatency  time.Duration          `json:"avg_latency"`
	Executions  int                    `json:"executions"`
}

// ShadowExecution records a parallel execution attempt
type ShadowExecution struct {
	ID            string         `json:"id"`
	StrategyID    string         `json:"strategy_id"`
	ParentTaskID  string         `json:"parent_task_id"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	Success       bool           `json:"success"`
	Result        any            `json:"result"`
	Error         string         `json:"error,omitempty"`
	Metrics       ExecutionMetrics `json:"metrics"`
}

// ExecutionMetrics tracks performance and quality
type ExecutionMetrics struct {
	Latency         time.Duration `json:"latency"`
	TokensUsed      int           `json:"tokens_used"`
	CostUSD         float64       `json:"cost_usd"`
	QualityScore    float64       `json:"quality_score"`
	Hallucinations  int           `json:"hallucinations"`
	ToolCalls       int           `json:"tool_calls"`
}

// EvolutionEvent represents a learned improvement
type EvolutionEvent struct {
	Timestamp      time.Time      `json:"timestamp"`
	StrategyID     string         `json:"strategy_id"`
	ChangeType     string         `json:"change_type"`
	OldValue       any            `json:"old_value"`
	NewValue       any            `json:"new_value"`
	Reason         string         `json:"reason"`
	ImpactScore    float64        `json:"impact_score"`
}

// ShadowManager coordinates parallel strategy execution
type ShadowManager struct {
	mu              sync.RWMutex
	strategies      map[string]*Strategy
	executions      []*ShadowExecution
	evolutions      []*EvolutionEvent
	mode            ExecutionMode
	ctx             context.Context
	cancel          context.CancelFunc
	evaluationChan  chan *ShadowExecution
	learningEnabled bool
}

// NewShadowManager creates a new shadow execution coordinator
func NewShadowManager(mode ExecutionMode) *ShadowManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ShadowManager{
		strategies:      make(map[string]*Strategy),
		executions:      make([]*ShadowExecution, 0),
		evolutions:      make([]*EvolutionEvent, 0),
		mode:            mode,
		ctx:             ctx,
		cancel:          cancel,
		evaluationChan:  make(chan *ShadowExecution, 100),
		learningEnabled: true,
	}
}

// Start begins shadow execution and learning
func (sm *ShadowManager) Start() error {
	log.Info().Str("mode", fmt.Sprintf("%v", sm.mode)).Msg("Starting shadow execution manager")

	// Start evaluation pipeline
	go sm.runEvaluationPipeline()

	// Start learning engine
	if sm.learningEnabled {
		go sm.runLearningEngine()
	}

	return nil
}

// Stop gracefully shuts down shadow manager
func (sm *ShadowManager) Stop() error {
	log.Info().Msg("Stopping shadow execution manager")
	sm.cancel()
	close(sm.evaluationChan)
	return nil
}

// RegisterStrategy adds a new execution strategy
func (sm *ShadowManager) RegisterStrategy(strategy *Strategy) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.strategies[strategy.ID]; exists {
		return fmt.Errorf("strategy %s already registered", strategy.ID)
	}

	sm.strategies[strategy.ID] = strategy
	log.Info().Str("strategy_id", strategy.ID).Str("name", strategy.Name).Msg("Registered strategy")
	return nil
}

// ExecuteShadow runs task with multiple strategies in parallel
func (sm *ShadowManager) ExecuteShadow(taskID string, primaryStrategy string, input any) ([]*ShadowExecution, error) {
	sm.mu.RLock()
	strategies := make([]*Strategy, 0, len(sm.strategies))
	for _, s := range sm.strategies {
		if s.Enabled {
			strategies = append(strategies, s)
		}
	}
	sm.mu.RUnlock()

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no enabled strategies")
	}

	executions := make([]*ShadowExecution, 0, len(strategies))
	var wg sync.WaitGroup
	resultsChan := make(chan *ShadowExecution, len(strategies))

	// Execute all strategies in parallel
	for _, strategy := range strategies {
		wg.Add(1)
		go func(s *Strategy) {
			defer wg.Done()
			exec := sm.executeStrategy(taskID, s, input)
			resultsChan <- exec
		}(strategy)
	}

	// Wait for all executions
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for exec := range resultsChan {
		executions = append(executions, exec)
		sm.evaluationChan <- exec
	}

	// Store executions
	sm.mu.Lock()
	sm.executions = append(sm.executions, executions...)
	sm.mu.Unlock()

	return executions, nil
}

// executeStrategy runs a single strategy
func (sm *ShadowManager) executeStrategy(taskID string, strategy *Strategy, input any) *ShadowExecution {
	exec := &ShadowExecution{
		ID:           fmt.Sprintf("shadow-%d", time.Now().UnixNano()),
		StrategyID:   strategy.ID,
		ParentTaskID: taskID,
		StartTime:    time.Now(),
	}

	// TODO: Implement actual strategy execution
	// For now, simulate execution
	time.Sleep(50 * time.Millisecond)

	exec.EndTime = time.Now()
	exec.Success = true
	exec.Result = map[string]any{"output": "simulated result"}
	exec.Metrics = ExecutionMetrics{
		Latency:      exec.EndTime.Sub(exec.StartTime),
		TokensUsed:   100,
		CostUSD:      0.001,
		QualityScore: 0.95,
		ToolCalls:    2,
	}

	log.Debug().
		Str("exec_id", exec.ID).
		Str("strategy", strategy.Name).
		Dur("latency", exec.Metrics.Latency).
		Msg("Shadow execution completed")

	return exec
}

// runEvaluationPipeline analyzes shadow executions
func (sm *ShadowManager) runEvaluationPipeline() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case exec := <-sm.evaluationChan:
			sm.evaluateExecution(exec)
		}
	}
}

// evaluateExecution analyzes execution performance
func (sm *ShadowManager) evaluateExecution(exec *ShadowExecution) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	strategy, exists := sm.strategies[exec.StrategyID]
	if !exists {
		return
	}

	// Update strategy metrics
	strategy.Executions++
	if exec.Success {
		strategy.SuccessRate = (strategy.SuccessRate*float64(strategy.Executions-1) + 1.0) / float64(strategy.Executions)
	} else {
		strategy.SuccessRate = (strategy.SuccessRate * float64(strategy.Executions-1)) / float64(strategy.Executions)
	}

	// Update average latency
	strategy.AvgLatency = time.Duration(
		(int64(strategy.AvgLatency)*int64(strategy.Executions-1) + int64(exec.Metrics.Latency)) / int64(strategy.Executions),
	)
}

// runLearningEngine identifies and applies improvements
func (sm *ShadowManager) runLearningEngine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.learnFromExecutions()
		}
	}
}

// learnFromExecutions analyzes patterns and evolves strategies
func (sm *ShadowManager) learnFromExecutions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.executions) < 10 {
		return // Need sufficient data
	}

	// Find best performing strategy
	var bestStrategy *Strategy
	bestScore := -1.0

	for _, s := range sm.strategies {
		if s.Executions < 5 {
			continue
		}

		// Combined score: success rate, speed, cost
		score := s.SuccessRate * (1.0 - float64(s.AvgLatency)/float64(time.Second))
		if score > bestScore {
			bestScore = score
			bestStrategy = s
		}
	}

	if bestStrategy == nil {
		return
	}

	// Apply learnings: promote best strategy
	for _, s := range sm.strategies {
		if s.ID != bestStrategy.ID && s.SuccessRate < bestStrategy.SuccessRate-0.1 {
			oldEnabled := s.Enabled
			s.Enabled = false

			evolution := &EvolutionEvent{
				Timestamp:   time.Now(),
				StrategyID:  s.ID,
				ChangeType:  "disable_underperforming",
				OldValue:    oldEnabled,
				NewValue:    false,
				Reason:      fmt.Sprintf("Success rate %.2f < best strategy %.2f", s.SuccessRate, bestStrategy.SuccessRate),
				ImpactScore: bestStrategy.SuccessRate - s.SuccessRate,
			}
			sm.evolutions = append(sm.evolutions, evolution)

			log.Info().
				Str("strategy", s.Name).
				Str("reason", evolution.Reason).
				Msg("Strategy disabled by learning engine")
		}
	}
}

// GetBestStrategy returns the highest performing strategy
func (sm *ShadowManager) GetBestStrategy() *Strategy {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var best *Strategy
	bestScore := -1.0

	for _, s := range sm.strategies {
		if !s.Enabled || s.Executions == 0 {
			continue
		}

		score := s.SuccessRate
		if score > bestScore {
			bestScore = score
			best = s
		}
	}

	return best
}

// GetEvolutionHistory returns learning events
func (sm *ShadowManager) GetEvolutionHistory() []*EvolutionEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return append([]*EvolutionEvent{}, sm.evolutions...)
}

// GetMetrics returns aggregate performance metrics
func (sm *ShadowManager) GetMetrics() map[string]any {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := map[string]any{
		"total_strategies": len(sm.strategies),
		"enabled_strategies": 0,
		"total_executions": len(sm.executions),
		"evolution_events": len(sm.evolutions),
		"strategies": make([]map[string]any, 0),
	}

	for _, s := range sm.strategies {
		if s.Enabled {
			metrics["enabled_strategies"] = metrics["enabled_strategies"].(int) + 1
		}

		metrics["strategies"] = append(metrics["strategies"].([]map[string]any), map[string]any{
			"id":           s.ID,
			"name":         s.Name,
			"enabled":      s.Enabled,
			"success_rate": s.SuccessRate,
			"avg_latency":  s.AvgLatency.String(),
			"executions":   s.Executions,
		})
	}

	return metrics
}
