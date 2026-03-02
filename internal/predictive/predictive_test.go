package predictive

import (
	"testing"
	"time"
)

func TestNewPredictiveEngine(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid confidence - negative",
			config: &Config{
				ConfidenceThreshold: -0.1,
				HistorySize:         100,
				MinPatternOccurrence: 3,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence - too high",
			config: &Config{
				ConfidenceThreshold: 1.5,
				HistorySize:         100,
				MinPatternOccurrence: 3,
			},
			wantErr: true,
		},
		{
			name: "invalid history size",
			config: &Config{
				ConfidenceThreshold: 0.7,
				HistorySize:         5,
				MinPatternOccurrence: 3,
			},
			wantErr: true,
		},
		{
			name: "invalid min occurrence",
			config: &Config{
				ConfidenceThreshold: 0.7,
				HistorySize:         100,
				MinPatternOccurrence: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPredictiveEngine(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPredictiveEngine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecordTask(t *testing.T) {
	engine, _ := NewPredictiveEngine(DefaultConfig())

	tests := []struct {
		name    string
		record  *TaskRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: &TaskRecord{
				ID:        "task-1",
				Type:      "inference",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "nil record",
			record:  nil,
			wantErr: true,
		},
		{
			name: "empty ID",
			record: &TaskRecord{
				ID:   "",
				Type: "inference",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			record: &TaskRecord{
				ID:   "task-2",
				Type: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RecordTask(tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectTemporalPatterns(t *testing.T) {
	config := DefaultConfig()
	config.MinPatternOccurrence = 2
	engine, _ := NewPredictiveEngine(config)

	// Add tasks at the same hour multiple times
	now := time.Now()
	history := []*TaskRecord{
		{ID: "1", Type: "morning_report", Timestamp: now.Add(-24 * time.Hour)},
		{ID: "2", Type: "morning_report", Timestamp: now.Add(-48 * time.Hour)},
		{ID: "3", Type: "morning_report", Timestamp: now.Add(-72 * time.Hour)},
		{ID: "4", Type: "other_task", Timestamp: now.Add(-12 * time.Hour)},
	}

	engine.detectTemporalPatterns(history)

	if len(engine.patterns) == 0 {
		t.Error("Expected to detect at least one temporal pattern")
	}

	// Check if pattern was detected
	found := false
	for _, pattern := range engine.patterns {
		if pattern.Type == PatternTemporal && pattern.ExpectedTask == "morning_report" {
			found = true
			if pattern.Occurrences < 2 {
				t.Errorf("Expected at least 2 occurrences, got %d", pattern.Occurrences)
			}
		}
	}
	if !found {
		t.Error("Expected to find morning_report temporal pattern")
	}
}

func TestDetectSequentialPatterns(t *testing.T) {
	config := DefaultConfig()
	config.MinPatternOccurrence = 2
	engine, _ := NewPredictiveEngine(config)

	// Add sequential tasks
	history := []*TaskRecord{
		{ID: "1", Type: "fetch_data", Timestamp: time.Now()},
		{ID: "2", Type: "process_data", Timestamp: time.Now()},
		{ID: "3", Type: "fetch_data", Timestamp: time.Now()},
		{ID: "4", Type: "process_data", Timestamp: time.Now()},
		{ID: "5", Type: "fetch_data", Timestamp: time.Now()},
		{ID: "6", Type: "process_data", Timestamp: time.Now()},
	}

	engine.detectSequentialPatterns(history)

	if len(engine.patterns) == 0 {
		t.Error("Expected to detect sequential pattern")
	}

	// Check for fetch -> process pattern
	found := false
	for _, pattern := range engine.patterns {
		if pattern.Type == PatternSequential {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find sequential pattern")
	}
}

func TestGetPrediction(t *testing.T) {
	engine, _ := NewPredictiveEngine(DefaultConfig())

	// Add a prediction
	engine.predictions["pred-1"] = &Prediction{
		ID:          "pred-1",
		TaskType:    "test_task",
		PreComputed: true,
	}

	// Should find it
	pred := engine.GetPrediction("test_task")
	if pred == nil {
		t.Error("Expected to find prediction, got nil")
	}

	// Should not find non-existent
	pred = engine.GetPrediction("non_existent")
	if pred != nil {
		t.Error("Expected nil for non-existent prediction")
	}
}

func TestGetUpcomingPredictions(t *testing.T) {
	engine, _ := NewPredictiveEngine(DefaultConfig())

	now := time.Now()
	engine.predictions = map[string]*Prediction{
		"pred-1": {
			ID:           "pred-1",
			ExpectedTime: now.Add(5 * time.Minute), // Within window
		},
		"pred-2": {
			ID:           "pred-2",
			ExpectedTime: now.Add(2 * time.Hour), // Outside window
		},
		"pred-3": {
			ID:           "pred-3",
			ExpectedTime: now.Add(-1 * time.Minute), // In the past
		},
	}

	upcoming := engine.GetUpcomingPredictions(30 * time.Minute)
	if len(upcoming) != 1 {
		t.Errorf("Expected 1 upcoming prediction, got %d", len(upcoming))
	}
	if upcoming[0].ID != "pred-1" {
		t.Errorf("Expected pred-1, got %s", upcoming[0].ID)
	}
}

func TestGetMetrics(t *testing.T) {
	engine, _ := NewPredictiveEngine(DefaultConfig())

	// Add some data
	engine.RecordTask(&TaskRecord{
		ID:   "task-1",
		Type: "test",
	})
	engine.patterns["pattern-1"] = &UserPattern{ID: "pattern-1"}
	engine.predictions["pred-1"] = &Prediction{
		ID:          "pred-1",
		PreComputed: true,
		Confidence:  0.9,
	}

	metrics := engine.GetMetrics()

	if metrics["patterns_learned"].(int) != 1 {
		t.Errorf("Expected 1 pattern, got %v", metrics["patterns_learned"])
	}
	if metrics["active_predictions"].(int) != 1 {
		t.Errorf("Expected 1 prediction, got %v", metrics["active_predictions"])
	}
	if metrics["pre_computed_tasks"].(int) != 1 {
		t.Errorf("Expected 1 pre-computed task, got %v", metrics["pre_computed_tasks"])
	}
}
