package predictive

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PatternType categorizes user behavior patterns
type PatternType string

const (
	PatternTemporal   PatternType = "temporal"   // Time-based patterns
	PatternContextual PatternType = "contextual" // Context-dependent patterns
	PatternSequential PatternType = "sequential" // Task sequences
)

// UserPattern represents a learned behavior pattern
type UserPattern struct {
	ID           string         `json:"id"`
	Type         PatternType    `json:"type"`
	Trigger      map[string]any `json:"trigger"`
	ExpectedTask string         `json:"expected_task"`
	Confidence   float64        `json:"confidence"`
	Occurrences  int            `json:"occurrences"`
	LastSeen     time.Time      `json:"last_seen"`
	Context      map[string]any `json:"context"`
}

// Prediction represents a forecasted task
type Prediction struct {
	ID           string        `json:"id"`
	TaskType     string        `json:"task_type"`
	ExpectedTime time.Time     `json:"expected_time"`
	Confidence   float64       `json:"confidence"`
	PatternID    string        `json:"pattern_id"`
	PreComputed  bool          `json:"pre_computed"`
	CachedResult any           `json:"cached_result,omitempty"`
	ComputeTime  time.Duration `json:"compute_time"`
}

// TaskRecord is a task execution record for learning
type TaskRecord struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
	Duration  time.Duration  `json:"duration"`
	Success   bool           `json:"success"`
}

// Config holds predictive engine configuration
type Config struct {
	ConfidenceThreshold  float64
	HistorySize          int
	LearningInterval     time.Duration
	PredictionInterval   time.Duration
	MinPatternOccurrence int
	PreComputeQueueSize  int
}

// DefaultConfig returns default predictive engine configuration
func DefaultConfig() *Config {
	return &Config{
		ConfidenceThreshold:  getEnvFloat("NEXUS_PREDICTIVE_CONFIDENCE", 0.7),
		HistorySize:          getEnvInt("NEXUS_PREDICTIVE_HISTORY_SIZE", 1000),
		LearningInterval:     getEnvDuration("NEXUS_PREDICTIVE_LEARNING_INTERVAL", 60*time.Second),
		PredictionInterval:   getEnvDuration("NEXUS_PREDICTIVE_PREDICTION_INTERVAL", 30*time.Second),
		MinPatternOccurrence: getEnvInt("NEXUS_PREDICTIVE_MIN_OCCURRENCE", 3),
		PreComputeQueueSize:  getEnvInt("NEXUS_PREDICTIVE_QUEUE_SIZE", 50),
	}
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
	config          *Config
	learningEnabled bool
}

