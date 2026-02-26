<div align="center">

<img src="https://capsule-render.vercel.app/api?type=venom&color=gradient&customColorList=6,11,20&height=280&section=header&text=NEXUS%20AI&fontSize=90&fontColor=fff&animation=fadeIn&fontAlignY=45&desc=The%20AI%20agent%20that%20actually%20works.%20Free%20forever.&descAlignY=65&descSize=18" width="100%"/>

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=22&duration=3000&pause=1000&color=7C3AED&center=true&vCenter=true&multiline=true&repeat=true&width=700&height=60&lines=Self-healing+%E2%80%A2+Drift-aware+%E2%80%A2+100%25+Free;Web+UI+%E2%80%A2+CLI+%E2%80%A2+Telegram+%E2%80%A2+Voice+%E2%80%A2+Email;Multi-agent+%E2%80%A2+Offline+%E2%80%A2+AES-256+Vault" alt="Typing SVG" />

<br/><br/>

[![CI](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml)
[![Stars](https://img.shields.io/github/stars/Omkar0612/nexus-ai?style=for-the-badge&logo=github&color=FFD700&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai/stargazers)
[![Forks](https://img.shields.io/github/forks/Omkar0612/nexus-ai?style=for-the-badge&logo=github&color=4ade80&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai/network/members)
[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a2e)](https://go.dev)
[![MIT](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge&labelColor=1a1a2e)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-7c3aed?style=for-the-badge&labelColor=1a1a2e)](CONTRIBUTING.md)
[![100% Free](https://img.shields.io/badge/Cost-Zero-f59e0b?style=for-the-badge&logo=opensourceinitiative&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai)

<br/>

> **I analysed 500+ Reddit complaints about AI agents and built a fix for every single one.**

<br/>

[🚀 Quick Start](#-quick-start) · [✨ Features](#-features) · [🌐 Web UI](#-web-ui-new-in-v16) · [🆓 Free LLMs](#-free-llm-providers) · [📺 Demo](#-demo) · [📋 Changelog](#-changelog) · [🔮 Roadmap](ROADMAP.md) · [🤝 Contribute](#-built-by-the-community)

</div>

---

## 🚀 Quick Start

```bash
# 1. Clone & build
git clone https://github.com/Omkar0612/nexus-ai
cd nexus-ai
CGO_ENABLED=1 go build -tags ci ./...

# 2. Add your free API key (console.groq.com — 60 sec signup)
cp config/nexus.example.toml ~/.nexus/nexus.toml

# 3. Run — Web UI opens at http://localhost:7070
nexus start
```

> 🆓 **No paid API needed.** Works with Groq (free), Gemini (free), Ollama (local), OpenRouter (free tier).

<details>
<summary><b>🌐 Web UI (new in v1.6) &rarr;</b></summary>

```bash
# Start with Web UI (default: http://localhost:7070)
nexus start

# Custom port
nexus start --webui-addr :8080

# CLI only, no Web UI
nexus start --no-webui

# Debug mode
nexus start --debug
```

The Web UI provides:
- 💬 **Live chat** — stream responses word by word from your local LLM
- 📡 **Agent activity feed** — real-time SSE stream of what every agent is doing
- 🏥 **Health endpoint** — `GET /api/health` → `{"status":"ok","version":"1.6"}`

</details>

<details>
<summary><b>🐳 Full cluster (Docker) &rarr;</b></summary>

```bash
docker compose up -d
# 3 load-balanced nodes + Ollama + n8n
# Health checks every 10s. Dead nodes auto-removed.
```

</details>

<details>
<summary><b>📱 Telegram companion &rarr;</b></summary>

```toml
# ~/.nexus/nexus.toml
[telegram]
token            = "your-bot-token"
allowed_user_ids = [your-telegram-id]
admin_chat_id    = your-telegram-id
```

```bash
nexus telegram start
```

```
/chat    — talk to NEXUS from anywhere
/drift   — what's stalled right now
/digest  — your morning brief
/approve — approve a high-risk action
/vault   — retrieve a secret
```

</details>

---

## 🌐 Web UI — New in v1.6

```
┌─────────────────────────────────────────────────────────────────┐
│  NEXUS AI  v1.6                              ● agents: 3 active │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  You: research top 5 AI tools launched this week         │   │
│  │                                                          │   │
│  │  NEXUS: Here are the top 5 AI tools launched this        │   │
│  │  week, analysed and ranked by practical utility...       │   │
│  │  ▌                                               (live)  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  📡 Agent Activity                                               │
│  ● researcher  running   — fetching latest ProductHunt data      │
│  ✓ analyst     done      — pricing comparison complete           │
│  ✓ writer      done      — summary generated                     │
│                                                                  │
│  [ Type a message...                              ] [ Send ▶ ]  │
└─────────────────────────────────────────────────────────────────┘
```

**Stack:** `net/http` stdlib · `//go:embed` static assets · SSE event stream · zero JS frameworks

**API endpoints:**

| Method | Endpoint | Description |
|:---:|:---|:---|
| `POST` | `/api/chat` | Stream LLM response as SSE |
| `GET` | `/api/events` | Live agent activity stream |
| `GET` | `/api/health` | Health check + version |
| `GET` | `/` | Embedded dark-mode UI |

---

## 📺 Demo

```
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║  $ nexus start                                               ║
║                                                              ║
║    NEXUS AI v1.6                                             ║
║    Gateway : 127.0.0.1:7700                                  ║
║    Web UI  : http://localhost:7070   ← open in browser       ║
║                                                              ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  $ nexus chat                                                ║
║                                                              ║
║  > research YC 2026 AI startups, analyse their pricing,      ║
║    write a competitive brief, save as report.md              ║
║                                                              ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  🔍  Researcher  ──▶  fetching 14 YC profiles       [done]  ║
║  📊  Analyst     ──▶  comparing pricing tiers       [done]  ║
║  ✍️   Writer      ──▶  drafting executive summary   [done]  ║
║  💾  File Agent  ──▶  writing report.md             [done]  ║
║                                                              ║
╠══════════════════════════════════════════════════════════════╣
║  ✅  Done in 47 seconds.  Cost: $0.00                        ║
╚══════════════════════════════════════════════════════════════╝
```

```
╔═══════════════════════════════════╗
║  🌅  Good morning, Omkar.         ║
║                                   ║
║  📈  Goals on track    3 / 4      ║
║  ⚠️   Stalled tasks      1        ║
║  💰  LLM cost today    $0.00      ║
║  📚  New KB entries     2         ║
║  📧  Unread emails      5         ║
╚═══════════════════════════════════╝

  💡 "nexus-api-refactor" stalled 2 days.
     Resume or close it?
```

---

## ✨ Features

<div align="center">
<img src="https://capsule-render.vercel.app/api?type=rect&color=gradient&customColorList=6,11,20&height=3&section=header" width="100%"/>
</div>

<br/>

### 🌐 Web UI Server
> *Chat with NEXUS in your browser. Zero JS frameworks. Fully embedded.*

A complete `net/http` server ships inside the binary via `//go:embed`. Open `http://localhost:7070` after `nexus start` — no separate frontend server, no Node.js, no build step.

```bash
# Stream a response from the API directly
curl -X POST http://localhost:7070/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message": "summarise my tasks for today"}'

# Watch live agent activity
curl http://localhost:7070/api/events
# data: {"agent":"researcher","status":"running","message":"fetching calendar"}
# data: {"agent":"researcher","status":"done"}
```

---

### 🔍 Drift Detector
> *The only AI agent that notices when your own work is stalling.*

Runs silently in the background. Monitors your conversation history for stalled tasks, missed follow-ups, and repetitive failures. Fires smart nudges — not spam.

```
🔴 [HIGH]   'nexus-api-refactor' stalled — last touched 2 days ago
             💡 Resume or close this task?

🟡 [MEDIUM] Follow-up missed — 'ping client about invoice' (3 days ago)
             💡 Did you follow up?

🔴 [HIGH]   Same error mentioned 3× — 'connection refused'
             💡 Let me help you solve this systematically
```

---

### 🏥 Self-Healing Engine
> *Fails once. Never twice.*

When a task fails, NEXUS diagnoses the root cause, switches to a backup LLM provider, and retries — all without you touching a thing.

```
⚠️  Task 'daily-briefing' failed (attempt 1/3)
    ROOT CAUSE: Groq rate limit hit at 06:00 UTC
    FIX: Switching to Gemini 2.0 Flash...
         Retrying in 30s...
✅  Task recovered successfully.
```

---

### 🛡️ Human-in-the-Loop Gate
> *Smart trust, not blind execution.*

Every action is risk-scored before execution. Low risk runs silently. Medium risk is logged. High risk — NEXUS asks you first.

| Risk Level | Actions | Behaviour |
|:---:|:---|:---|
| 🟢 Low | Read file, Search web, Chat | Silent execute |
| 🟡 Medium | Write file, Send message | Execute + audit log |
| 🔴 High | Delete data, Push to GitHub, Make call | Pause → ask you |

---

### 🧠 Memory & Knowledge Base
> *Remembers everything. Forgets nothing.*

- **Episodic memory** — full conversation history in local SQLite
- **Knowledge base** — ingest docs, PDFs, URLs → semantic search
- **Working memory** — in-session context window management
- **Privacy Vault** — AES-256-GCM encrypted secrets, never sent to an LLM

---

### 🤖 Multi-Agent Bus
> *The right agent for every job.*

NEXUS automatically decomposes your request and routes it across specialised agents in parallel:

```mermaid
graph LR
    YOU([🧑 You]) --> HITL

    HITL{🛡️ Risk Gate}
    HITL -->|low| AUTO[✅ Auto]
    HITL -->|medium| AUDIT[📋 Audit]
    HITL -->|high| ASK[📱 Ask]

    AUTO & AUDIT & ASK --> BUS

    subgraph BUS["🤖 Agent Bus"]
        direction TB
        R[🔍 Researcher]
        C[💻 Coder]
        W[✍️ Writer]
        A[📊 Analyst]
        V[👁️ Reviewer]
    end

    BUS --> MEM[(🧠 Memory)]
    BUS --> OUT([📤 Result → You])
```

---

### 🎭 Emotional Intelligence
> *Talks to you like a person, not a robot.*

Detects frustration, urgency, or stress in your messages and adjusts its tone in real-time.

```
You:    "this is STILL not working ugh"
NEXUS:  detects → frustrated + stressed
        responds → empathetic, direct, solution-first
        "Here's the exact fix: [answer]"
        (no preamble, no filler)
```

---

### 🎤 Voice Interface
> *Fully offline. No cloud. No subscription.*

OpenAI Whisper runs locally for speech-to-text. Local TTS for responses. Works on a plane.

```bash
nexus voice start
# 🎤 Listening...
# "schedule a briefing for tomorrow morning"
# ✅ Scheduled: daily-briefing at 07:00
```

---

### 📴 Offline Mode
> *Loses internet. Keeps working.*

Detects connectivity loss. Switches to local Ollama. Queues outbound tasks. Syncs everything when you're back online.

---

### 🌐 Integrations

<div align="center">

| Integration | What NEXUS can do |
|:---:|:---|
| 🌐 **Web UI** | Chat, agent activity feed, health API — all at `localhost:7070` |
| 📧 **Email** (IMAP/SMTP) | Read, classify, summarise, reply |
| 🐙 **GitHub** | Create issues, open PRs, review code — HITL on destructive ops |
| 📞 **Phone** (Twilio) | Call or SMS on your behalf |
| 📱 **Telegram** | Full remote control from your phone |
| 🌐 **Browser** | Fetch and read any webpage, inject into context |
| ⚙️ **n8n** | Trigger any n8n workflow as a NEXUS skill |
| 📝 **Notes** | Capture, search, export meeting notes |
| 📊 **Dashboard** | Live analytics at `localhost:7700` |

</div>

---

## 🆓 Free LLM Providers

<div align="center">

| Provider | Model | Speed | Daily Limit | Cost |
|:---:|:---:|:---:|:---:|:---:|
| ![Groq](https://img.shields.io/badge/Groq-F55036?style=flat-square&logo=groq&logoColor=white) | Llama 3.3 70B | ⚡ 300 tok/s | Unlimited | ![free](https://img.shields.io/badge/-free-22c55e?style=flat-square) |
| ![Gemini](https://img.shields.io/badge/Gemini-4285F4?style=flat-square&logo=google&logoColor=white) | 2.0 Flash | ⚡ Fast | 1M tokens/day | ![free](https://img.shields.io/badge/-free-22c55e?style=flat-square) |
| ![Ollama](https://img.shields.io/badge/Ollama-000000?style=flat-square&logoColor=white) | Any model | 🖥️ Local | Unlimited | ![free](https://img.shields.io/badge/-free-22c55e?style=flat-square) |
| ![OpenRouter](https://img.shields.io/badge/OpenRouter-6C47FF?style=flat-square&logoColor=white) | Multiple | ⚡ Fast | Free tier | ![free tier](https://img.shields.io/badge/-free%20tier-22c55e?style=flat-square) |
| ![Together](https://img.shields.io/badge/Together_AI-FF6B6B?style=flat-square&logoColor=white) | Multiple | ⚡ Fast | $25 credits | ![credits](https://img.shields.io/badge/-\$25%20credits-f59e0b?style=flat-square) |

</div>

---

## ⚔️ NEXUS vs The World

<div align="center">

| Capability | NEXUS | AutoGPT | CrewAI | n8n AI | LangChain |
|:---|:---:|:---:|:---:|:---:|:---:|
| Fixes its own failures | ✅ | ❌ | ❌ | ❌ | ❌ |
| Detects stalled work | ✅ | ❌ | ❌ | ❌ | ❌ |
| Risk gate before every action | ✅ | ⚠️ | ⚠️ | ⚠️ | ❌ |
| Works fully offline | ✅ | ❌ | ❌ | ❌ | ❌ |
| AES-256 local secrets vault | ✅ | ❌ | ❌ | ❌ | ❌ |
| Browser-based Web UI | ✅ | ❌ | ❌ | ✅ | ❌ |
| Reads your email | ✅ | ❌ | ⚠️ | ⚠️ | ❌ |
| Calls your phone | ✅ | ❌ | ❌ | ⚠️ | ❌ |
| Reads your tone | ✅ | ❌ | ❌ | ❌ | ❌ |
| 3-node load-balanced cluster | ✅ | ❌ | ❌ | ✅ | ❌ |
| Morning briefing | ✅ | ❌ | ❌ | ❌ | ❌ |
| 100% free to run | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ |
| Written in Go (fast + single binary) | ✅ | ❌ | ❌ | ❌ | ❌ |

</div>

---

## 🔮 Roadmap — Replacing $289/mo of paid tools

<div align="center">

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=17&duration=3500&pause=1200&color=F59E0B&center=true&vCenter=true&width=700&lines=ChatGPT+%2B+Notion+%2B+Grammarly+%2B+Midjourney+%2B+Cursor...;One+binary.+Zero+cost.+Forever.;%F0%9F%91%89+See+the+full+roadmap" alt="roadmap typing" />

</div>

<br/>

```
v1.5  ██████████████████░░  90%   Plugin SDK · Calendar Agent · Vision · Semantic Memory
v1.6  ████████████████████ 100%  ✅ Web UI · SSE Agent Feed · Desktop Scaffold · CI fixed
v1.7  ░░░░░░░░░░░░░░░░░░░░   0%   Image · Voice Synthesis · Video · Music · Writing Studio
v1.8  ░░░░░░░░░░░░░░░░░░░░   0%   Code Copilot (LSP) · DevOps Agent · Code Search
v1.9  ░░░░░░░░░░░░░░░░░░░░   0%   Finance · Health · CRM · Learning · Travel Agents
v2.0  ░░░░░░░░░░░░░░░░░░░░   0%   NEXUS OS — AI layer for everything
```

<div align="center">

**[📖 Read the full ROADMAP →](ROADMAP.md)**

</div>

---

## 📋 Changelog

<details open>
<summary><b>v1.6 — Web UI · Desktop Scaffold · CI Hardening</b> &nbsp;🆕</summary>

<br/>

- 🌐 **Web UI Server** — dark-mode chat interface embedded in the binary (`//go:embed`), zero external deps
- 💬 **Streaming Chat** — `POST /api/chat` streams LLM response word-by-word as Server-Sent Events
- 📡 **SSE Agent Activity Bus** — `GET /api/events` streams live agent state changes to any connected client
- 🖥️ **Desktop Scaffold** — system tray + global hotkey (`Ctrl+Shift+Space`) + clipboard monitor stub (v1.7)
- 🔧 **CLI flags** — `nexus start --webui-addr :7070`, `--no-webui`, `--debug`
- 🏗️ **CI hardened** — `-tags ci` build strategy, desktop stub for headless runners, clean dep chain
- 🐛 **Bug fixes** — `TestDetectConflicts` event overlap, `TestSSEHubPublish` ping skip, `zerolog` import

</details>

<details>
<summary><b>v1.5 — Calendar · Vision · Semantic Memory · Plugin SDK</b></summary>

<br/>

- 📅 **Calendar Agent** — natural language scheduling, conflict detection, free-slot finder, morning digest
- 👁️ **Vision Agent** — describe images, extract text (OCR), detect objects via local model
- 🧠 **Semantic Memory** — cosine similarity search across all past conversations and KB entries
- 🔌 **Plugin SDK** — build and register new NEXUS skills as Go plugins

</details>

<details>
<summary><b>v1.4 — Analytics · Phone · Email · Notes · GitHub Agent</b></summary>

<br/>

- 📊 **Analytics Dashboard** — live cost, agent activity, goal progress at `localhost:7700`
- 📞 **Phone Agent** — call or SMS via Twilio with HITL confirmation
- 📧 **Email Agent** — read, classify, summarise and reply via IMAP/SMTP
- 📝 **Notes Agent** — capture, tag, search and export meeting notes
- 🐙 **GitHub Agent** — issues, PRs, branches; HITL gate on all destructive ops
- 📱 **Telegram Companion** — full remote control from your phone
- 🐛 **Hallucination Detector** — cross-checks LLM outputs before presenting results
- 🔁 **Loop Detector** — breaks infinite agent loops automatically
- 🧠 **Adaptive Learner** — personalises prompts based on your interaction patterns

</details>

<details>
<summary><b>v1.3 — Multi-Agent Bus · Daily Digest · Voice · Browser</b></summary>

<br/>

- 🤖 **Multi-Agent Bus** — Researcher, Coder, Writer, Analyst, Reviewer working in parallel
- 🌅 **Daily Digest** — automated morning briefing with goals, drift, and cost summary
- 🛡️ **HITL Gate** — risk-scored approval gate before every action
- 🎤 **Voice Interface** — offline Whisper STT + local TTS
- 🌐 **Browser Agent** — fetch and read any URL, inject into context

</details>

<details>
<summary><b>v1.0–1.2 — Core Intelligence</b></summary>

<br/>

- 🔍 **Drift Detector** — stalled tasks, missed follow-ups, repetitive failures
- 🏥 **Self-Healing** — auto-retry with LLM provider switching
- 🎭 **Emotional Intelligence** — tone detection + adaptive responses
- 🎯 **Goal Tracker** — persistent goals with deviation warnings
- 📬 **Session Briefing** — smart catch-up when you return after absence
- 🔐 **Privacy Vault** — AES-256-GCM local secrets manager
- 🎭 **Persona Engine** — Work / Focus / Client / Research modes
- 📴 **Offline Mode** — local Ollama fallback + task queue
- ⚖️ **Load Balancer** — 3-node cluster with health checks and auto-failover

</details>

---

## 🤝 Built by the Community

```bash
git clone https://github.com/Omkar0612/nexus-ai
# Pick an issue labelled 'good first issue'
# Build a new skill (browser, phone, email are great examples)
# Open a PR
```

See [CONTRIBUTING.md](CONTRIBUTING.md) · Most wanted: **new skills · free API integrations · use-case examples**

---

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Omkar0612/nexus-ai&type=Date)](https://star-history.com/#Omkar0612/nexus-ai)

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=600&size=16&duration=4000&pause=2000&color=4ADE80&center=true&vCenter=true&width=500&lines=If+NEXUS+saved+you+time+%E2%80%94+a+%E2%AD%90+means+a+lot.;Built+with+%E2%9D%A4%EF%B8%8F+and+500%2B+Reddit+complaints.;Free+forever.+MIT+licensed." alt="footer typing" />

<br/>

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=120&section=footer" width="100%"/>

</div>
