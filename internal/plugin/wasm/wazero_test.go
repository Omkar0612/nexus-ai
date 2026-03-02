package wasm

import (
	"context"
	"testing"
)

// This test verifies that we can spin up a wazero runtime safely without CGO dependencies.
func TestRuntimeInit(t *testing.T) {
	ctx := context.Background()
	r := NewRuntime(ctx)
	defer r.Close(ctx)

	if r.runtime == nil {
		t.Fatal("expected wazero runtime to be initialized")
	}
}
