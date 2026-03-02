package predictive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PatternType categorizes user behavior patterns
type PatternType string

const (
	PatternTemporal  PatternType = "temporal"   // Time-based patterns
	PatternContextual PatternType = "contextual" // Context-dependent patterns  
	PatternSequential PatternType = "sequential" // Task sequences
)

// UserPattern represents a learned behavior pattern
type UserPattern struct {
	ID          string                 `json:"id"`
	Type        PatternType            `json:"type"`
	Trigger     map[string]any         `json:"trigger"`
	ExpectedTask string                `json:"expected_task"`
	Confidence  float64                `json:"confidence"`
	Occurrences int                    `json:"occurrences"`
	LastSeen    time.Time              `json:"last_seen"`
	Context     map[string]any         `json:"context"`
}

// Prediction represents a forecasted task
type Prediction struct {
	ID             string         `json:"id"`
	TaskType       string         `json:"task_type"`
	ExpectedTime   time.Time      `json:"expected_time"`
	Confidence     float64        `json:"confidence"`
	PatternID      string         `json:"pattern_id"`
	PreComputed    bool           `json:"pre_computed"`
	CachedResult   any            `json:"cached_result,omitempty"`
	ComputeTime    time.Duration  `json:"compute_time"`
}

// Task execution record for learning
type TaskRecord struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Timestamp   time.Time      `json:"timestamp"`
	Context     map[string]any `json:"context"`
	Duration    time.Duration  `json:"duration"`
	Success     bool           `json:"success"`
}

// PredictiveEngine learns patterns and pre-computes tasks
type PredictiveEngine struct {
	mu              sync.RWMutex
	patterns        map[string]*UserPattern
	predictions     map[string]*Prediction
	taskHistory     []*TaskRecord
	precomputeQueue chan *Prediction
	ctx             context.Context
	cancel          context.CancelFunc
	learningEnabled bool
	confidenceThreshold float64
}

// NewPredictiveEngine creates a new prediction engine
func NewPredictiveEngine(confidenceThreshold float64) *PredictiveEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &PredictiveEngine{
		patterns:            make(map[string]*UserPattern),
		predictions:         make(map[string]*Prediction),
		taskHistory:         make([]*TaskRecord, 0, 1000),
		precomputeQueue:     make(chan *Prediction, 50),
		ctx:                 ctx,
		cancel:              cancel,
		learningEnabled:     true,
		confidenceThreshold: confidenceThreshold,
	}
}

// Start begins pattern learning and pre-computation
func (pe *PredictiveEngine) Start() error {
	log.Info().Msg("Starting predictive engine")

	// Start pattern learning pipeline
	if pe.learningEnabled {
		go pe.runPatternLearning()
	}

	// Start prediction generator
	go pe.runPredictionGenerator()

	// Start pre-computation worker
	go pe.runPreComputeWorker()

	return nil
}

// Stop gracefully shuts down the engine
func (pe *PredictiveEngine) Stop() error {
	log.Info().Msg("Stopping predictive engine")
	pe.cancel()
	close(pe.precomputeQueue)
	return nil
}

// RecordTask adds a task execution to the learning dataset
func (pe *PredictiveEngine) RecordTask(record *TaskRecord) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.taskHistory = append(pe.taskHistory, record)

	// Keep only recent history (last 1000 tasks)
	if len(pe.taskHistory) > 1000 {
		pe.taskHistory = pe.taskHistory[len(pe.taskHistory)-1000:]
	}

	log.Debug().
		Str("task_id", record.ID).
		Str("type", record.Type).
		Msg("Recorded task execution")
}

// GetPrediction retrieves a prediction if available
func (pe *PredictiveEngine) GetPrediction(taskType string) *Prediction {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	for _, pred := range pe.predictions {
		if pred.TaskType == taskType && pred.PreComputed {
			log.Info().
				Str("task_type", taskType).
				Float64("confidence", pred.Confidence).
				Msg("Returning pre-computed result")
			return pred
		}
	}

	return nil
}

