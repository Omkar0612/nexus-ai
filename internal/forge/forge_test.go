package forge

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/Omkar0612/nexus-ai/internal/plugin/wasm"
)

func TestLocalCompileAndLoad(t *testing.T) {
	// Skip if tinygo isn't installed on the test machine
	if _, err := exec.LookPath("tinygo"); err != nil {
		t.Skip("tinygo not installed, skipping local compile test")
	}

	// Minimal valid Wasm agent source code
	source := `
package main

func main() {}

//export run
func run() {
	// Simple stub for testing
}
`
	ctx := context.Background()
	runtime := wasm.NewRuntime(ctx)
	defer runtime.Close(ctx)

	f := New(runtime)

	var stdout bytes.Buffer
	agent, err := f.BuildAndLoad(ctx, "test_agent", source, &stdout)
	if err != nil {
		t.Fatalf("BuildAndLoad failed: %v", err)
	}
	defer agent.Close(ctx)

	// Verify we can execute the exported function
	_, err = agent.Execute(ctx, "")
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}
}
