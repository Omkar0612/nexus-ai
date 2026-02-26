#!/usr/bin/env python3
"""
NEXUS Worker — Python-side background task processor.

Handles tasks that are better in Python than Go:
  - Web scraping (BeautifulSoup / httpx)
  - ML inference calls (optional)
  - Crypto price fetching (CoinGecko free API)
  - HackerNews + Reddit monitoring
  - Telegram notification delivery

Communicates with the NEXUS Go daemon via HTTP (localhost:7700).
"""

import asyncio
import json
import os
import logging
from datetime import datetime
from typing import Optional

try:
    import httpx
except ImportError:
    httpx = None

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
log = logging.getLogger("nexus.worker")

NEXUS_API = os.getenv("NEXUS_API", "http://localhost:7700")
TELEGRAM_BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")
TELEGRAM_CHAT_ID = os.getenv("TELEGRAM_CHAT_ID", "")
WORKER_INTERVAL = int(os.getenv("WORKER_INTERVAL_SECONDS", "30"))


class NexusWorker:
    """Background worker that polls for tasks and executes them."""

    def __init__(self):
        self.running = False
        self.task_handlers = {
            "fetch_hn": self.fetch_hackernews,
            "fetch_crypto": self.fetch_crypto_prices,
            "send_telegram": self.send_telegram,
            "health_ping": self.health_ping,
        }

    async def start(self):
        """Start the worker loop."""
        self.running = True
        log.info("NEXUS Worker started — polling every %ds", WORKER_INTERVAL)
        while self.running:
            try:
                await self.poll_and_execute()
            except Exception as e:
                log.error("Worker loop error: %s", e)
            await asyncio.sleep(WORKER_INTERVAL)

    async def stop(self):
        """Stop the worker loop."""
        self.running = False
        log.info("NEXUS Worker stopped")

    async def poll_and_execute(self):
        """Poll NEXUS daemon for pending tasks and execute them."""
        if not httpx:
            log.warning("httpx not installed — skipping task poll")
            return
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.get(f"{NEXUS_API}/api/tasks/pending")
                if resp.status_code == 200:
                    tasks = resp.json().get("tasks", [])
                    for task in tasks:
                        await self.execute_task(task)
        except Exception as e:
            log.debug("Daemon not reachable: %s", e)

    async def execute_task(self, task: dict):
        """Execute a single task by type."""
        task_type = task.get("type", "")
        handler = self.task_handlers.get(task_type)
        if handler:
            log.info("Executing task: %s", task_type)
            try:
                result = await handler(task)
                await self.report_result(task.get("id"), result)
            except Exception as e:
                log.error("Task %s failed: %s", task_type, e)
        else:
            log.warning("Unknown task type: %s", task_type)

    async def fetch_hackernews(
        self, task: dict, min_score: int = 200, keywords: Optional[list] = None
    ) -> dict:
        """Fetch top HackerNews stories matching keywords."""
        if not httpx:
            return {"error": "httpx not available"}
        if keywords is None:
            keywords = task.get("keywords", ["AI", "startup", "automation", "n8n"])
        async with httpx.AsyncClient(timeout=15.0) as client:
            resp = await client.get(
                "https://hacker-news.firebaseio.com/v0/topstories.json"
            )
            story_ids = resp.json()[:50]
            stories = []
            for sid in story_ids[:30]:
                try:
                    sr = await client.get(
                        f"https://hacker-news.firebaseio.com/v0/item/{sid}.json"
                    )
                    story = sr.json()
                    if story and story.get("score", 0) >= min_score:
                        title = story.get("title", "").lower()
                        if not keywords or any(k.lower() in title for k in keywords):
                            stories.append(
                                {
                                    "title": story.get("title"),
                                    "score": story.get("score"),
                                    "url": story.get("url", ""),
                                    "comments": story.get("descendants", 0),
                                }
                            )
                except Exception:
                    continue
        log.info("HN: found %d matching stories", len(stories))
        return {"stories": stories, "fetched_at": datetime.utcnow().isoformat()}

    async def fetch_crypto_prices(self, task: dict) -> dict:
        """Fetch crypto prices from CoinGecko (free, no API key)."""
        if not httpx:
            return {"error": "httpx not available"}
        coins = task.get("coins", "bitcoin,ethereum,solana")
        url = (
            f"https://api.coingecko.com/api/v3/simple/price"
            f"?ids={coins}&vs_currencies=usd&include_24hr_change=true"
        )
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.get(url)
            data = resp.json()
        prices = {}
        for coin, info in data.items():
            prices[coin] = {
                "usd": info.get("usd"),
                "change_24h": round(info.get("usd_24h_change", 0), 2),
            }
        return {"prices": prices, "fetched_at": datetime.utcnow().isoformat()}

    async def send_telegram(self, task: dict) -> dict:
        """Send a message via Telegram bot."""
        if not httpx:
            return {"error": "httpx not available"}
        token = task.get("token") or TELEGRAM_BOT_TOKEN
        chat_id = task.get("chat_id") or TELEGRAM_CHAT_ID
        text = task.get("text", "")
        if not token or not chat_id:
            return {"error": "TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID required"}
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.post(
                f"https://api.telegram.org/bot{token}/sendMessage",
                json={"chat_id": chat_id, "text": text, "parse_mode": "Markdown"},
            )
        return {"ok": resp.json().get("ok", False)}

    async def health_ping(self, task: dict) -> dict:
        """Simple health check ping."""
        return {"status": "ok", "worker": "nexus-python", "ts": datetime.utcnow().isoformat()}

    async def report_result(self, task_id: str, result: dict):
        """Report task result back to NEXUS daemon."""
        if not httpx or not task_id:
            return
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                await client.post(
                    f"{NEXUS_API}/api/tasks/{task_id}/result",
                    json=result,
                )
        except Exception:
            pass


def main():
    worker = NexusWorker()
    try:
        asyncio.run(worker.start())
    except KeyboardInterrupt:
        log.info("Shutting down NEXUS worker")


if __name__ == "__main__":
    main()