// NewPredictiveEngine creates a new prediction engine
func NewPredictiveEngine(config *Config) (*PredictiveEngine, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		return nil, fmt.Errorf("confidence threshold must be between 0 and 1, got: %f", config.ConfidenceThreshold)
	}
	if config.HistorySize < 10 {
		return nil, fmt.Errorf("history size must be at least 10, got: %d", config.HistorySize)
	}
	if config.MinPatternOccurrence < 2 {
		return nil, fmt.Errorf("minimum pattern occurrence must be at least 2, got: %d", config.MinPatternOccurrence)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PredictiveEngine{
		patterns:        make(map[string]*UserPattern),
		predictions:     make(map[string]*Prediction),
		taskHistory:     make([]*TaskRecord, 0, config.HistorySize),
		precomputeQueue: make(chan *Prediction, config.PreComputeQueueSize),
		ctx:             ctx,
		cancel:          cancel,
		config:          config,
		learningEnabled: true,
	}, nil
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
func (pe *PredictiveEngine) RecordTask(record *TaskRecord) error {
	if record == nil {
		return fmt.Errorf("task record cannot be nil")
	}
	if record.ID == "" {
		return fmt.Errorf("task record ID cannot be empty")
	}
	if record.Type == "" {
		return fmt.Errorf("task record type cannot be empty")
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.taskHistory = append(pe.taskHistory, record)

	// Keep only recent history
	if len(pe.taskHistory) > pe.config.HistorySize {
		pe.taskHistory = pe.taskHistory[len(pe.taskHistory)-pe.config.HistorySize:]
	}

	log.Debug().
		Str("task_id", record.ID).
		Str("type", record.Type).
		Msg("Recorded task execution")

	return nil
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
	ticker := time.NewTicker(pe.config.LearningInterval)
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
	// Clone task history to avoid race conditions
	pe.mu.RLock()
	if len(pe.taskHistory) < 10 {
		pe.mu.RUnlock()
		return // Need sufficient data
	}
	history := make([]*TaskRecord, len(pe.taskHistory))
	copy(history, pe.taskHistory)
	pe.mu.RUnlock()

	// Now work with cloned data without holding lock
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Detect temporal patterns (e.g., daily at 9 AM)
	pe.detectTemporalPatterns(history)

	// Detect sequential patterns (e.g., task A → task B)
	pe.detectSequentialPatterns(history)

	// Detect contextual patterns (e.g., when X happens, do Y)
	pe.detectContextualPatterns(history)

	log.Info().
		Int("patterns", len(pe.patterns)).
		Msg("Pattern learning cycle completed")
}

// detectTemporalPatterns finds time-based patterns
func (pe *PredictiveEngine) detectTemporalPatterns(history []*TaskRecord) {
	// Group tasks by hour of day
	hourlyTasks := make(map[int]map[string]int)

	for _, task := range history {
		hour := task.Timestamp.Hour()
		if hourlyTasks[hour] == nil {
			hourlyTasks[hour] = make(map[string]int)
		}
		hourlyTasks[hour][task.Type]++
	}

	// Find patterns with high frequency
	for hour, tasks := range hourlyTasks {
		for taskType, count := range tasks {
			if count >= pe.config.MinPatternOccurrence {
				patternID := fmt.Sprintf("temporal-%d-%s", hour, taskType)
				confidence := float64(count) / float64(len(history))

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
func (pe *PredictiveEngine) detectSequentialPatterns(history []*TaskRecord) {
	if len(history) < 2 {
		return
	}

	// Find common sequences
	sequences := make(map[string]int)

	for i := 0; i < len(history)-1; i++ {
		seq := fmt.Sprintf("%s->%s", history[i].Type, history[i+1].Type)
		sequences[seq]++
	}

	// Create patterns for frequent sequences
	for seq, count := range sequences {
		if count >= pe.config.MinPatternOccurrence {
			patternID := fmt.Sprintf("sequential-%s", seq)
			confidence := float64(count) / float64(len(history))

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
func (pe *PredictiveEngine) detectContextualPatterns(history []*TaskRecord) {
	// Group by context keys
	contextualTasks := make(map[string]map[string]int)

	for _, task := range history {
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
			if count >= pe.config.MinPatternOccurrence {
				patternID := fmt.Sprintf("contextual-%s-%s", contextKey, taskType)
				confidence := float64(count) / float64(len(history))

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
	ticker := time.NewTicker(pe.config.PredictionInterval)
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
		if pattern.Confidence < pe.config.ConfidenceThreshold {
			continue
		}

		var expectedTime time.Time

		switch pattern.Type {
		case PatternTemporal:
			// Predict for next occurrence of hour
			if hour, ok := pattern.Trigger["hour"].(int); ok {
				expectedTime = time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
				if expectedTime.Before(now) {
					expectedTime = expectedTime.Add(24 * time.Hour)
				}
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
	// This would call the agent's task executor with the predicted task
	// Example: result, err := pe.taskExecutor.Execute(prediction.TaskType, prediction.Context)

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
		"patterns_learned":   len(pe.patterns),
		"active_predictions": len(pe.predictions),
		"pre_computed_tasks": preComputeCount,
		"task_history_size":  len(pe.taskHistory),
		"avg_confidence":     math.Round(avgConfidence*100) / 100,
		"learning_enabled":   pe.learningEnabled,
	}
}

// Helper functions

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultValue
}
