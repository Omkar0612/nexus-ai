package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// HardwareProfile describes the compute capabilities of a NEXUS node.
type HardwareProfile struct {
	HasGPU       bool    `json:"has_gpu"`
	TotalRAM     uint64  `json:"total_ram"`
	CPUModel     string  `json:"cpu_model"`
	LoadAverage  float64 `json:"load_average"`
}

// Node represents a single instance of NEXUS running on a device (Phone, PC, VPS).
type Node struct {
	ID       string          `json:"id"`
	Address  string          `json:"address"` // e.g., "192.168.1.50:7070"
	Profile  HardwareProfile `json:"profile"`
	LastSeen time.Time       `json:"-"`
}

// TaskRequest represents a payload sent from a weak node to a strong node.
type TaskRequest struct {
	TaskType string `json:"task_type"` // e.g., "IMAGE_GEN", "LLM_INFERENCE"
	Payload  []byte `json:"payload"`
}

// TaskResponse represents the result of offloaded computation.
type TaskResponse struct {
	Result []byte `json:"result"`
	Error  string `json:"error,omitempty"`
}

// NodeClient handles the HTTP communication between peers in the mesh.
type NodeClient interface {
	Dispatch(ctx context.Context, targetAddress string, req *TaskRequest) (*TaskResponse, error)
}
