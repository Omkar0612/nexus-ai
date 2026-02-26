# 🧠 NEXUS — The AI Agent That Actually Works

[![CI](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml)
[![Stars](https://img.shields.io/github/stars/Omkar0612/nexus-ai?style=social)](https://github.com/Omkar0612/nexus-ai/stargazers)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Free](https://img.shields.io/badge/cost-100%25%20free-brightgreen)](https://github.com/Omkar0612/nexus-ai)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev)

> I analysed 500+ Reddit complaints about AI agents and built fixes for every single one.

**NEXUS is the most capable open-source AI agent ever built:**

| v1.0–1.2 | v1.3 | v1.4 |
|---|---|---|
| Drift Detection, Self-Healing, Emotional Intelligence, Goal Tracking, Session Briefing, Adaptive Learning, Privacy Vault, Persona Engine, Offline Mode, Cluster/Load Balancer | Multi-Agent Bus, Daily Digest, HITL Gate, Voice Interface, Browser Agent | Analytics Dashboard, Phone Agent, Email Agent, Notes Agent, GitHub Agent, Telegram Companion |

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

### 📱 Use NEXUS on your phone via Telegram

```bash
# Add to nexus.toml:
# [telegram]
# token = "your-bot-token"
# allowed_user_ids = [your-telegram-id]
# admin_chat_id = your-telegram-id

nexus telegram start
# Now control NEXUS from anywhere — no app install needed
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

## 🧠 Core Features (v1.0–1.2)

### Drift Detector
```bash
nexus drift
# 🔴 [HIGH] Task appears stalled: 'building the webhook handler' (last touched 2 days ago)
#    💡 Resume or close: 'building the webhook handler'
# 🟡 [MEDIUM] Follow-up may have been missed (mentioned 3 days ago)
#    💡 Did you follow up on: 'ping the client about the invoice'?
```

### Self-Healing
```bash
# When a task fails, NEXUS auto-diagnoses:
# ⚠️  Task 'daily-briefing' failed (attempt 1/3)
# ROOT CAUSE: Groq API rate limit exceeded at 06:00 UTC
# FIX: Switching to Gemini Flash fallback. Retrying in 30s...
# ✅ Task recovered successfully.
```

### Emotional Intelligence
```bash
# You type: "this is STILL not working ugh"
# NEXUS detects: frustrated + stressed
# Response: empathetic, brief, solution-first
# "I can see this has been frustrating. Here's the fix: [direct answer]"
```

### Persona Engine
```bash
nexus persona use work      # formal, full tools, code-heavy
nexus persona use creative  # brainstorming mode
nexus persona use client    # professional, no internal data exposed
nexus persona use focus     # max 200 word responses, zero fluff
nexus persona use research  # deep, cited, thorough
nexus persona create my-mode --prompt "Always respond in bullet points"
```

### Privacy Vault
```bash
nexus vault store GROQ_API_KEY gsk_xxxxx --zone business
nexus vault store PERSONAL_NOTE "My strategy for Q2" --zone personal
nexus vault list
# Secrets are AES-256 encrypted and NEVER sent to any LLM
```

### Offline Mode
```bash
# Lose internet? NEXUS auto-switches to local Ollama
# All tasks queue and execute when you're back online
nexus status
# 📴 Offline mode active (Ollama). 3 tasks queued for sync.
```

---

## 🤖 Multi-Agent System (v1.3)

### Multi-Agent Bus
NEXUS spawns and coordinates typed sub-agents over a central message bus:

```bash
nexus chat
> research the top 5 AI startups from YC 2026, analyze their pricing,
  write a competitive analysis and save it as report.md

# NEXUS automatically routes across agents:
# [1/4] 🔍 Researcher Agent → fetching YC 2026 batch data
# [2/4] 📊 Analyst Agent    → analyzing pricing models
# [3/4] ✍️  Writer Agent     → drafting competitive analysis
# [4/4] 💾 File Agent       → saving report.md
# ✅ Done in 47s
```

Available agent roles: `researcher` · `coder` · `writer` · `analyst` · `reviewer`

### Human-in-the-Loop (HITL) Gate
Every action is risk-classified before execution:

```
🟢 low risk    → auto-executes silently
🟡 medium risk → executes with audit log entry
🔴 high risk   → pauses, sends Telegram approval request
🛑 emergency  → nexus lock  (blocks all medium/high actions instantly)
```

```bash
nexus lock    # engage emergency lock
nexus unlock  # release
```

### Voice Interface
```bash
nexus voice start
# 🎤 Listening... (Whisper transcription, fully offline)
# Speak your command — NEXUS replies via TTS
# Supports ElevenLabs, piper (local), or silent mode
```

### Browser Agent
```bash
nexus browse "go to github.com/trending and extract the top 10 repos"
# 🌐 Navigating → github.com/trending
# 📸 Extracting content...
# ✅ Found 10 repos. Injecting into context.
# Safety: URL allowlist, depth limit, loop detection built-in
```

### Daily Digest
```bash
# Delivered every morning automatically:
nexus digest
# 🌅 Good morning, Omkar.
# 📈 Goals on track: 3/4
# ⚠️  Drift signals: 1 stalled task
# 💰 LLM spend yesterday: $0.00 (free tier)
# 📚 KB highlights: 2 new notes indexed
```

---

## 📊 Analytics & Integrations (v1.4)

### Analytics Dashboard
```bash
nexus dashboard
# Web UI at http://localhost:7700/dashboard
# Shows: cost over time, agent stats, goal progress,
#        audit trail, drift history, KB usage
```

### Phone Agent (Twilio)
```bash
nexus phone call +971501234567 --message "Your meeting is in 10 minutes"
nexus phone sms  +971501234567 --message "Task complete: report.md saved"
# Inbound calls auto-routed to NEXUS voice pipeline
```

### Email Agent (IMAP/SMTP)
```bash
nexus email read     # fetch + classify inbox
nexus email reply 42 # draft + send reply via LLM
nexus email rules    # view auto-responder rules
# Secrets auto-redacted before any LLM processing
```

### Notes Agent
```bash
nexus notes capture  # auto-capture meeting notes from voice/text
nexus notes search "Q2 strategy"
nexus notes export meeting-2026-02-26 --format markdown
# Action items extracted automatically
```

### GitHub Agent
```bash
nexus github issue create --repo myorg/myrepo --title "Bug: login fails"
nexus github pr review 42
nexus github branch create feature/new-thing
# Destructive operations (delete, force-push) require HITL approval
```

### Telegram Mobile Companion
```
📱 Full NEXUS from your phone:
  /chat    — chat with NEXUS
  /drift   — check stalled tasks
  /digest  — morning briefing on demand
  /vault   — retrieve secrets
  /approve — approve/reject high-risk actions
  + inline keyboard for quick actions
  + voice message support
```

---

## 🏗️ Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                    NEXUS CLUSTER                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │  Node 1  │  │  Node 2  │  │  Node 3  │            │
│  │  :7701   │  │  :7702   │  │  :7703   │            │
│  └────┬────┘  └────┬────┘  └────┬────┘            │
│       └─────────────┤─────────────┘                   │
│               ┌─────┤─────┐                          │
│               │ Load Balancer │                          │
│               │    :7700      │                          │
│               └──────────────┘                          │
│                                                          │
│  Gateways:  CLI │ Telegram │ Discord │ Web API          │
│  Agents:    Researcher │ Coder │ Writer │ Analyst │ Reviewer │
│  Memory:    SQLite (episodic + semantic + KB)             │
│  Vault:     AES-256-GCM encrypted SQLite                 │
│  Workers:   Python (research / browser / vision)         │
│  Integrations: Twilio │ IMAP/SMTP │ GitHub │ n8n         │
└───────────────────────────────────────────────────────────────┘
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

## 📊 NEXUS vs Other Agents

| Feature | NEXUS | OpenClaw | n8n AI | AutoGPT |
|---|:---:|:---:|:---:|:---:|
| Drift Detection | ✅ | ❌ | ❌ | ❌ |
| Self-Healing | ✅ | ❌ | ❌ | ❌ |
| Emotional Intelligence | ✅ | ❌ | ❌ | ❌ |
| Goal Tracking | ✅ | ❌ | ❌ | ⚠️ |
| Privacy Vault | ✅ | ❌ | ❌ | ❌ |
| Offline Mode | ✅ | ❌ | ❌ | ❌ |
| Persona Engine | ✅ | ❌ | ❌ | ❌ |
| Session Briefing | ✅ | ❌ | ❌ | ❌ |
| Multi-Agent Bus | ✅ | ❌ | ⚠️ | ⚠️ |
| HITL Gate | ✅ | ❌ | ⚠️ | ⚠️ |
| Voice Interface | ✅ | ❌ | ❌ | ❌ |
| Browser Agent | ✅ | ❌ | ❌ | ✅ |
| Daily Digest | ✅ | ❌ | ❌ | ❌ |
| Analytics Dashboard | ✅ | ❌ | ⚠️ | ❌ |
| Phone / SMS Agent | ✅ | ❌ | ⚠️ | ❌ |
| Email Agent | ✅ | ❌ | ⚠️ | ❌ |
| Notes Agent | ✅ | ❌ | ❌ | ❌ |
| GitHub Agent | ✅ | ❌ | ❌ | ❌ |
| Telegram Companion | ✅ | ❌ | ❌ | ❌ |
| Load Balanced Cluster | ✅ | ❌ | ✅ | ❌ |
| 100% Free | ✅ | ⚠️ | ⚠️ | ⚠️ |

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
