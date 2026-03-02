package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CloudCompiler uses a remote NEXUS build server to compile WebAssembly.
// This removes the requirement for the user to have TinyGo installed locally.
type CloudCompiler struct {
	Endpoint string
}

func NewCloudCompiler(endpoint string) *CloudCompiler {
	if endpoint == "" {
		endpoint = "https://build.nexus.sh/v1/compile" // Public free compiling endpoint
	}
	return &CloudCompiler{Endpoint: endpoint}
}

// Compile sends the source code to the NEXUS cloud to be compiled to Wasm.
func (c *CloudCompiler) Compile(ctx context.Context, sourceCode string) ([]byte, error) {
	payload := map[string]string{
		"source": sourceCode,
		"target": "wasi",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloud compile request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud compile failed with status %d: %s", resp.StatusCode, string(errBody))
	}

	// The response body is the raw .wasm binary
	wasmBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read cloud response: %w", err)
	}

	return wasmBytes, nil
}
