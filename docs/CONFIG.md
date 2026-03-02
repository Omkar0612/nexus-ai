# Nexus AI Configuration Guide

This document describes all environment variables for configuring Nexus AI's advanced features.

## Mesh Network Configuration

The mesh network enables P2P GPU resource sharing across multiple nodes.

### Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXUS_MESH_PORT` | int | `5353` | UDP port for peer discovery (must be 1024-65535) |
| `NEXUS_MESH_DISCOVERY_INTERVAL` | duration | `5s` | How often to broadcast peer presence |
| `NEXUS_MESH_PEER_TIMEOUT` | duration | `60s` | Time before considering a peer stale |
| `NEXUS_MESH_TASK_QUEUE_SIZE` | int | `100` | Maximum number of queued tasks |
| `NEXUS_MESH_RESULT_QUEUE_SIZE` | int | `100` | Maximum number of queued results |

### Example Configuration

```bash
export NEXUS_MESH_PORT=5353
export NEXUS_MESH_DISCOVERY_INTERVAL=5s
export NEXUS_MESH_PEER_TIMEOUT=60s
export NEXUS_MESH_TASK_QUEUE_SIZE=100
export NEXUS_MESH_RESULT_QUEUE_SIZE=100
```

### Security Considerations

⚠️ **WARNING**: The current mesh implementation uses unencrypted UDP multicast.

**Production recommendations:**
1. Deploy mesh nodes on a private network/VLAN
2. Use firewall rules to restrict UDP port access
3. Enable network-level encryption (VPN/WireGuard)
4. Future versions will include TLS and authentication

### Usage Example

```go
package main

import (
    "github.com/Omkar0612/nexus-ai/internal/mesh"
)

func main() {
    // Create mesh manager with default config
    gpuInfo := mesh.GPUInfo{
        Model:       "NVIDIA RTX 4090",
        MemoryTotal: 24576,
        MemoryFree:  20480,
        ComputeCaps: "8.9",
        DriverVer:   "535.54.03",
    }
    
    manager, err := mesh.NewMeshManager(gpuInfo, nil)
    if err != nil {
        panic(err)
    }
    
    if err := manager.Start(); err != nil {
        panic(err)
    }
    defer manager.Stop()
    
    // Submit a task
    task := &mesh.TaskRequest{
        ID:             "task-123",
        Type:           "inference",
        RequiredMemory: 2048,
        Priority:       1,
    }
    
    if err := manager.SubmitTask(task); err != nil {
        panic(err)
    }
    
    // Get result
    result, err := manager.GetResult(30 * time.Second)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Task completed: %+v\n", result)
}
```

---

## Predictive Engine Configuration

The predictive engine learns user patterns and pre-computes tasks.

### Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXUS_PREDICTIVE_CONFIDENCE` | float | `0.7` | Minimum confidence threshold (0.0-1.0) |
| `NEXUS_PREDICTIVE_HISTORY_SIZE` | int | `1000` | Maximum number of tasks to remember |
| `NEXUS_PREDICTIVE_LEARNING_INTERVAL` | duration | `60s` | How often to analyze patterns |
| `NEXUS_PREDICTIVE_PREDICTION_INTERVAL` | duration | `30s` | How often to generate predictions |
| `NEXUS_PREDICTIVE_MIN_OCCURRENCE` | int | `3` | Minimum times a pattern must occur |
| `NEXUS_PREDICTIVE_QUEUE_SIZE` | int | `50` | Pre-computation queue size |

### Example Configuration

```bash
export NEXUS_PREDICTIVE_CONFIDENCE=0.7
export NEXUS_PREDICTIVE_HISTORY_SIZE=1000
export NEXUS_PREDICTIVE_LEARNING_INTERVAL=60s
export NEXUS_PREDICTIVE_PREDICTION_INTERVAL=30s
export NEXUS_PREDICTIVE_MIN_OCCURRENCE=3
export NEXUS_PREDICTIVE_QUEUE_SIZE=50
```

### Pattern Types

1. **Temporal Patterns**: Tasks that occur at specific times
   - Example: Daily report generation at 9 AM
   
2. **Sequential Patterns**: Tasks that follow each other
   - Example: Fetch data → Process data → Send notification
   
3. **Contextual Patterns**: Tasks triggered by specific conditions
   - Example: When temperature > 30°C, turn on cooling

### Usage Example

