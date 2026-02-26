package agents

import (
	"testing"
)

func TestLoopDetectorNoLoop(t *testing.T) {
	d := NewLoopDetector(3, 20)
	for i := 0; i < 2; i++ {
		isLoop, _ := d.Record("web-search", "AI agents open source")
		if isLoop {
			t.Error("should not detect loop before threshold")
		}
	}
}

func TestLoopDetectorDetectsLoop(t *testing.T) {
	d := NewLoopDetector(3, 20)
	var fired *LoopEvent
	d.SetLoopCallback(func(e LoopEvent) { fired = &e })

	for i := 0; i < 3; i++ {
		isLoop, event := d.Record("web-search", "nexus AI open source agent")
		if i < 2 && isLoop {
			t.Errorf("unexpected loop at iteration %d", i)
		}
		if i == 2 {
			if !isLoop {
				t.Error("expected loop detection on 3rd identical call")
			}
			if event == nil {
				t.Fatal("expected non-nil loop event")
			}
			if event.RepeatCount < 3 {
				t.Errorf("expected repeat count >= 3, got %d", event.RepeatCount)
			}
		}
	}
	if fired == nil {
		t.Error("callback was never fired")
	}
}

func TestLoopDetectorDifferentInputsNoLoop(t *testing.T) {
	d := NewLoopDetector(3, 20)
	inputs := []string{"query one", "query two", "query three", "query four"}
	for _, inp := range inputs {
		isLoop, _ := d.Record("web-search", inp)
		if isLoop {
			t.Errorf("false loop for different input: %s", inp)
		}
	}
}

func TestLoopDetectorReset(t *testing.T) {
	d := NewLoopDetector(3, 20)
	for i := 0; i < 2; i++ {
		d.Record("web-search", "same query")
	}
	d.Reset()
	// After reset, count starts fresh — should not loop
	isLoop, _ := d.Record("web-search", "same query")
	if isLoop {
		t.Error("should not loop after reset")
	}
}
