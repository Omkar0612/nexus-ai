# 🧠 NEXUS — The AI Agent That Actually Works

[![Stars](https://img.shields.io/github/stars/Omkar0612/nexus-ai?style=social)](https://github.com/Omkar0612/nexus-ai/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Free](https://img.shields.io/badge/cost-100%25%20free-brightgreen)](https://github.com/Omkar0612/nexus-ai)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev)

> I analyzed 500+ Reddit complaints about AI agents and built fixes for every single one.

**NEXUS is the first open-source AI agent with:**
- 🔍 **Drift Detection** — spots stalled work before it fails
- 🏥 **Self-Healing** — auto-diagnoses and fixes broken tasks
- 🎭 **Emotional Intelligence** — adapts tone based on how stressed you sound
- 🎯 **Goal Tracking** — warns when you're off-track from your goals
- 👋 **Session Briefing** — briefs you when you return after being away
- 📈 **Adaptive Learning** — gets smarter the more you use it
- 🔐 **Privacy Vault** — AES-256 encrypted secrets, auto-redacted from LLM prompts
- 🎭 **Persona Engine** — switch between work/creative/client/focus modes
- 📴 **Offline Mode** — works fully without internet, queues tasks automatically

---

## ⚡ Quick Start (2 minutes)

```bash
# Clone and build
git clone https://github.com/Omkar0612/nexus-ai
cd nexus-ai
make build

# Configure (add your free Groq API key from console.groq.com)
cp config/nexus.example.toml ~/.nexus/nexus.toml

# Start
nexus start

# Chat
nexus chat
```

---

## 🆓 100% Free Forever

| Provider | Model | Free Tier |
|---|---|---|
| **Groq** | Llama 3.3 70B | 300+ tok/sec, free |
| **Gemini** | 2.0 Flash | 1M tokens/day free |
| **OpenRouter** | Multiple | Free model tier |
| **Ollama** | Any | Unlimited local |
| **Together AI** | Multiple | $25 free credits |

---

## 🧠 Features That Never Existed Before

### 1. Drift Detector
```bash
nexus drift
# Output:
# 🔴 [HIGH] Task appears stalled: 'building the webhook handler' (last touched 2 days ago)
#    💡 Resume or close: 'building the webhook handler'
# 🟡 [MEDIUM] Follow-up may have been missed (mentioned 3 days ago)
#    💡 Did you follow up on: 'ping the client about the invoice'?
```

### 2. Self-Healing
```bash
# When a task fails, NEXUS auto-diagnoses:
# ⚠️ Task 'daily-briefing' failed (attempt 1/3)
# ROOT CAUSE: Groq API rate limit exceeded at 06:00 UTC
# FIX: Switching to Gemini Flash fallback. Retrying in 30s...
# ✅ Task recovered successfully.
```

### 3. Emotional Intelligence
```bash
# You type: "this is STILL not working ugh"
# NEXUS detects: frustrated + stressed
# Response: empathetic, brief, solution-first
# "I can see this has been frustrating. Here's the fix: [direct answer]"
```

### 4. Persona Engine
```bash
nexus persona use work      # formal, full tools, code-heavy
nexus persona use creative  # brainstorming mode
nexus persona use client    # professional, no internal data exposed
nexus persona use focus     # max 200 word responses, zero fluff
nexus persona use research  # deep, cited, thorough
nexus persona create my-mode --prompt "Always respond in bullet points"
```

### 5. Privacy Vault
```bash
nexus vault store GROQ_API_KEY gsk_xxxxx --zone business
nexus vault store PERSONAL_NOTE "My strategy for Q2" --zone personal
nexus vault list
# Secrets are AES-256 encrypted and NEVER sent to any LLM
```

### 6. Offline Mode
```bash
# Lose internet? NEXUS auto-switches to local Ollama
# All tasks queue up and run when you're back online
nexus status
# 📴 Offline mode active (Ollama). 3 tasks queued for sync.
```

---

## 🤖 AI Coworker Mode

NEXUS chains multiple agents automatically:

```bash
nexus chat
> research the top 5 AI startups from YC 2026, analyze their pricing,
  write a competitive analysis and save it as report.md

# NEXUS automatically:
# [1/4] 🔍 Research Agent → fetching YC 2026 batch data
# [2/4] 📊 Data Agent    → analyzing pricing models
# [3/4] ✍️  Writer Agent  → drafting competitive analysis
# [4/4] 💾 File Agent    → saving report.md
# ✅ Done in 47s
```

---

## 🔌 Connect to Anything

```bash
# n8n (2000+ integrations)
nexus skill install n8n-bridge

# MCP Protocol (GitHub, Postgres, Slack, Google Maps...)
# Add to nexus.toml:
# [[mcp.servers]]
# name = "github"
# command = "npx @modelcontextprotocol/server-github"

# Free APIs (zero keys needed)
nexus skill install free-apis
# Includes: weather, Wikipedia, crypto, HackerNews,
#           currency, IP geo, dictionary, Reddit, GitHub
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────┐
│                  NEXUS CLUSTER                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Node 1  │  │  Node 2  │  │  Node 3  │      │
│  │ :7701    │  │ :7702    │  │ :7703    │      │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘      │
│       └─────────────┼─────────────┘            │
│              ┌──────┴──────┐                   │
│              │Load Balancer│                   │
│              │    :7700    │                   │
│              └─────────────┘                   │
│  Gateways: CLI | Telegram | Discord | Web API  │
│  Memory:   SQLite (episodic + semantic)         │
│  Vault:    AES-256-GCM encrypted SQLite         │
│  Workers:  Python (research/browser/vision)     │
└─────────────────────────────────────────────────┘
```

---

## 🐳 One-Command Cluster

```bash
docker compose up -d
# Starts: 3 NEXUS nodes + Python workers + Ollama + n8n
# Load balancer auto-removes dead nodes
# Health checks every 10s
```

---

## 📊 vs Other Agents

| Feature | NEXUS | OpenClaw | n8n AI | AutoGPT |
|---|---|---|---|---|
| Drift Detection | ✅ | ❌ | ❌ | ❌ |
| Self-Healing | ✅ | ❌ | ❌ | ❌ |
| Emotional Intelligence | ✅ | ❌ | ❌ | ❌ |
| Goal Tracking | ✅ | ❌ | ❌ | ⚠️ |
| Privacy Vault | ✅ | ❌ | ❌ | ❌ |
| Offline Mode | ✅ | ❌ | ❌ | ❌ |
| Persona Engine | ✅ | ❌ | ❌ | ❌ |
| Session Briefing | ✅ | ❌ | ❌ | ❌ |
| 100% Free | ✅ | ⚠️ | ⚠️ | ⚠️ |
| Load Balanced | ✅ | ❌ | ✅ | ❌ |

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). We especially want:
- New skills (`.toml` manifest + Python worker)
- New free API integrations
- New use case examples

## ⭐ Star History

If NEXUS saved you time, please star the repo!

[![Star History Chart](https://api.star-history.com/svg?repos=Omkar0612/nexus-ai&type=Date)](https://star-history.com/#Omkar0612/nexus-ai)

---

## 📄 License

MIT — free forever, use it however you want.
