package fuzzer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/plugin/wasm"
)

// mockAgent wraps the Execute function to simulate vulnerabilities.
type mockAgent struct {
	*wasm.AgentModule
}

// Override the Execute method for testing purposes
func (m *mockAgent) Execute(ctx context.Context, input string) (string, error) {
	// Simulate an infinite loop if the payload is huge (10MB)
	if len(input) > 5*1024*1024 {
		time.Sleep(5 * time.Second) // Will trigger context timeout
	}

	// Simulate a crash on null byte
	for _, c := range input {
		if c == '\x00' {
			return "", fmt.Errorf("wasm runtime panic: invalid memory boundary")
		}
	}

	return "success", nil
}

func (m *mockAgent) Name() string {
	return "test_mock_agent"
}

func TestAgenticFuzzer(t *testing.T) {
	fuzzer := New(100 * time.Millisecond) // Short timeout for testing

	// In a real scenario, this is returned by Forge.BuildAndLoad
	agent := &mockAgent{}

	report := fuzzer.Fuzz(context.Background(), agent.AgentModule)

	if report.Passed {
		t.Fatal("expected fuzzer to fail due to simulated vulnerabilities")
	}

	if report.TestsRun != len(SecurityPayloads)+len(EdgeCasePayloads) {
		t.Errorf("expected %d tests run, got %d", len(SecurityPayloads)+len(EdgeCasePayloads), report.TestsRun)
	}

	if report.Vulnerabilities != 2 {
		t.Errorf("expected exactly 2 vulnerabilities (Null Byte and OOM Loop), got %d", report.Vulnerabilities)
	}
}
