package forge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Compiler defines the interface for converting Go source code into WebAssembly.
type Compiler interface {
	Compile(ctx context.Context, sourceCode string) ([]byte, error)
}

// LocalCompiler uses a locally installed TinyGo toolchain to build Wasm agents.
type LocalCompiler struct{}

func NewLocalCompiler() *LocalCompiler {
	return &LocalCompiler{}
}

// Compile writes the source to a temp directory and invokes `tinygo build`.
func (c *LocalCompiler) Compile(ctx context.Context, sourceCode string) ([]byte, error) {
	// Check if tinygo is installed
	if _, err := exec.LookPath("tinygo"); err != nil {
		return nil, fmt.Errorf("tinygo is not installed or not in PATH: %w", err)
	}

	// Create a temporary directory for the isolated build
	tmpDir, err := os.MkdirTemp("", "nexus-forge-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the raw agent source code to main.go
	mainPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(sourceCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	// Output path for the compiled WebAssembly binary
	outPath := filepath.Join(tmpDir, "agent.wasm")

	// Build the command: tinygo build -o agent.wasm -target=wasi main.go
	cmd := exec.CommandContext(ctx, "tinygo", "build", "-o", outPath, "-target=wasi", "-no-debug", "main.go")
	cmd.Dir = tmpDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tinygo compilation failed: %v\nstderr: %s", err, stderr.String())
	}

	// Read the compiled WebAssembly module
	wasmBytes, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compiled wasm: %w", err)
	}

	return wasmBytes, nil
}
