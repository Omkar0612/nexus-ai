# NEXUS AI: System Architecture

This document provides a high-level overview of the NEXUS AI ecosystem, detailing how the core components and the next-generation autonomous features (introduced in v1.8) interact within a single, lightweight Go binary.

## 1. High-Level Architecture

NEXUS is designed as an edge-first, multi-agent operating system. It eschews heavy containerization (Docker) and complex Python environments in favor of a monolithic Go binary with hot-swappable WebAssembly modules.

```mermaid
graph TD
    %% External Triggers
    User[User (CLI/WebUI/Mobile)]
    Events[External Events (Calendar, GitHub)]
    
    %% Gateway & Intelligence
    subgraph NEXUS Binary
        Gateway[API / SSE Gateway]
        Predictive[Predictive Pre-Computation]
        
        %% Core Bus
        Bus((Multi-Agent Event Bus))
        
        %% Core Engines
        subgraph Autonomous Engines
            Forge[Auto-Forge Wasm Compiler]
            Shadow[Shadow Mode Evolution]
            UI2API[UI-to-API Reverse Engineer]
        end
        
        %% Data & Security
        subgraph Storage & Security
            Vault[(AES-256 Vault)]
            Mem[(SQLite Vector DB)]
            HITL{Risk Gate HITL}
        end
        
        %% Execution Environment
        subgraph Execution Sandbox
            Wasm[Wazero Sandbox]
            Native[Native Agents]
        end
        
        %% Mesh
        Mesh[Hive-Mind Mesh Router]
    end
    
    %% Network Peers
    RemoteNode1[NEXUS Node: Desktop PC GPU]
    RemoteNode2[NEXUS Node: VPS Server]

    %% Connections
    User <--> Gateway
    Events --> Predictive
    Predictive --> Bus
    Gateway <--> Bus
    
    Bus <--> Forge
    Bus <--> Shadow
    Bus <--> UI2API
    
    Forge --> Wasm
    UI2API --> Forge
    UI2API --> Vault
    
    Wasm <--> HITL
    Native <--> HITL
    HITL <--> Vault
    HITL <--> Mem
    
    Bus <--> Mesh
    Mesh <--> RemoteNode1
    Mesh <--> RemoteNode2
```

## 2. Core Components

### 2.1 The Multi-Agent Event Bus
The central nervous system of NEXUS. All requests (from the user, or from other agents) are serialized as events and placed on the bus. Agents subscribe to specific event types (e.g., `WRITE_FILE`, `GENERATE_IMAGE`, `ANALYZE_REPO`) and publish their results back to the bus.

### 2.2 Security: AES-256 Vault & HITL Risk Gate
NEXUS operates on a Zero-Trust principle.
- **Vault**: All API keys, session cookies, and tokens are encrypted via `AES-256-GCM` using a local master passphrase.
- **HITL (Human-in-the-Loop) Gate**: Any action that modifies external state (sending an email, pushing code, spending money) or accesses the Vault is intercepted by the Risk Gate. If the risk level is `HIGH`, execution pauses until the user approves via the WebUI or CLI.

### 2.3 Memory: SQLite Vector DB
NEXUS maintains both episodic memory (chat history) and semantic memory. Text is embedded locally using Ollama (`nomic-embed-text`) and stored in a highly optimized SQLite database utilizing vector extensions (`sqlite-vec`).

## 3. Next-Generation Features

### 3.1 Predictive Pre-Computation (Zero-Latency AI)
The `predictive` package runs entirely decoupled from user input. It polls connected integrations (Calendar, GitHub Webhooks). When it detects an upcoming event (e.g., a meeting in 1 hour), it preemptively dispatches a research task to the bus, caching the result. When the user opens the WebUI, the result loads in 0ms.

### 3.2 Auto-Forge (Wasm Hot-Loading)
Instead of requiring Docker restarts to add new tools, NEXUS writes new agent code in TinyGo, compiles it to WebAssembly (locally or via the NEXUS Cloud Compiler), and mounts it into the `wazero` sandbox. The new agent instantly registers its capabilities with the Event Bus.

### 3.3 UI-to-API Reverse Engineering
For legacy systems without an OpenAPI spec, the `ui2api` package drives a headless Chrome browser to interact with the target UI. It intercepts the network traffic, extracts undocumented Bearer tokens into the Vault, and streams the raw HAR logs to an LLM. The LLM writes the API wrapper, which is then fed into Auto-Forge to become a permanent Tool.

### 3.4 Shadow Mode Evolution
The `shadow` package continuously monitors workflows. If a task takes too long or costs too much, NEXUS spins up a hidden background worker using a cheaper model or an optimized prompt. It compares the metrics. If the new method is >10% cheaper or >20% faster without quality degradation, it triggers the HITL gate asking the user to approve the permanent system upgrade.

### 3.5 Hive-Mind Mesh Computing
The `mesh` package allows multiple NEXUS binaries on the same network (e.g., your iPhone, your MacBook, and your Windows gaming PC) to discover each other via mDNS. Tasks are routed based on hardware profiles. A heavy Stable Diffusion request sent from the phone will automatically be forwarded over the mesh to execute on the Windows PC's GPU, returning the image back to the phone seamlessly.
