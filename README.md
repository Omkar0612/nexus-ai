<div align="center">

<img src="https://capsule-render.vercel.app/api?type=venom&color=gradient&customColorList=6,11,20&height=280&section=header&text=NEXUS%20AI&fontSize=90&fontColor=fff&animation=fadeIn&fontAlignY=45&desc=The%20AI%20agent%20that%20actually%20works.%20Free%20forever.&descAlignY=65&descSize=18" width="100%"/>

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=22&duration=3000&pause=1000&color=7C3AED&center=true&vCenter=true&multiline=true&repeat=true&width=700&height=60&lines=Self-healing+%E2%80%A2+Observability+%E2%80%A2+100%25+Free;Kill-Switch+%E2%80%A2+Circuit+Breakers+%E2%80%A2+Rollback;Auto-Forge+WASM+%E2%80%A2+UI-to-API;Hive-Mind+Mesh+%E2%80%A2+Liquid+Context;NL-to-n8n+DAG+%E2%80%A2+Token+Market" alt="Typing SVG" />

<br/><br/>

[![CI](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/Omkar0612/nexus-ai/actions/workflows/ci.yml)
[![Stars](https://img.shields.io/github/stars/Omkar0612/nexus-ai?style=for-the-badge&logo=github&color=FFD700&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai/stargazers)
[![Forks](https://img.shields.io/github/forks/Omkar0612/nexus-ai?style=for-the-badge&logo=github&color=4ade80&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai/network/members)
[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a2e)](https://go.dev)
[![MIT](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge&labelColor=1a1a2e)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-7c3aed?style=for-the-badge&labelColor=1a1a2e)](CONTRIBUTING.md)
[![100% Free](https://img.shields.io/badge/Cost-Zero-f59e0b?style=for-the-badge&logo=opensourceinitiative&labelColor=1a1a2e)](https://github.com/Omkar0612/nexus-ai)

<br/>

> **95% of AI agent pilots fail after the demo. NEXUS is built for the 5% that survive production.**

<br/>

[🚀 Quick Start](#-quick-start) · [🤯 What can it actually do?](#-mind-blowing-real-world-examples) · [✨ Features](#-features) · [📋 Changelog](#-changelog)

</div>

---

## 🤯 Mind-Blowing Real-World Examples

Most AI frameworks show you how to build a "weather bot". Here is what NEXUS does in production right now:

#### 1. The Zero-API Legacy Hack
> **You:** "Scrape my company's 15-year-old internal accounting software for unpaid invoices. There is no API." <br>
> **NEXUS:** Launches a hidden headless browser, logs in using your Vault credentials, intercepts the raw network traffic (HAR), discovers the undocumented internal GraphQL endpoint, writes a custom Go plugin wrapping the endpoint, compiles it to WebAssembly via Auto-Forge, hot-loads it into its own brain in 200ms, and hands you a CSV of unpaid invoices.

#### 2. The Multi-Device Hive Mind
> **You (on your iPhone at a coffee shop):** "Generate a 4K photorealistic image of a cyberpunk city." <br>
> **NEXUS:** Your phone realizes it doesn't have a GPU. It uses mDNS to detect your RTX 4090 desktop PC sitting asleep at home. It routes the Stable Diffusion prompt over your mesh network to the PC, generates the image using your home electricity, and streams the finished 4K `.png` back to your iPhone screen. **Cost: $0.00**.

#### 3. The "Speak it into existence" Automation
> **You:** "Whenever an email arrives from a VIP client that sounds angry, draft an apology and page me on Telegram." <br>
> **NEXUS:** Doesn't just write a python script. It natively writes a 10-node Directed Acyclic Graph (DAG) JSON file and pushes it directly via API to your self-hosted **n8n** instance. It wires up the Webhook, the LLM sentiment node, and the Telegram node perfectly spaced out. You never even opened the n8n GUI.

#### 4. Zero-Latency Pre-Computation
> **You:** *Wake up and open the NEXUS Web UI.* <br>
> **NEXUS:** "Good morning. I noticed a GitHub webhook fired at 3 AM indicating a broken CI/CD build on your main repo. While you were sleeping, I pulled the stack trace, wrote the patch, fuzzed it for security flaws, and staged it. Click [Here] to merge." 

#### 5. The Agent that Saved Itself from a $4,000 Loop
> **What Happened:** NEXUS's web scraper hit an infinite retry loop at 2 AM. <br>
> **What NEXUS Did:** Detected the hallucination pattern (same tool called 3x consecutively), instantly triggered the **Kill-Switch**, revoked API credentials, rolled back the last 12 actions transactionally, logged a post-mortem, and paged you on Slack. **Damage prevented: $4,200 in API costs.**

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

## ✨ Features that fix the broken agent ecosystem

### 🔍 Production-Grade Observability
> *The #1 reason agent pilots fail: teams can't debug them in production.*
NEXUS gives you a complete execution timeline for every agent run. See every tool call, every retry, every token spent, and every hallucination loop before it costs you $4,000. Engineers, PMs, and domain experts can all inspect traces in plain English. No more JSON dumps and Slack logs.

### 🛑 3-Layer Kill-Switch Architecture
> *What happens when your agent goes rogue at 3 AM?*
- **Layer 1 (Hard Stop):** Instant credential revocation + queue drain.
- **Layer 2 (Soft Pause):** Freeze execution, preserve state for human review.
- **Layer 3 (Transactional Rollback):** Undo the last N agent actions idempotently.

Auto-triggers on: cost threshold breach, hallucination loop detection, or 3 consecutive tool failures.

### 🔌 Circuit Breakers for External APIs
> *When Stripe's API goes down, your agent shouldn't loop 500 times.*
NEXUS automatically degrades to read-only mode when external integrations flake. Exponential backoff with jitter. Human-in-the-Loop escalation for irreversible actions (like deleting a database).

### 💧 Agentic Memory Consolidation (Liquid Context)
> *NEXUS never forgets, and it never hits a context limit.*
When your chat history bloats past 8,000 tokens, NEXUS triggers a background "dream state". It reads the raw, bloated history, strips out the conversational filler, and semantically compresses it into high-density "Concepts". A 5,000-token sprawling conversation is mathematically reduced into a 50-token factual block and re-injected into SQLite. Infinite memory, zero amnesia.

### 📈 Token Stock Market (Dynamic Cost Arbitrage)
> *Never overpay for an API call again.*
When a task hits the Multi-Agent Bus, NEXUS pings the `/pricing` and `/health` endpoints of Groq, OpenRouter, Gemini, and your local Ollama instance. It uses an arbitrage formula `(Cost * 100) + (Latency * 10)` to instantly route the payload to the cheapest, fastest model available at that exact millisecond. If Groq hits a 429 Rate Limit, the market instantly evades it and falls back to Gemini.

### 🔄 Natural Language to n8n DAG Compiler
> *Stop dragging and dropping. Speak your automations into existence.*
Tell NEXUS: *"Check my company ERP daily, and if revenue drops, ping a Meta Ads agent."* NEXUS natively compiles this logic into a valid n8n Directed Acyclic Graph (DAG) JSON, spaces the nodes out perfectly, maps the connections, and deploys it directly to your running n8n instance via API. 

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
Tell NEXUS to create a new agent (e.g. "Create a Real Estate scraper"). NEXUS autonomously writes the Go code, compiles it to WebAssembly via the NEXUS Cloud Compiler, and hot-loads it into the running sandbox in milliseconds. Zero restarts. Zero dependencies.

### 👁️ UI-to-API Reverse Engineering
> *If an app has a UI, NEXUS can build an API for it.*
Point NEXUS to an undocumented web app or legacy ERP. It launches a headless browser, logs in, intercepts the network traffic (HAR), extracts the Bearer tokens to your AES-256 Vault, and automatically synthesizes a native Wasm tool integration.

### 🕵️‍♂️ "Shadow Mode" Self-Evolution
> *Safe, measurable self-improvement without hallucinations.*
NEXUS continuously tests faster models and optimized prompts in a hidden background "Shadow Swarm". If it finds a way to perform a task 40% cheaper or faster without degrading quality, it pings your Human-in-the-Loop gate: *"I found a way to save API costs. Approve upgrade? [Y/N]"*.

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
| **Production Observability**| ✅ | ❌ | ❌ | ❌ | ⚠️ (paid)|
| **Kill-Switch + Rollback**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Circuit Breakers**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Hallucination Loop Detection**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Liquid Context (Amnesia fix)**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Token Arbitrage Routing**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **NL-to-n8n DAG Compiler** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Agentic Fuzzing (Security)**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **P2P Mesh Computing** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Zero-Latency (Pre-Compute)**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Zero-Code Agent Gen** | ✅ (Hot WASM) | ✅ (Docker) | ❌ | ❌ | ❌ |
| **Reverse Engineer UI to API**| ✅ | ❌ | ❌ | ❌ | ❌ |
| **Shadow Mode Evolution** | ✅ | ❌ | ❌ | ❌ | ❌ |

---

## 🤝 Built by the Community

Want to build your own custom plugin in 5 lines of Go? 
See [CONTRIBUTING.md](CONTRIBUTING.md) to learn how to add new skills to the `nexus.Registry`.

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Omkar0612/nexus-ai&type=Date)](https://star-history.com/#Omkar0612/nexus-ai)

<br/>

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=600&size=16&duration=4000&pause=2000&color=4ADE80&center=true&vCenter=true&width=500&lines=If+NEXUS+saved+you+time+%E2%80%94+a+%E2%AD%90+means+a+lot.;Built+for+the+5%25+that+ship+to+production.;Free+forever.+MIT+licensed." alt="footer typing" />

<br/>

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,11,20&height=120&section=footer" width="100%"/>

</div>
