package wasm

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WazeroRuntime implements the Runtime interface using Wazero
type WazeroRuntime struct {
	runtime wazero.Runtime
}

// NewWazeroRuntime creates a new Wazero-based WASM runtime
func NewWazeroRuntime(ctx context.Context) (*WazeroRuntime, error) {
	rt := wazero.NewRuntime(ctx)

	// Instantiate WASI
	_, err := wasi_snapshot_preview1.Instantiate(ctx, rt)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &WazeroRuntime{
		runtime: rt,
	}, nil
}

// LoadModule loads a WASM module from a file
func (w *WazeroRuntime) LoadModule(ctx context.Context, path string) (*Module, error) {
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}

	compiled, err := w.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	mod, err := w.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	log.Info().Str("path", path).Msg("WASM module loaded")

	return &Module{
		Name: path,
		mod:  mod,
	}, nil
}

// LoadModuleBytes loads a WASM module from bytes
func (w *WazeroRuntime) LoadModuleBytes(ctx context.Context, name string, wasmBytes []byte) (*Module, error) {
	compiled, err := w.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	mod, err := w.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	log.Info().Str("name", name).Msg("WASM module loaded from bytes")

	return &Module{
		Name: name,
		mod:  mod,
	}, nil
}

// Close cleans up the runtime
func (w *WazeroRuntime) Close(ctx context.Context) error {
	return w.runtime.Close(ctx)
}
