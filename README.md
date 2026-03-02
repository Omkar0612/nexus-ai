<div align="center">

<img src="https://capsule-render.vercel.app/api?type=venom&color=gradient&customColorList=6,11,20&height=280&section=header&text=NEXUS%20AI&fontSize=90&fontColor=fff&animation=fadeIn&fontAlignY=45&desc=The%20AI%20agent%20that%20actually%20works.%20Free%20forever.&descAlignY=65&descSize=18" width="100%"/>

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=22&duration=3000&pause=1000&color=7C3AED&center=true&vCenter=true&multiline=true&repeat=true&width=700&height=60&lines=Self-healing+%E2%80%A2+Drift-aware+%E2%80%A2+100%25+Free;Web+UI+%E2%80%A2+Image+Gen+%E2%80%A2+Voice+%E2%80%A2+Writing+Studio;Multi-agent+%E2%80%A2+Offline+%E2%80%A2+AES-256+Vault" alt="Typing SVG" />

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

[🚀 Quick Start](#-quick-start) · [✨ Features](#-features) · [🌐 Web UI](#-web-ui--v16) · [🎨 Creative Studio](#-v17-creative-studio) · [📋 Changelog](#-changelog) · [🔮 Roadmap](ROADMAP.md)

</div>

---

## 🚀 Quick Start

You can run NEXUS AI without writing a single line of code. Choose your preferred method below:

### Option 1: Docker (Fastest, zero setup)
Run the pre-compiled Web UI instantly:
```bash
docker run -p 7070:7070 ghcr.io/omkar0612/nexus-ai:latest
```
*Then open `http://localhost:7070` in your browser.*

### Option 2: Download Binaries (Windows / Linux)
1. Go to the [Releases Page](https://github.com/Omkar0612/nexus-ai/releases/latest)
2. Download the `.exe` (Windows) or binary (Linux) for your architecture.
3. Run `nexus start` from your terminal.

### Option 3: Compile from Source (macOS & Developers)
*Note: Due to SQLite's CGO requirements, Mac users must have Xcode Command Line Tools installed (`xcode-select --install`).*
```bash
# 1. Clone & build
git clone https://github.com/Omkar0612/nexus-ai
cd nexus-ai
CGO_ENABLED=1 go install ./cmd/nexus

# 2. Add your free API key (console.groq.com — 60 sec signup)
cp config/nexus.example.toml ~/.nexus/nexus.toml

# 3. Run — Web UI starts at http://localhost:7070
nexus start
```

> 🆓 **No paid API needed.** Works natively with Groq (free), Gemini (free), Ollama (100% offline), and OpenRouter.

<details>
<summary><b>🌐 CLI & Web UI flags &rarr;</b></summary>

```bash
nexus start                        # Web UI at :7070 (default)
nexus start --webui-addr :8080     # Custom port
nexus start --no-webui             # CLI only
nexus start --debug                # Verbose logging
```
</details>

---

## 🌐 Web UI — Embedded Dark Mode

```text
┌─────────────────────────────────────────────────────────────────┐
│  NEXUS AI  v1.7                              ● agents: 3 active │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  You: generate an image of Dubai skyline at sunset       │   │
│  │  NEXUS: ✅ Image saved → ./output/dubai-sunset.png       │   │
│  │                                               (done)     │   │
│  └──────────────────────────────────────────────────────────┘   │
│  📡 Agent Activity                                               │
│  ● imagegen   running   — Stable Diffusion generating...         │
│  ✓ writer     done      — caption generated                      │
│  [ Type a message...                              ] [ Send ▶ ]  │
└─────────────────────────────────────────────────────────────────┘
```

The WebUI uses an ultra-lightweight `//go:embed` architecture with zero JS frameworks (pure Server-Sent Events).

---

## 🎨 v1.7 Creative Studio

> *Replace Midjourney, ElevenLabs, Grammarly, Suno, and ChatGPT — for free.*

### 🖼️ Image Generation
```bash
# 1. Local Stable Diffusion (Automatic1111) — fully free, fully private
nexus imagine "minimalist logo, purple gradient, tech startup"

# 2. Together AI FLUX.1-schnell — free $25 credits
nexus imagine --backend together --width 1024 --height 768 \
  "Dubai Marina at golden hour, photorealistic, 8K"
```

### 🔊 Voice Synthesis
```bash
# Coqui TTS — 100% local, no API key required
nexus speak --backend coqui "System architecture is stable."

# System TTS — always works, zero setup (macOS/Windows native)
nexus speak "Reminder: standup in 5 minutes"
```

### ✍️ Writing Studio
```bash
# Full writing pipeline — draft → proofread → translate
nexus write draft --topic "Why AI agents beat SaaS tools" \
  --style persuasive --words 800 --out article.md

nexus write proofread --file article.md
```

### 📅 Calendar Agent
```bash
nexus calendar today
nexus calendar conflicts --week
nexus calendar free --duration 1h
```

---

## ✨ Features that fix the broken agent ecosystem

### 🔍 Drift Detector
> *The only AI agent that notices when your work is stalling.*
```text
🔴 [HIGH]   'nexus-api-refactor' stalled — last touched 2 days ago
🟡 [MEDIUM] Follow-up missed — 'ping client about invoice' (3 days)
```

### 🏥 Self-Healing Engine
> *Fails once. Never twice.*
```text
⚠️  Task 'daily-briefing' failed — Groq rate limit
    Switching to Gemini 2.0 Flash... Retrying in 30s...
✅  Task recovered. Cost: $0.00
```

### 🛡️ Human-in-the-Loop Risk Gate
| Risk | Examples | Behaviour |
|:---:|:---|:---|
| 🟢 Low | Read, Search, Chat | Silent execute |
| 🟡 Medium | Write file, Send message | Execute + log |
| 🔴 High | Delete, Push to GitHub, Call | Pause → ask permission |

### 🧠 Deep Memory
- **Episodic** — SQLite conversation history
- **Semantic** — TF-IDF cosine similarity search (zero external dependencies)
- **Vault** — AES-256-GCM encrypted local secrets

---

## 🆓 Free LLM Providers Supported

| Provider | Model | Speed | Cost | Privacy |
|:---:|:---:|:---:|:---:|:---:|
| **Ollama** | Any model | Local GPU | **Free** | 🔒 100% Offline |
| **Groq** | Llama 3.3 70B | ⚡ 300 tok/s | **Free** | Cloud |
| **Gemini** | 2.0 Flash | ⚡ Fast | **Free** (1M tok/day)| Cloud |
| **Together**| FLUX / Mixtral | ⚡ Fast | **Free** ($25 creds)| Cloud |

---

## ⚔️ NEXUS vs The World

| Capability | NEXUS | AutoGPT | CrewAI | n8n AI | LangChain |
|:---|:---:|:---:|:---:|:---:|:---:|
| **Self-healing failures** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Drift detection** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Risk gate (HITL)** | ✅ | ⚠️ | ⚠️ | ⚠️ | ❌ |
| **100% Offline mode** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **AES-256 Vault** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Single binary, Go** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **100% free to run** | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ |

---

## 🤝 Built by the Community

Want to build your own custom plugin in 5 lines of Go? 
See [CONTRIBUTING.md](CONTRIBUTING.md) to learn how to add new skills to the `nexus.Registry`.

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Omkar0612/nexus-ai&type=Date)](https://star-history.com/#Omkar0612/nexus-ai)

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=600&size=16&duration=4000&pause=2000&color=4ADE80&center=true&vCenter=true&width=500&lines=If+NEXUS+saved+you+time+%E2%80%94+a+%E2%AD%90+means+a+lot.;Built+with+%E2%9D%A4%EF%B8%8F+and+500%2B+Reddit+complaints.;Free+forever.+MIT+licensed." alt="footer typing" />

<br/>

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=120&section=footer" width="100%"/>

</div>
