package agents

import (
	"testing"
)

func TestAnalyzeEmotionFrustrated(t *testing.T) {
	ctx := AnalyzeEmotion("This is not working and it's broken again ugh")
	if ctx.PrimaryEmotion != "frustrated" {
		t.Errorf("expected frustrated, got %s", ctx.PrimaryEmotion)
	}
}

func TestAnalyzeEmotionUrgent(t *testing.T) {
	ctx := AnalyzeEmotion("This is urgent!! Need this ASAP, critical deadline!")
	if ctx.PrimaryEmotion != "urgent" {
		t.Errorf("expected urgent, got %s", ctx.PrimaryEmotion)
	}
}

func TestAnalyzeEmotionNeutral(t *testing.T) {
	ctx := AnalyzeEmotion("Can you help me write a function?")
	if ctx.PrimaryEmotion != "neutral" {
		t.Errorf("expected neutral, got %s", ctx.PrimaryEmotion)
	}
}

func TestAdaptToneFrustrated(t *testing.T) {
	ctx := EmotionalContext{PrimaryEmotion: "frustrated"}
	tone := AdaptTone(ctx)
	if tone.Formality != "empathetic" {
		t.Errorf("expected empathetic, got %s", tone.Formality)
	}
	if tone.Verbosity != "brief" {
		t.Errorf("expected brief, got %s", tone.Verbosity)
	}
}

func TestSelfHealingRetry(t *testing.T) {
	agent := NewSelfHealingAgent()
	res := agent.RecordFailure("test-task", "connection refused")
	if res.Status != "retrying" {
		t.Errorf("expected retrying, got %s", res.Status)
	}
}

func TestSelfHealingEscalate(t *testing.T) {
	agent := NewSelfHealingAgent()
	agent.maxRetry = 2
	for i := 0; i < 3; i++ {
		agent.RecordFailure("bad-task", "timeout")
	}
	res := agent.RecordFailure("bad-task", "timeout")
	if res.Status != "escalated" {
		t.Errorf("expected escalated, got %s", res.Status)
	}
}

func TestGoalStore(t *testing.T) {
	gs := NewGoalStore("user1")
	goal := gs.SetGoal("Launch NEXUS", "Get to 1000 GitHub stars", 5, nil)
	if goal.Title != "Launch NEXUS" {
		t.Errorf("unexpected title: %s", goal.Title)
	}
	goals := gs.List()
	if len(goals) != 1 {
		t.Errorf("expected 1 goal, got %d", len(goals))
	}
}

func TestPersonaEngineSwitch(t *testing.T) {
	pe, err := NewPersonaEngine(t.TempDir())
	if err != nil {
		t.Fatalf("NewPersonaEngine: %v", err)
	}
	if err := pe.Switch("work"); err != nil {
		t.Fatalf("Switch to work: %v", err)
	}
	if pe.Active().Name != "work" {
		t.Errorf("expected work persona, got %s", pe.Active().Name)
	}
}

func TestPersonaEngineBadSwitch(t *testing.T) {
	pe, _ := NewPersonaEngine(t.TempDir())
	if err := pe.Switch("nonexistent"); err == nil {
		t.Error("expected error for nonexistent persona")
	}
}

func TestOfflineManager(t *testing.T) {
	om := NewOfflineManager()
	// With a fast timeout and a valid URL, just test it doesn't panic
	_ = om.IsOnline()
	om.QueueTask("send-report", "send morning report", "telegram")
	tasks := om.FlushQueue()
	if len(tasks) != 1 {
		t.Errorf("expected 1 queued task, got %d", len(tasks))
	}
	tasks = om.FlushQueue()
	if len(tasks) != 0 {
		t.Errorf("expected 0 after flush, got %d", len(tasks))
	}
}

func TestAdaptiveLearner(t *testing.T) {
	al := NewAdaptiveLearner(nil, "user1")
	al.Learn("web-search", "help me research AI agents", "- bullet\n- point", true)
	al.Learn("code", "write a function", "```go\nfunc main(){}\n```", true)
	prompt := al.PersonalizeSystemPrompt("You are NEXUS.")
	if prompt == "" {
		t.Error("personalized prompt should not be empty")
	}
}
