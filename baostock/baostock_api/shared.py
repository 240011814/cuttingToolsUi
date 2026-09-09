from __future__ import annotations

import json
import os
import threading
from dataclasses import asdict, dataclass
from datetime import date
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler
from pathlib import Path
from typing import Any

try:
    import baostock as bs
except ModuleNotFoundError as error:
    raise SystemExit(
        "Missing Python dependency 'baostock'. "
        "Create/use apps/baostock-api/.venv and run "
        "'apps/baostock-api/.venv/bin/pip install -r apps/baostock-api/requirements.txt'."
    ) from error


HOST = os.environ.get("BAOSTOCK_API_HOST", "127.0.0.1")
PORT = int(os.environ.get("BAOSTOCK_API_PORT", "3002"))
DAILY_LIMIT = int(os.environ.get("BAOSTOCK_API_DAILY_LIMIT", "100000"))
DATA_DIR = Path(__file__).resolve().parent.parent / "data"
USAGE_FILE = DATA_DIR / "usage.json"


@dataclass
class UsageStats:
    date: str
    count: int
    limit: int

    @property
    def remaining(self) -> int:
        return max(self.limit - self.count, 0)

    def to_dict(self) -> dict[str, Any]:
        payload = asdict(self)
        payload["remaining"] = self.remaining
        return payload


class UsageCounter:
    def __init__(self, usage_file: Path, daily_limit: int):
        self.usage_file = usage_file
        self.daily_limit = daily_limit
        self.lock = threading.Lock()
        self.usage_file.parent.mkdir(parents=True, exist_ok=True)
        self.stats = self._load_stats()

    def _today(self) -> str:
        return date.today().isoformat()

    def _default_stats(self) -> UsageStats:
        return UsageStats(date=self._today(), count=0, limit=self.daily_limit)

    def _load_stats(self) -> UsageStats:
        if not self.usage_file.exists():
            stats = self._default_stats()
            self._save_stats(stats)
            return stats

        try:
            payload = json.loads(self.usage_file.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, OSError):
            stats = self._default_stats()
            self._save_stats(stats)
            return stats

        stats = UsageStats(
            date=str(payload.get("date") or self._today()),
            count=int(payload.get("count") or 0),
            limit=int(payload.get("limit") or self.daily_limit),
        )
        return self._reset_if_needed(stats)

    def _save_stats(self, stats: UsageStats) -> None:
        self.usage_file.write_text(
            json.dumps(stats.to_dict(), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

    def _reset_if_needed(self, stats: UsageStats) -> UsageStats:
        today = self._today()
        if stats.date == today and stats.limit == self.daily_limit:
            return stats

        refreshed = UsageStats(date=today, count=0, limit=self.daily_limit)
        self._save_stats(refreshed)
        return refreshed

    def get_stats(self) -> UsageStats:
        with self.lock:
            self.stats = self._reset_if_needed(self.stats)
            return UsageStats(
                date=self.stats.date,
                count=self.stats.count,
                limit=self.stats.limit,
            )

    def consume(self) -> UsageStats:
        with self.lock:
            self.stats = self._reset_if_needed(self.stats)

            if self.stats.count >= self.stats.limit:
                raise RuntimeError("daily_limit_exceeded")

            self.stats.count += 1
            self._save_stats(self.stats)
            return UsageStats(
                date=self.stats.date,
                count=self.stats.count,
                limit=self.stats.limit,
            )


usage_counter = UsageCounter(USAGE_FILE, DAILY_LIMIT)
baostock_lock = threading.Lock()


def json_response(
    handler: BaseHTTPRequestHandler,
    status: HTTPStatus,
    payload: dict[str, Any],
) -> None:
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    handler.send_response(status.value)
    handler.send_header("Content-Type", "application/json; charset=utf-8")
    handler.send_header("Content-Length", str(len(body)))
    handler.end_headers()
    handler.wfile.write(body)


def text_response(
    handler: BaseHTTPRequestHandler,
    status: HTTPStatus,
    content: str,
) -> None:
    body = content.encode("utf-8")
    handler.send_response(status.value)
    handler.send_header("Content-Type", "text/plain; charset=utf-8")
    handler.send_header("Content-Length", str(len(body)))
    handler.end_headers()
    handler.wfile.write(body)


def make_error_payload(message: str, code: str, usage: UsageStats) -> dict[str, Any]:
    return {
        "ok": False,
        "error": {
            "code": code,
            "message": message,
        },
        "usage": usage.to_dict(),
    }
