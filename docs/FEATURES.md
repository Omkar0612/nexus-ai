# 🧠 NEXUS Feature Reference

All features shipped. All free. All open-source.

---

## Core Features (v1.0)

### 🔴 Drift Detector
- Detects stalled tasks, missed follow-ups, repetitive failures
- **CLI:** `nexus drift`

### 🔧 Self-Healing Agent
- Auto-diagnoses and retries broken tasks
- **CLI:** `nexus health`

### 🎭 Emotional Intelligence
- Detects frustration/stress/excitement and adapts tone

### 🎯 Goal Tracker
- Tracks long-term goals, warns on misalignment
- **CLI:** `nexus goals list` | `nexus goals set "goal"`

### 👋 Session Briefer
- Context briefing after 30+ min away

### 📚 Adaptive Learner
- Learns your patterns, personalises every response
- **CLI:** `nexus insights`

### 🔐 Privacy Vault
- AES-256-GCM encrypted local secrets
- **CLI:** `nexus vault set KEY value` | `nexus vault get KEY`

### 🎭 Persona Engine
- 6 switchable modes: default, work, creative, client, focus, research
- **CLI:** `nexus persona use work`

### 📡 Offline Mode
- Auto-switches to Ollama, queues tasks

---

## v1.1 Features

### 💰 Token Cost Tracker
- Real-time cost with budget caps + Telegram alerts
- **CLI:** `nexus cost report` | `nexus budget set --daily 1.00`

### 📋 Audit Log
- Queryable decision log with rationale + risk classification
- **CLI:** `nexus audit show` | `nexus audit show --risk high`

### 🔁 Loop Detector
- Breaks infinite agent loops before they burn tokens

---

## v1.2 Features

### 🤥 Hallucination Detector
- Tags every response: ✅ VERIFIED | 💡 UNCERTAIN | ⚠️ UNVERIFIED | 🚨 CONTRADICTED
- Zero external API, works offline

### ⏰ Smart Scheduler
- Timezone-aware cron + condition-based triggers
- `FileExistsCondition`, custom Go conditions, retry backoff
- **CLI:** `nexus heartbeat add "name" "0 9 * * *" "task"`

### 📁 Knowledge Base (RAG)
- TF-IDF search over `~/.nexus/kb/` files, offline, zero cost
- **CLI:** `nexus kb stats` | `nexus kb search "query"`

---

## v1.3 Features

### 🤖 Multi-Agent Bus
Coordinate specialised sub-agents over a typed message bus.
- Roles: Researcher, Coder, Writer, Analyst, Reviewer
- Auto-routes tasks to best-fit agent based on keywords
- Broadcast to all agents in parallel
- Loop detection built into the bus
- **CLI:** `nexus agent bus list` | `nexus agent route "task"`

### 🌅 Daily Intelligence Digest
Automated morning briefing from all NEXUS systems.
- Pulls: drift signals, goals, cost summary, audit highlights, scheduled jobs
- Priority tiers: 🔴 high / 🟡 medium / 🟢 low
- Formats: Telegram Markdown, CLI, JSON
- **CLI:** `nexus digest` | `nexus digest --json`

### 🔐 HITL Gate (Human-in-the-Loop)
Approval gate for high-risk agent actions.
- Low risk: auto-execute | Medium: audit log | High: Telegram approval
- Timeout = safe cancel (fail-closed)
- Emergency lock: blocks ALL non-low-risk actions instantly
- **CLI:** `nexus hitl approve <id>` | `nexus hitl reject <id>` | `nexus hitl lock`

### 🎤 Voice Interface
Speak to NEXUS, hear it speak back.
- Wake word detection ("Hey NEXUS")
- Push-to-talk mode
- Whisper transcription (local, offline)
- TTS: espeak or piper (local, offline)
- Simulated mode for CI/testing
- **CLI:** `nexus voice start` | `nexus voice status`

### 🌐 Browser Agent
Autonomous web browsing.
- Navigate, click, fill forms, extract text, screenshot
- Safety allowlist + blocked hosts (no localhost/internal)
- Visit limiter (loop protection)
- Task planner converts natural language → action sequence
- **CLI:** `nexus browse "go to https://example.com and extract text"`

---

## v1.4 Roadmap

- [ ] 📊 Analytics Dashboard — web UI for cost, audit, goals, agent stats
- [ ] 📞 Phone Call Agent — make and receive calls via Twilio
- [ ] 📧 Email Agent — read, reply, draft emails via IMAP/SMTP
- [ ] 📝 Note-taking Agent — auto-capture and organise meeting notes
- [ ] 🤝 GitHub Agent — open issues, review PRs, create branches autonomously
- [ ] 📱 Mobile Companion — NEXUS on iOS/Android via Telegram bot