// GetUpcomingPredictions returns forecasted tasks
func (pe *PredictiveEngine) GetUpcomingPredictions(window time.Duration) []*Prediction {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	now := time.Now()
	upcoming := make([]*Prediction, 0)

	for _, pred := range pe.predictions {
		if pred.ExpectedTime.After(now) && pred.ExpectedTime.Before(now.Add(window)) {
			upcoming = append(upcoming, pred)
		}
	}

	return upcoming
}

// runPatternLearning analyzes task history for patterns
func (pe *PredictiveEngine) runPatternLearning() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pe.ctx.Done():
			return
		case <-ticker.C:
			pe.learnPatterns()
		}
	}
}

// learnPatterns extracts patterns from task history
func (pe *PredictiveEngine) learnPatterns() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if len(pe.taskHistory) < 10 {
		return // Need sufficient data
	}

	// Detect temporal patterns (e.g., daily at 9 AM)
	pe.detectTemporalPatterns()

	// Detect sequential patterns (e.g., task A → task B)
	pe.detectSequentialPatterns()

	// Detect contextual patterns (e.g., when X happens, do Y)
	pe.detectContextualPatterns()

	log.Info().
		Int("patterns", len(pe.patterns)).
		Msg("Pattern learning cycle completed")
}

// detectTemporalPatterns finds time-based patterns
func (pe *PredictiveEngine) detectTemporalPatterns() {
	// Group tasks by hour of day
	hourlyTasks := make(map[int]map[string]int)

	for _, task := range pe.taskHistory {
		hour := task.Timestamp.Hour()
		if hourlyTasks[hour] == nil {
			hourlyTasks[hour] = make(map[string]int)
		}
		hourlyTasks[hour][task.Type]++
	}

	// Find patterns with high frequency
	for hour, tasks := range hourlyTasks {
		for taskType, count := range tasks {
			if count >= 3 { // Minimum occurrences
				patternID := fmt.Sprintf("temporal-%d-%s", hour, taskType)
				confidence := float64(count) / float64(len(pe.taskHistory))

				pe.patterns[patternID] = &UserPattern{
					ID:   patternID,
					Type: PatternTemporal,
					Trigger: map[string]any{
						"hour": hour,
					},
					ExpectedTask: taskType,
					Confidence:   confidence,
					Occurrences:  count,
					LastSeen:     time.Now(),
				}

				log.Debug().
					Str("pattern_id", patternID).
					Float64("confidence", confidence).
					Msg("Detected temporal pattern")
			}
		}
	}
}

// detectSequentialPatterns finds task sequences
func (pe *PredictiveEngine) detectSequentialPatterns() {
	if len(pe.taskHistory) < 2 {
		return
	}

	// Find common sequences
	sequences := make(map[string]int)

	for i := 0; i < len(pe.taskHistory)-1; i++ {
		seq := fmt.Sprintf("%s->%s", pe.taskHistory[i].Type, pe.taskHistory[i+1].Type)
		sequences[seq]++
	}

	// Create patterns for frequent sequences
	for seq, count := range sequences {
		if count >= 3 {
			patternID := fmt.Sprintf("sequential-%s", seq)
			confidence := float64(count) / float64(len(pe.taskHistory))

			pe.patterns[patternID] = &UserPattern{
				ID:           patternID,
				Type:         PatternSequential,
				ExpectedTask: seq,
				Confidence:   confidence,
				Occurrences:  count,
				LastSeen:     time.Now(),
			}
		}
	}
}

// detectContextualPatterns finds context-dependent patterns
func (pe *PredictiveEngine) detectContextualPatterns() {
	// Group by context keys
	contextualTasks := make(map[string]map[string]int)

	for _, task := range pe.taskHistory {
		for key, value := range task.Context {
			contextKey := fmt.Sprintf("%s=%v", key, value)
			if contextualTasks[contextKey] == nil {
				contextualTasks[contextKey] = make(map[string]int)
			}
			contextualTasks[contextKey][task.Type]++
		}
	}

	// Create patterns
	for contextKey, tasks := range contextualTasks {
		for taskType, count := range tasks {
			if count >= 3 {
				patternID := fmt.Sprintf("contextual-%s-%s", contextKey, taskType)
				confidence := float64(count) / float64(len(pe.taskHistory))

				pe.patterns[patternID] = &UserPattern{
					ID:           patternID,
					Type:         PatternContextual,
					ExpectedTask: taskType,
					Confidence:   confidence,
					Occurrences:  count,
					LastSeen:     time.Now(),
				}
			}
		}
	}
}

