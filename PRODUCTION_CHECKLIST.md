# 🚀 NEXUS Production Readiness Checklist

## Security
- [ ] Change vault passphrase: `nexus vault set-passphrase`
- [ ] Set `security.allow_unsigned_skills = false`
- [ ] Bind web gateway to `127.0.0.1` (not `0.0.0.0`)
- [ ] Set `gateways.telegram.chat_id` to YOUR chat ID only
- [ ] Review audit log weekly: `nexus audit show`
- [ ] Enable `human_in_loop_high_risk = true`

## Performance
- [ ] Set Groq as default LLM (fastest free: 300+ tok/sec)
- [ ] Install Ollama for offline fallback: `ollama pull llama3.2`
- [ ] Set `max_episodic = 5000` (prevents memory DB bloat)
- [ ] Configure memory encryption key (auto-generated on first run)

## Reliability
- [ ] Enable self-healing: `heartbeat.enabled = true`
- [ ] Test offline mode: `nexus status --offline-test`
- [ ] Configure at least 2 LLM providers (primary + fallback)
- [ ] Set `agents.timeout_seconds = 120`

## Features to Enable First
- [ ] `nexus persona use work` — switch to work mode
- [ ] `nexus goal set "Your primary goal" --priority 5`
- [ ] `nexus drift` — run first drift scan
- [ ] `nexus memory list` — verify memory is working
- [ ] `nexus skill install web-search` — install core skills

## Multi-node Deployment
- [ ] `docker compose up -d` — starts 3-node cluster
- [ ] Verify LB health: `curl http://localhost:7700/lb/stats`
- [ ] Check all workers: `curl http://localhost:7700/api/health`
