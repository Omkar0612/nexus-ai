package wasm

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime manages hot-loaded WebAssembly agents safely inside NEXUS.
type Runtime struct {
	ctx     context.Context
	runtime wazero.Runtime
}

// AgentModule represents an instantiated, sandboxed Wasm agent.
type AgentModule struct {
	mod    api.Module
	name   string
	stdout io.Writer
}

// Name returns the identifier of the hot-loaded agent.
func (a *AgentModule) Name() string {
	return a.name
}

// NewRuntime initializes the zero-dependency Wasm runtime.
func NewRuntime(ctx context.Context) *Runtime {
	r := wazero.NewRuntime(ctx)
	
	// Instantiate WASI to allow Wasm modules to use basic features (like writing to console).
	// We do NOT map the filesystem or network by default (Zero-Trust sandbox).
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	
	// Export the NEXUS Host ABI
	// Agents will call these functions to request data from the main Go process safely.
	_, err := r.NewHostModuleBuilder("nexus_env").
		NewFunctionBuilder().
		WithFunc(hostLog).
		Export("host_log").
		Instantiate(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to mount nexus_env: %w", err))
	}

	return &Runtime{
		ctx:     ctx,
		runtime: r,
	}
}

// LoadAgent reads a .wasm binary and hot-loads it into the running system.
func (r *Runtime) LoadAgent(name string, wasmBytes []byte, stdout io.Writer) (*AgentModule, error) {
	if stdout == nil {
		stdout = os.Stdout
	}

	// Configure the sandbox limits. No files, no env vars.
	config := wazero.NewModuleConfig().
		WithName(name).
		WithStdout(stdout).
		WithStderr(os.Stderr)

	mod, err := r.runtime.InstantiateWithConfig(r.ctx, wasmBytes, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate wasm agent %s: %w", err)
	}

	return &AgentModule{
		mod:    mod,
		name:   name,
		stdout: stdout,
	}, nil
}

// Execute triggers the exported "run" function inside the Wasm agent.
func (a *AgentModule) Execute(ctx context.Context, input string) (string, error) {
	runFunc := a.mod.ExportedFunction("run")
	if runFunc == nil {
		return "", fmt.Errorf("agent %s does not export a 'run' function", a.name)
	}

	// Note: String passing between Go Host and Wasm Guest requires writing to Wasm memory.
	// For this initial implementation, we execute without arguments.
	// A full memory allocator for string passing will be added in the next PR.
	_, err := runFunc.Call(ctx)
	if err != nil {
		return "", fmt.Errorf("agent execution failed: %w", err)
	}

	return "Executed successfully", nil
}

// Close gracefully unloads the agent and frees memory.
func (a *AgentModule) Close(ctx context.Context) error {
	return a.mod.Close(ctx)
}

// Close gracefully shuts down the entire Wasm runtime.
func (r *Runtime) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}

// --- NEXUS HOST ABI FUNCTIONS ---

// hostLog allows the Wasm agent to write logs back to the NEXUS host logger.
func hostLog(ctx context.Context, mod api.Module, ptr uint32, length uint32) {
	if bytes, ok := mod.Memory().Read(ptr, length); ok {
		fmt.Printf("[WASM Agent] %s\n", string(bytes))
	}
}