// runPredictionGenerator creates predictions from patterns
func (pe *PredictiveEngine) runPredictionGenerator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pe.ctx.Done():
			return
		case <-ticker.C:
			pe.generatePredictions()
		}
	}
}

// generatePredictions creates forecasts from patterns
func (pe *PredictiveEngine) generatePredictions() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	now := time.Now()

	for _, pattern := range pe.patterns {
		if pattern.Confidence < pe.confidenceThreshold {
			continue
		}

		var expectedTime time.Time

		switch pattern.Type {
		case PatternTemporal:
			// Predict for next occurrence of hour
			hour := pattern.Trigger["hour"].(int)
			expectedTime = time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
			if expectedTime.Before(now) {
				expectedTime = expectedTime.Add(24 * time.Hour)
			}

		case PatternSequential:
			// Predict based on last task
			if len(pe.taskHistory) > 0 {
				lastTask := pe.taskHistory[len(pe.taskHistory)-1]
				expectedTime = lastTask.Timestamp.Add(5 * time.Minute)
			}

		case PatternContextual:
			// Predict when context matches
			expectedTime = now.Add(10 * time.Minute)
		}

		if expectedTime.IsZero() {
			continue
		}

		prediction := &Prediction{
			ID:           fmt.Sprintf("pred-%d", time.Now().UnixNano()),
			TaskType:     pattern.ExpectedTask,
			ExpectedTime: expectedTime,
			Confidence:   pattern.Confidence,
			PatternID:    pattern.ID,
			PreComputed:  false,
		}

		pe.predictions[prediction.ID] = prediction

		// Queue for pre-computation if confidence is high
		if pattern.Confidence >= 0.8 {
			select {
			case pe.precomputeQueue <- prediction:
			default:
				// Queue full, skip
			}
		}
	}
}

// runPreComputeWorker executes predicted tasks in advance
func (pe *PredictiveEngine) runPreComputeWorker() {
	for {
		select {
		case <-pe.ctx.Done():
			return
		case prediction := <-pe.precomputeQueue:
			pe.executePreComputation(prediction)
		}
	}
}

// executePreComputation runs task in advance and caches result
func (pe *PredictiveEngine) executePreComputation(prediction *Prediction) {
	start := time.Now()

	log.Info().
		Str("task_type", prediction.TaskType).
		Time("expected_time", prediction.ExpectedTime).
		Float64("confidence", prediction.Confidence).
		Msg("Pre-computing task")

	// TODO: Implement actual task execution
	// For now, simulate computation
	time.Sleep(100 * time.Millisecond)

	pe.mu.Lock()
	prediction.PreComputed = true
	prediction.CachedResult = map[string]any{
		"result": "pre-computed output",
		"status": "success",
	}
	prediction.ComputeTime = time.Since(start)
	pe.mu.Unlock()

	log.Info().
		Str("prediction_id", prediction.ID).
		Dur("compute_time", prediction.ComputeTime).
		Msg("Pre-computation completed")
}

// GetPatterns returns all learned patterns
func (pe *PredictiveEngine) GetPatterns() []*UserPattern {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	patterns := make([]*UserPattern, 0, len(pe.patterns))
	for _, p := range pe.patterns {
		patterns = append(patterns, p)
	}
	return patterns
}

// GetMetrics returns engine performance metrics
func (pe *PredictiveEngine) GetMetrics() map[string]any {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	preComputeCount := 0
	var avgConfidence float64

	for _, pred := range pe.predictions {
		if pred.PreComputed {
			preComputeCount++
		}
		avgConfidence += pred.Confidence
	}

	if len(pe.predictions) > 0 {
		avgConfidence /= float64(len(pe.predictions))
	}

	return map[string]any{
		"patterns_learned":    len(pe.patterns),
		"active_predictions":  len(pe.predictions),
		"pre_computed_tasks":  preComputeCount,
		"task_history_size":   len(pe.taskHistory),
		"avg_confidence":      math.Round(avgConfidence*100) / 100,
		"learning_enabled":    pe.learningEnabled,
	}
}
