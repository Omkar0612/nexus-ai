<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=200&section=header&text=NEXUS&fontSize=80&fontColor=fff&animation=fadeIn" width="100%"/>

### The Autonomous AI Agent Built for Production

[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**95% of AI agent pilots fail after the demo.** <br>
NEXUS is engineered for the 5% that ship to production.

[Quick Start](#-quick-start) • [Architecture](#-architecture) • [Production Features](#-production-grade-features) • [Roadmap](ROADMAP.md)

</div>

---

## 🎯 What Makes NEXUS Different

Most AI agent frameworks are research toys. NEXUS is production infrastructure:

- **Observability First** - Complete execution traces, not JSON dumps in Slack
- **Kill Switches** - 3-layer emergency stop with transactional rollback
- **Cost Control** - Real-time token market routing + hallucination loop detection
- **Zero-Trust Security** - Agentic fuzzing + WebAssembly sandboxing
- **Actually Free** - No $200/month "Pro" tier. Works with Ollama, Groq, Gemini.

---

## 🚀 Quick Start

### Option 1: Docker (Recommended)
```bash
docker run -p 7070:7070 \
  -e NEXUS_LLM_PROVIDER=ollama \
  -e NEXUS_LLM_BASE_URL=http://host.docker.internal:11434/v1 \
  ghcr.io/omkar0612/nexus-ai:latest
```
Then open http://localhost:7070

### Option 2: Binary Release
```bash
# Download from https://github.com/Omkar0612/nexus-ai/releases
wget https://github.com/Omkar0612/nexus-ai/releases/latest/download/nexus-linux-amd64
chmod +x nexus-linux-amd64
./nexus-linux-amd64 start
```

### Option 3: Build from Source
```bash
git clone https://github.com/Omkar0612/nexus-ai
cd nexus-ai
CGO_ENABLED=1 go build -o nexus ./cmd/nexus
./nexus start
```

> **Note:** SQLite requires CGO. macOS users need Xcode Command Line Tools (`xcode-select --install`).

---

## 🏗️ Architecture

NEXUS is a modular agent runtime with hot-swappable components:

```
┌─────────────────────────────────────────────────────────┐
│                    CLI / Web UI                         │
├─────────────────────────────────────────────────────────┤
│  Router (LLM Orchestration + Intent Classification)     │
├──────────┬──────────────────────────────────┬───────────┤
│ Observe  │  Kill Switch  │  Circuit Breaker │  Routing  │
│ (Traces) │  (3-Layer)    │  (Per-Tool)      │  (Market) │
├──────────┴──────────────────────────────────┴───────────┤
│                    Agent Registry                        │
│  • calendar  • email  • github  • imagegen  • n8n       │
│  • browser   • voice  • vision  • writing   • music     │
├─────────────────────────────────────────────────────────┤
│            Plugin Layer (WASM + Native)                  │
│  • Auto-Forge (Hot compilation)                          │
│  • Fuzzer (Security testing)                             │
│  • UI-to-API (Reverse engineering)                       │
├─────────────────────────────────────────────────────────┤
│         Infrastructure                                   │
│  • Vault (AES-256 credentials)                           │
│  • Memory (SQLite Vector DB + Liquid Context)            │
│  • Mesh (P2P mDNS discovery)                             │
│  • Scheduler (Cron + Predictive)                         │
└─────────────────────────────────────────────────────────┘
```

**Core Design Principles:**
1. **Fail-Safe by Default** - Every agent action is logged, traced, and rollback-capable
2. **Zero-Trust Execution** - Auto-generated code runs in WASM sandbox with fuzzing
3. **Cost-Aware** - Token market routes to cheapest provider in real-time
4. **Human-in-the-Loop** - Irreversible actions (delete, transfer) require approval

---

## 🛡️ Production-Grade Features

### 1. Observability Stack (`internal/observe`)
**The #1 reason agents fail: teams can't debug them in production.**

```go
// Every agent execution gets a structured trace
trace := tracer.StartTrace(ctx, "research_agent", traceID)
trace.RecordStep(TraceStep{
    Action:     "tool_call",
    ToolName:   "web_search",
    ToolArgs:   map[string]interface{}{"q": "NEXUS AI"},
    ToolOutput: "Found 1,200 results",
    LatencyMs:  340,
    Success:    true,
})
trace.EndTrace(StatusSuccess, tokens, costUSD)
```

**What you get:**
- Timeline reconstruction of every agent decision
- Per-tool latency, retry count, and failure reason
- Hallucination loop detection (same tool called 3x)
- Token/cost tracking with budget alerts
- Cross-functional dashboards (engineers, PMs, domain experts)

### 2. Kill-Switch Architecture (`internal/killswitch`)
**What happens when your agent goes rogue at 3 AM?**

- **Layer 1 (Hard Stop):** Instant credential revocation + queue drain
- **Layer 2 (Soft Pause):** Freeze execution, preserve state for review
- **Layer 3 (Rollback):** Undo last N actions transactionally

**Auto-triggers:**
- Cost threshold breach ($50 spent in 1 hour)
- Hallucination loop detected
- 3 consecutive tool failures
- Manual panic button (CLI: `nexus kill <session-id>`)

**Post-mortem:**
- Classifies failure type (cost, loop, tool, timeout)
- Adds trace to regression test suite
- Updates prompt validators if needed

### 3. Circuit Breakers (`internal/circuit`)
**When Stripe's API goes down, your agent shouldn't retry 500 times.**

```go
breaker.Call(ctx, "stripe_api", func() error {
    return stripe.CreatePayment(payload)
})
// Circuit OPEN after 3 consecutive fails
// Auto-retry after 30s (half-open state)
// Degrades to read-only mode
```

### 4. Token Stock Market (`internal/routing`)
**Never overpay for an API call.**

```
Provider    Model          Cost/1M    Latency    Score
────────────────────────────────────────────────────
Ollama      llama3.2       $0.00      120ms      12   ← SELECTED
Groq        llama-70b      $0.59      80ms       66
Gemini      2.0-flash      $0.15      150ms      30
OpenAI      gpt-4o         $2.50      200ms      270
```

Formula: `(Cost/1M * 100) + (Latency * 0.1)`

### 5. Agentic Fuzzing (`internal/fuzzer`)
**An adversarial agent attacks your agent's code before deployment.**

```go
// Before deploying auto-generated WASM agent
fuzzer.Test(wasmModule, []Attack{
    NullByteInjection,
    SQLInjection,
    PathTraversal,
    MemoryExhaustion,  // 10MB payload
    InfiniteLoop,
})
// Deployment rejected if agent panics
```

### 6. Liquid Context (`internal/memory`)
**Infinite memory without hitting context limits.**

When chat history exceeds 8,000 tokens:
1. Background worker grabs old episodes (>24h)
2. LLM compresses: *"User needs n8n deployment. Prefers Go."*
3. Deletes bloated raw chat (5,000 tokens)
4. Stores dense concept (50 tokens)

**Result:** NEXUS never forgets, never amnesias, never hits API limits.

---

## 🎨 Agent Capabilities

**Built-In Agents:**
- `calendar` - Google Calendar integration (conflicts, free slots)
- `email` - Gmail/Outlook automation
- `github` - PR reviews, issue tracking, CI/CD monitoring
- `imagegen` - Stable Diffusion, FLUX, DALL-E
- `voice` - Text-to-speech (Coqui, ElevenLabs, system TTS)
- `vision` - OCR, image analysis
- `writing` - Draft, rewrite, translate, proofread
- `music` - AudioCraft, MusicGen
- `browser` - Headless web automation, UI-to-API reverse engineering
- `n8n` - Natural language to workflow DAG compiler

**Plugin System:**
- Hot-load WASM agents at runtime
- Zero Docker restarts
- Natural language agent creation: *"Create a Real Estate scraper"*

---

## 🔒 Security Model

1. **WebAssembly Sandbox** - Auto-generated code runs in WASM with no filesystem/network access
2. **Agentic Fuzzing** - Every generated agent is attacked before deployment
3. **AES-256 Vault** - Credentials encrypted at rest
4. **Human-in-the-Loop** - Irreversible actions require manual approval
5. **Audit Logs** - Every agent action is logged for compliance

---

## 🆚 NEXUS vs. Alternatives

| Feature | NEXUS | AutoGPT | LangChain | CrewAI |
|---------|-------|---------|-----------|--------|
| **Production Observability** | ✅ Full traces | ❌ | ⚠️ Paid | ❌ |
| **Kill Switch + Rollback** | ✅ 3-layer | ❌ | ❌ | ❌ |
| **Circuit Breakers** | ✅ | ❌ | ❌ | ❌ |
| **Hallucination Detection** | ✅ Auto | ❌ | ❌ | ❌ |
| **Cost Arbitrage** | ✅ Real-time | ❌ | ❌ | ❌ |
| **WASM Sandbox** | ✅ | ❌ | ❌ | ❌ |
| **100% Free** | ✅ | ✅ | ⚠️ Freemium | ⚠️ Freemium |
| **Self-Hosted** | ✅ | ✅ | ✅ | ✅ |

---

## 🗺️ Roadmap

**v2.0 (Current)**
- ✅ Production observability
- ✅ Kill-switch architecture
- ✅ Circuit breakers
- ✅ Token market routing
- ✅ Liquid context
- ✅ WebAssembly sandbox

**v2.1 (Q2 2026)**
- 🔄 Mesh P2P GPU sharing
- 🔄 Shadow mode self-evolution
- 🔄 Predictive pre-computation
- 🔄 n8n DAG compiler

**v2.2 (Q3 2026)**
- 📋 Multi-agent orchestration
- 📋 Distributed tracing (OpenTelemetry)
- 📋 Desktop app (Wails)
- 📋 Mobile app (React Native)

See [ROADMAP.md](ROADMAP.md) for full details.

---

## 📚 Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Production Checklist](PRODUCTION_CHECKLIST.md)
- [Plugin Development](docs/PLUGIN_DEVELOPMENT.md)
- [API Reference](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- How to add a new agent
- Code style guide
- Testing requirements
- PR process

---

## 🆓 Free LLM Providers

| Provider | Model | Speed | Cost | Setup Time |
|----------|-------|-------|------|------------|
| **Ollama** | Any model | Local GPU | Free | 2 min |
| **Groq** | Llama 3.3 70B | 300 tok/s | Free | 60 sec |
| **Gemini** | 2.0 Flash | Fast | Free (1M tok/day) | 2 min |
| **Together** | FLUX / Mixtral | Fast | Free ($25 credits) | 3 min |

---

## 📜 License

MIT © 2026 Omkar Parab

Free forever. No "Pro" tier. No bait-and-switch.

---

<div align="center">

**If NEXUS saved you from a production incident, a ⭐ means a lot.**

Built for the 5% that actually ship to production.

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=100&section=footer" width="100%"/>

</div>
