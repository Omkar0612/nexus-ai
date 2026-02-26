#!/usr/bin/env python3
"""
Autonomous Business Intelligence Monitor

Fetches: HackerNews top stories + Reddit signals + Crypto prices
Sends:   Daily briefing to Telegram (or stdout if no token set)
Cost:    $0.00/month (all free APIs)

Setup:
    pip install httpx
    export TELEGRAM_BOT_TOKEN=your_token
    export TELEGRAM_CHAT_ID=your_chat_id
    python main.py
"""

import asyncio
import json
import os
import sys
from datetime import datetime
from typing import Optional

try:
    import httpx
    HAS_HTTPX = True
except ImportError:
    HAS_HTTPX = False

TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")
TELEGRAM_CHAT_ID = os.getenv("TELEGRAM_CHAT_ID", "")
NEXUS_API = os.getenv("NEXUS_API", "http://localhost:7700")
HN_MIN_SCORE = int(os.getenv("HN_MIN_SCORE", "300"))
HN_KEYWORDS = os.getenv("HN_KEYWORDS", "AI,automation,startup,n8n,SaaS,open-source").split(",")
CRYPTO_COINS = os.getenv("CRYPTO_COINS", "bitcoin,ethereum,solana")


async def fetch_hn_stories(keywords: list, min_score: int = 300) -> list:
    """Fetch HN top stories matching keywords above score threshold."""
    if not HAS_HTTPX:
        return [{"title": "httpx not installed", "score": 0, "url": ""}]
    async with httpx.AsyncClient(timeout=20.0) as client:
        resp = await client.get("https://hacker-news.firebaseio.com/v0/topstories.json")
        ids = resp.json()[:60]
        matched = []
        for sid in ids[:40]:
            try:
                r = await client.get(f"https://hacker-news.firebaseio.com/v0/item/{sid}.json")
                s = r.json()
                if not s or s.get("score", 0) < min_score:
                    continue
                title = s.get("title", "").lower()
                if any(k.lower() in title for k in keywords):
                    matched.append({
                        "title": s.get("title"),
                        "score": s.get("score"),
                        "url": s.get("url", f"https://news.ycombinator.com/item?id={sid}"),
                        "comments": s.get("descendants", 0),
                    })
            except Exception:
                continue
    return sorted(matched, key=lambda x: x["score"], reverse=True)[:5]


async def fetch_reddit_signals(subreddits: Optional[list] = None) -> list:
    """Fetch hot posts from relevant subreddits (no auth required)."""
    if not HAS_HTTPX:
        return []
    if subreddits is None:
        subreddits = ["entrepreneur", "selfhosted", "LocalLLaMA", "n8n", "startups"]
    posts = []
    headers = {"User-Agent": "nexus-ai-monitor/1.0"}
    async with httpx.AsyncClient(timeout=15.0, headers=headers) as client:
        for sub in subreddits[:3]:
            try:
                r = await client.get(f"https://www.reddit.com/r/{sub}/hot.json?limit=5")
                data = r.json()
                for post in data.get("data", {}).get("children", []):
                    p = post.get("data", {})
                    if p.get("score", 0) > 50:
                        posts.append({
                            "subreddit": sub,
                            "title": p.get("title"),
                            "score": p.get("score"),
                            "url": f"https://reddit.com{p.get('permalink', '')}",
                        })
            except Exception:
                continue
    return sorted(posts, key=lambda x: x["score"], reverse=True)[:6]


async def fetch_crypto_prices(coins: str = "bitcoin,ethereum,solana") -> dict:
    """Fetch crypto prices from CoinGecko free API (no key needed)."""
    if not HAS_HTTPX:
        return {}
    url = f"https://api.coingecko.com/api/v3/simple/price?ids={coins}&vs_currencies=usd&include_24hr_change=true"
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            r = await client.get(url)
            data = r.json()
        return {
            coin: {
                "usd": info.get("usd"),
                "change_24h": round(info.get("usd_24h_change", 0), 2),
            }
            for coin, info in data.items()
        }
    except Exception:
        return {}


def format_briefing(hn: list, reddit: list, crypto: dict) -> str:
    """Format the daily intelligence briefing."""
    today = datetime.utcnow().strftime("%b %d, %Y")
    lines = [f"\ud83d\udcca *NEXUS Business Briefing \u2014 {today}*\n"]

    if hn:
        lines.append("\ud83d\udd25 *Hacker News*")
        for s in hn[:3]:
            lines.append(f"\u2022 {s['title']} \u2014 {s['score']} pts")
        lines.append("")

    if reddit:
        lines.append("\ud83d\udcac *Reddit Signals*")
        for p in reddit[:3]:
            lines.append(f"\u2022 r/{p['subreddit']}: {p['title']} ({p['score']} upvotes)")
        lines.append("")

    if crypto:
        parts = []
        symbols = {"bitcoin": "BTC", "ethereum": "ETH", "solana": "SOL"}
        for coin, info in crypto.items():
            sym = symbols.get(coin, coin.upper())
            change = info["change_24h"]
            sign = "+" if change >= 0 else ""
            parts.append(f"{sym} ${info['usd']:,} ({sign}{change}%)")
        lines.append("\ud83d\udcb0 *Crypto:* " + " | ".join(parts))

    return "\n".join(lines)


async def send_telegram(text: str, token: str, chat_id: str) -> bool:
    """Send message to Telegram."""
    if not HAS_HTTPX or not token or not chat_id:
        return False
    async with httpx.AsyncClient(timeout=10.0) as client:
        r = await client.post(
            f"https://api.telegram.org/bot{token}/sendMessage",
            json={"chat_id": chat_id, "text": text, "parse_mode": "Markdown"},
        )
    return r.json().get("ok", False)


async def main():
    print("\ud83e\udde0 NEXUS Autonomous Business Monitor starting...\n")

    print("Fetching HackerNews stories...")
    hn = await fetch_hn_stories(HN_KEYWORDS, HN_MIN_SCORE)

    print("Fetching Reddit signals...")
    reddit = await fetch_reddit_signals()

    print("Fetching crypto prices...")
    crypto = await fetch_crypto_prices(CRYPTO_COINS)

    briefing = format_briefing(hn, reddit, crypto)
    print("\n" + "=" * 60)
    print(briefing)
    print("=" * 60 + "\n")

    if TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID:
        ok = await send_telegram(briefing, TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID)
        print("Telegram:", "sent \u2705" if ok else "failed \u274c")
    else:
        print("\u2139\ufe0f  Set TELEGRAM_BOT_TOKEN + TELEGRAM_CHAT_ID to get Telegram alerts")


if __name__ == "__main__":
    asyncio.run(main())
