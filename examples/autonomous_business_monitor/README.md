# Autonomous Business Intelligence Monitor

A complete working example using **zero paid APIs**. Runs forever. Alerts via Telegram.

## What it does

Every morning:
1. Monitors Hacker News for your keywords (score threshold configurable)
2. Monitors Reddit subreddits for signals
3. Gets live crypto prices from CoinGecko (free, no key)
4. Sends everything to NEXUS for AI synthesis
5. Delivers a clean briefing to your Telegram

## Setup (5 minutes)

```bash
pip install httpx

export TELEGRAM_BOT_TOKEN=your_bot_token
export TELEGRAM_CHAT_ID=your_chat_id
export NEXUS_API=http://localhost:7700

# Start NEXUS first
nexus start &

# Run the monitor
python main.py
```

## Schedule it (cron)

```bash
# Run every morning at 9am
0 9 * * * cd /path/to/examples/autonomous_business_monitor && python main.py
```

Or use NEXUS heartbeat (recommended):
```bash
nexus heartbeat add "morning-briefing" "0 9 * * *" "Run the autonomous business monitor and send briefing to Telegram"
```

## Sample Output

```
📊 NEXUS Business Briefing — Feb 26, 2026

🔥 Hacker News
• We built $2M ARR SaaS in 6 months using AI agents — 847 pts
• Show HN: Open-source Zapier replacement — 612 pts

💬 Reddit Signals
• r/entrepreneur: AI tools that actually save time (234 upvotes)
• r/startups: Anyone else replacing Zapier with n8n? (189 upvotes)

💰 Crypto: BTC $87,420 (+2.3%) | ETH $3,210 (+1.1%)

⚠️ Opportunity: 3 posts mention 'n8n alternative' — possible market gap
```

## Cost

**$0.00/month** — all APIs used are completely free.