```go
package main

import (
    "github.com/Omkar0612/nexus-ai/internal/predictive"
)

func main() {
    // Create engine with default config
    engine, err := predictive.NewPredictiveEngine(nil)
    if err != nil {
        panic(err)
    }
    
    if err := engine.Start(); err != nil {
        panic(err)
    }
    defer engine.Stop()
    
    // Record task executions
    for i := 0; i < 10; i++ {
        record := &predictive.TaskRecord{
            ID:        fmt.Sprintf("task-%d", i),
            Type:      "morning_report",
            Timestamp: time.Now(),
            Context:   map[string]any{"hour": 9},
            Duration:  2 * time.Second,
            Success:   true,
        }
        engine.RecordTask(record)
    }
    
    // Wait for pattern learning
    time.Sleep(65 * time.Second)
    
    // Check for predictions
    prediction := engine.GetPrediction("morning_report")
    if prediction != nil && prediction.PreComputed {
        fmt.Printf("Found pre-computed result: %+v\n", prediction.CachedResult)
    }
    
    // Get metrics
    metrics := engine.GetMetrics()
    fmt.Printf("Engine metrics: %+v\n", metrics)
}
```

---

## Shadow Mode Configuration

Shadow mode runs experimental strategies in parallel with production.

### Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXUS_SHADOW_MODE` | string | `passive` | Execution mode: `passive`, `active`, or `abtest` |
| `NEXUS_SHADOW_MAX_STRATEGIES` | int | `5` | Maximum concurrent strategies |
| `NEXUS_SHADOW_EVALUATION_WINDOW` | duration | `1h` | Time window for performance evaluation |

### Execution Modes

1. **Passive Mode**: Observe only, no side effects
2. **Active Mode**: Full execution with rollback capability
3. **A/B Test Mode**: Split traffic between strategies

---

## n8n DAG Compiler Configuration

### Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEXUS_N8N_MAX_NODES` | int | `100` | Maximum nodes per workflow |
| `NEXUS_N8N_TIMEOUT` | duration | `5m` | Workflow execution timeout |

---

## General Best Practices

### Performance Tuning

1. **High-throughput scenarios**:
   ```bash
   NEXUS_MESH_TASK_QUEUE_SIZE=500
   NEXUS_PREDICTIVE_QUEUE_SIZE=200
   ```

2. **Memory-constrained environments**:
   ```bash
   NEXUS_PREDICTIVE_HISTORY_SIZE=500
   NEXUS_MESH_TASK_QUEUE_SIZE=50
   ```

3. **Fast learning**:
   ```bash
   NEXUS_PREDICTIVE_LEARNING_INTERVAL=30s
   NEXUS_PREDICTIVE_MIN_OCCURRENCE=2
   ```

### Monitoring

All modules expose metrics via their `GetMetrics()` methods:

```go
// Mesh metrics
meshMetrics := meshManager.GetActivePeers()

// Predictive metrics
predictiveMetrics := predictiveEngine.GetMetrics()

// Shadow metrics
shadowMetrics := shadowManager.GetMetrics()
```

### Logging

All modules use structured logging via `zerolog`. Configure log level:

```bash
export NEXUS_LOG_LEVEL=debug  # debug, info, warn, error
```

---

## Docker Compose Example

```yaml
version: '3.8'

services:
  nexus-ai-node1:
    image: nexus-ai:latest
    environment:
      - NEXUS_MESH_PORT=5353
      - NEXUS_PREDICTIVE_CONFIDENCE=0.8
      - NEXUS_SHADOW_MODE=passive
      - NEXUS_LOG_LEVEL=info
    ports:
      - "8080:8080"
    networks:
      - nexus-mesh

  nexus-ai-node2:
    image: nexus-ai:latest
    environment:
      - NEXUS_MESH_PORT=5353
      - NEXUS_PREDICTIVE_CONFIDENCE=0.8
      - NEXUS_SHADOW_MODE=passive
      - NEXUS_LOG_LEVEL=info
    ports:
      - "8081:8080"
    networks:
      - nexus-mesh

networks:
  nexus-mesh:
    driver: bridge
```

---

## Troubleshooting

### Mesh Network Issues

**Problem**: Peers not discovering each other

**Solutions**:
1. Check firewall allows UDP on configured port
2. Ensure nodes are on same network segment for multicast
3. Verify `NEXUS_MESH_PORT` matches across all nodes
4. Enable debug logging: `NEXUS_LOG_LEVEL=debug`

### Predictive Engine Issues

**Problem**: No patterns detected

**Solutions**:
1. Ensure sufficient task history (>10 tasks)
2. Lower `NEXUS_PREDICTIVE_MIN_OCCURRENCE` to 2
3. Increase `NEXUS_PREDICTIVE_HISTORY_SIZE`
4. Check tasks have consistent types and timestamps

### Performance Issues

**Problem**: High memory usage

**Solutions**:
1. Reduce `NEXUS_PREDICTIVE_HISTORY_SIZE`
2. Decrease queue sizes
3. Increase cleanup intervals

---

## Support

For issues and questions:
- GitHub Issues: https://github.com/Omkar0612/nexus-ai/issues
- Documentation: https://github.com/Omkar0612/nexus-ai/wiki
