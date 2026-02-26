package plugin

import (
	"context"
	"testing"
)

func TestPluginRegistry(t *testing.T) {
	reg := NewRegistry()

	echo := NewSkill("echo", "Echoes the command back", func(in Input) Output {
		return Output{Text: "ECHO: " + in.Command}
	})

	if err := reg.Register(echo); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Duplicate should error
	if err := reg.Register(echo); err == nil {
		t.Error("expected error on duplicate registration")
	}

	out, err := reg.Execute("echo", Input{Command: "hello world", Context: context.Background()})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.Text != "ECHO: hello world" {
		t.Errorf("unexpected output: %q", out.Text)
	}

	list := reg.List()
	if len(list) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(list))
	}
}

func TestPluginNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Execute("nonexistent", Input{Context: context.Background()})
	if err == nil {
		t.Error("expected error for unknown plugin")
	}
}
