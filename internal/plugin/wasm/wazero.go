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

// Execute triggers the exported "run" function inside the Wasm agent, passing a string payload.
func (a *AgentModule) Execute(ctx context.Context, input string) (string, error) {
	// 1. Get the export functions
	malloc := a.mod.ExportedFunction("malloc")
	free := a.mod.ExportedFunction("free")
	runFunc := a.mod.ExportedFunction("run")

	if runFunc == nil {
		return "", fmt.Errorf("agent %s does not export a 'run' function", a.name)
	}
	
	// Fast path: if the agent doesn't take args, just run it
	if malloc == nil || len(input) == 0 {
		_, err := runFunc.Call(ctx)
		if err != nil {
			return "", fmt.Errorf("agent execution failed: %w", err)
		}
		return "Executed successfully without payload", nil
	}

	// 2. Allocate memory inside Wasm sandbox for the input string
	inputSize := uint64(len(input))
	results, err := malloc.Call(ctx, inputSize)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory in wasm: %w", err)
	}
	inputPtr := results[0]

	// Ensure we free the memory when done
	defer free.Call(ctx, inputPtr)

	// 3. Write the string to the Wasm sandbox memory
	if !a.mod.Memory().Write(uint32(inputPtr), []byte(input)) {
		return "", fmt.Errorf("failed to write payload to wasm memory")
	}

	// 4. Execute the agent passing the pointer and length
	_, err = runFunc.Call(ctx, inputPtr, inputSize)
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

func hostLog(ctx context.Context, mod api.Module, ptr uint32, length uint32) {
	if bytes, ok := mod.Memory().Read(ptr, length); ok {
		fmt.Printf("[WASM Agent] %s\n", string(bytes))
	}
}
