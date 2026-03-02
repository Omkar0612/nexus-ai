<div align="center">

<img src="https://capsule-render.vercel.app/api?type=venom&color=gradient&customColorList=6,11,20&height=280&section=header&text=NEXUS%20AI&fontSize=90&fontColor=fff&animation=fadeIn&fontAlignY=45&desc=The%20AI%20agent%20that%20actually%20works.%20Free%20forever.&descAlignY=65&descSize=18" width="100%"/>

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=22&duration=3000&pause=1000&color=7C3AED&center=true&vCenter=true&multiline=true&repeat=true&width=700&height=60&lines=Self-healing+%E2%80%A2+Drift-aware+%E2%80%A2+100%25+Free;Web+UI+%E2%80%A2+Image+Gen+%E2%80%A2+Voice+%E2%80%A2+Writing+Studio;Auto-Forge+WASM+%E2%80%A2+Shadow+Mode+%E2%80%A2+UI-to-API;Hive-Mind+P2P+Mesh+%E2%80%A2+Neuro-Fuzzing" alt="Typing SVG" />

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

## ✨ Features that fix the broken agent ecosystem

### 🛡️ Agentic Fuzzing (Neuro-Fuzzing)
> *An adversarial agent that attacks the code the creator agent just wrote.*
Before any auto-generated code is deployed to the Event Bus, the internal "Attacker Agent" heavily bombards the WebAssembly module with Null Bytes, SQL Injections, Path Traversals, and 10MB memory-exhaustion payloads. If the generated agent panics or hits an infinite loop (DoS), the deployment is instantly rejected. 

### 🕸️ Hive-Mind Mesh Computing
> *Turn your devices into a unified AI supercomputer.*
Run NEXUS on your phone, laptop, and VPS. Using the P2P Mesh Network, they automatically discover each other. If you ask your phone to generate a heavy Stable Diffusion image, the phone's NEXUS dynamically routes the compute payload to your Desktop's GPU over your local network, and returns the result to your phone. 

### 🧠 Predictive Pre-Computation (Zero-Latency AI)
> *Why wait 3 minutes for Deep Research when NEXUS already did it?*
AutoAgent and LangChain sit idle until you type a prompt. NEXUS monitors your world. If it sees you have a meeting at 2:00 PM, or you just pushed a broken commit, it spins up the background workers instantly. By the time you open the WebUI, the meeting brief and the code fix are already cached and waiting for you. Zero latency.

### ⚡ Auto-Forge (Hot-Loaded WASM Agents)
> *Natural language agent creation without Docker restarts.*
Tell NEXUS to create a new agent (e.g. "Create a Dubai Real Estate scraper"). NEXUS autonomously writes the Go code, compiles it to WebAssembly via the NEXUS Cloud Compiler, and hot-loads it into the running sandbox in milliseconds. Zero restarts. Zero dependencies.

### 👁️ UI-to-API Reverse Engineering
> *If an app has a UI, NEXUS can build an API for it.*
Point NEXUS to an undocumented web app or legacy ERP. It launches a headless browser, logs in, intercepts the network traffic (HAR), extracts the Bearer tokens to your AES-256 Vault, and automatically synthesizes a native Wasm tool integration.

### 🕵️‍♂️ "Shadow Mode" Self-Evolution
> *Safe, measurable self-improvement without hallucinations.*
NEXUS continuously tests faster models and optimized prompts in a hidden background "Shadow Swarm". If it finds a way to perform a task 40% cheaper or faster without degrading quality, it pings your Human-in-the-Loop gate: *"I found a way to save API costs. Approve upgrade? [Y/N]"*.

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

### 🧠 Deep Memory
- **Episodic** — SQLite conversation history
- **Semantic** — SQLite Vector Search with Ollama Embeddings (fully local)
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

| Capability | NEXUS | AutoAgent | AutoGPT | CrewAI | LangChain |
|:---|:---:|:---:|:---:|:---:|:---:|
| **Agentic Fuzzing (Security)**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **P2P Mesh Computing** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Zero-Latency (Pre-Compute)**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Zero-Code Agent Generation**| ✅ (Hot WASM) | ✅ (Docker) | ❌ | ❌ | ❌ |
| **Reverse Engineer UI to API**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Self-healing failures** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Shadow Mode Evolution** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Risk gate (HITL)** | ✅ | ❌ | ⚠️ | ⚠️ | ❌ |
| **100% Offline mode** | ✅ | ✅ | ❌ | ❌ | ❌ |

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
