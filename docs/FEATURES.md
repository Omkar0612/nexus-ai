# 🧠 NEXUS — Feature Reference

The most fully-featured open-source AI agent. All free. All local. All yours.

---

## v1.0 — Core Agent
| Feature | CLI |
|---|---|
| 🔴 Drift Detector | `nexus drift` |
| 🔧 Self-Healing Agent | `nexus health` |
| 🎭 Emotional Intelligence | auto |
| 🎯 Goal Tracker | `nexus goals` |
| 👋 Session Briefer | auto |
| 📚 Adaptive Learner | `nexus insights` |
| 🔐 Privacy Vault | `nexus vault` |
| 🎭 Persona Engine | `nexus persona use work` |
| 📡 Offline Mode | auto |

## v1.1 — Observability
| Feature | CLI |
|---|---|
| 💰 Token Cost Tracker | `nexus cost report` |
| 📋 Audit Log | `nexus audit show` |
| 🔁 Loop Detector | auto |

## v1.2 — Intelligence
| Feature | CLI |
|---|---|
| 🤥 Hallucination Detector | auto |
| ⏰ Smart Scheduler | `nexus heartbeat add` |
| 📁 Knowledge Base (RAG) | `nexus kb search` |

## v1.3 — Autonomy
| Feature | CLI |
|---|---|
| 🤖 Multi-Agent Bus | `nexus agent route` |
| 🌅 Daily Intelligence Digest | `nexus digest` |
| 🔐 HITL Gate | `nexus hitl approve` |
| 🎤 Voice Interface | `nexus voice start` |
| 🌐 Browser Agent | `nexus browse` |

## v1.4 — Connected
| Feature | CLI / Access |
|---|---|
| 📊 Analytics Dashboard | `http://localhost:8080` |
| 📞 Phone Agent | `nexus phone call` / `nexus phone sms` |
| 📧 Email Agent | `nexus email inbox` / `nexus email send` |
| 📝 Note-taking Agent | `nexus notes new` / `nexus notes search` |
| 💙 GitHub Agent | `nexus github issues` / `nexus github pr` |
| 📱 Mobile Companion | Telegram bot |

---

## Architecture

```
nexus-ai/
├── cmd/                    # CLI entry points
├── internal/
│   ├── agents/             # core agent logic
│   │   ├── multiagent_bus.go
│   │   ├── hallucination_detector.go
│   │   ├── hitl_gate.go
│   │   └── loop_detector.go
│   ├── audit/              # audit log
│   ├── browser/            # web browsing
│   ├── dashboard/          # analytics HTTP API
│   ├── digest/             # daily briefing
│   ├── email/              # IMAP/SMTP agent
│   ├── github/             # GitHub operations
│   ├── kb/                 # knowledge base (RAG)
│   ├── mobile/             # Telegram companion
│   ├── notes/              # note-taking agent
│   ├── phone/              # Twilio phone/SMS
│   ├── scheduler/          # smart cron
│   ├── telemetry/          # cost tracker
│   └── voice/              # STT/TTS interface
└── docs/
    └── FEATURES.md
```

---

## v1.5 Roadmap

- [ ] 🌐 Web Scraping Agent — structured data extraction from any URL
- [ ] 💹 Crypto/Finance Agent — price alerts, portfolio tracking
- [ ] 📅 Calendar Agent — read/write Google Calendar + reminders
- [ ] 🧩 Plugin System — drop a .so or .wasm plugin into ~/.nexus/plugins/
- [ ] 🔒 E2E Encryption — encrypt all data at rest with user key
- [ ] 📊 Web UI — React dashboard for all NEXUS metrics
- [ ] 🤖 Agent Marketplace — install community agents with `nexus agent install`
