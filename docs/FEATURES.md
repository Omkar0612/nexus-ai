# 🧠 NEXUS Feature Reference

All features shipped. All free. All open-source.

---

## Core Features (v1.0)

### 🔴 Drift Detector
Silently monitors your conversation history for work drift signals.
- Detects stalled tasks (mentioned but never completed)
- Detects missed follow-ups ("follow up with X" — never resolved)
- Detects repetitive failures (same error 3+ times)
- **CLI:** `nexus drift`

### 🔧 Self-Healing Agent
Auto-diagnoses and retries broken tasks.
- Captures full error context on failure
- Attempts automatic retry with LLM root-cause analysis
- Escalates to user after max retries with plain-language diagnosis
- **CLI:** `nexus health`

### 🎭 Emotional Intelligence
Detects your emotional state and adapts NEXUS tone accordingly.
- Detects: frustrated, urgent, stressed, excited, confused, neutral
- Adapts: verbosity, formality, response style, emoji usage
- Zero configuration — works automatically

### 🎯 Goal Tracker
Tracks your long-term goals and aligns every response.
- Infer goals from conversation (no setup needed)
- Warns when a task is misaligned with your declared goals
- Reports goals you haven't worked toward in 7+ days
- **CLI:** `nexus goals list` | `nexus goals set "goal"`

### 👋 Session Briefer
Proactive context briefing when you return after being away.
- Summarises where you left off
- Surfaces high-priority drift signals
- Suggests 3 quick actions to resume
- Triggers after 30+ minutes away

### 📚 Adaptive Learner
Learns your workflow patterns and personalises every response.
- Tracks which agents, topics, and formats you prefer
- Automatically adjusts LLM system prompts to match your style
- **CLI:** `nexus insights`

### 🔐 Privacy Vault
AES-256-GCM encrypted local secrets manager.
- Stores API keys, passwords, tokens encrypted at rest
- Auto-redacts secret values from all LLM prompts
- Personal and business privacy zones
- **CLI:** `nexus vault set KEY value` | `nexus vault get KEY` | `nexus vault list`

### 🎭 Persona Engine
6 switchable AI work modes.
| Persona | Best For |
|---|---|
| `default` | General use |
| `work` | Focused, professional, code-heavy |
| `creative` | Brainstorming, ideation |
| `client` | Client-facing, no internal data |
| `focus` | Deep work, 200-word cap |
| `research` | Thorough, cited, academic |
- **CLI:** `nexus persona use work`

### 📡 Offline Mode
Full functionality without internet.
- Auto-detects connectivity loss
- Switches all LLM calls to local Ollama automatically
- Queues tasks and flushes when back online

---

## New Features (v1.1)

### 💰 Token Cost Tracker
Real-time token cost tracking with budget protection.
- Built-in pricing for Groq, Anthropic, OpenAI, Ollama
- Daily and monthly budget caps with auto-pause
- Telegram alert before and when budget is breached
- Suggests cheaper model alternatives automatically
- **CLI:** `nexus cost report` | `nexus budget set --daily 1.00`

### 📋 Audit Log
Fully queryable agent decision log — know WHY NEXUS did anything.
- Records: action, rationale, context used, alternatives considered
- Auto-classifies risk level (low/medium/high)
- Human-in-loop approval tracking
- JSON export for compliance
- **CLI:** `nexus audit show` | `nexus audit show --risk high` | `nexus audit export`

### 🔁 Loop Detector
Detects and breaks infinite agent loops before they waste tokens.
- Monitors tool call patterns in real-time
- Fires after 3 identical (tool, input) pairs
- Estimates tokens and cost saved by catching the loop
- Suggests specific fixes based on tool type
- Resets cleanly per session

---

## Roadmap

- [ ] 🤥 Hallucination Detector — flag unverified factual claims
- [ ] 🤝 Multi-Agent Bus — spawn and coordinate sub-agents
- [ ] 📁 Knowledge Base (RAG) — search your own docs
- [ ] 🕐 Smart Scheduler — condition-based + event-triggered cron
- [ ] 📊 Daily Digest — automated morning intelligence briefing
- [ ] 🌍 Timezone Context — time-aware task scheduling
- [ ] 🔐 HITL Gate — Telegram approval for high-risk actions
