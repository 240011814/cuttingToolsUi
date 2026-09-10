from __future__ import annotations

import json
import os
import socket
import threading
from dataclasses import asdict, dataclass
from datetime import date
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler
from pathlib import Path
from typing import Any

# baostock 底层 socket 无超时, 服务端偶发卡顿会永久阻塞并占住全局锁, 导致所有请求连锁超时;
# 统一加 60s 超时兜底, 把挂死变成可重试的错误
socket.setdefaulttimeout(60)

try:
    import baostock as bs  # noqa: F401
except ModuleNotFoundError as error:
    raise SystemExit(
        "Missing Python dependency 'baostock'. "
        "Create/use apps/baostock-api/.venv and run "
        "'apps/baostock-api/.venv/bin/pip install -r apps/baostock-api/requirements.txt'."
    ) from error


HOST = os.environ.get("BAOSTOCK_API_HOST", "127.0.0.1")
PORT = int(os.environ.get("BAOSTOCK_API_PORT", "3002"))
DAILY_LIMIT = int(os.environ.get("BAOSTOCK_API_DAILY_LIMIT", "45000"))
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

# --- baostock 会话保持 ---
# endpoints 每个请求都 login/logout, 两次网络握手既慢又容易触发服务端卡顿;
# 包装 login/logout: 已登录时 login 复用会话, logout 不再真正登出, 查询异常时强制断开重连
_real_login = bs.login
_real_logout = bs.logout
_session_lock = threading.Lock()
_logged_in = False
_login_result: Any = None


def _patched_login() -> Any:
    global _logged_in, _login_result
    with _session_lock:
        if _logged_in:
            return _login_result
        result = _real_login()
        if result.error_code == "0":
            _logged_in = True
            _login_result = result
        return result


def _patched_logout() -> None:
    # endpoint 在 finally 中调用, 会话保持模式下不真正登出
    pass


def force_disconnect() -> None:
    """查询异常后调用: 强制断开会话, 下一个请求重新登录"""
    global _logged_in, _login_result
    with _session_lock:
        if _logged_in:
            try:
                _real_logout()
            except Exception:
                pass
        _logged_in = False
        _login_result = None


bs.login = _patched_login
bs.logout = _patched_logout


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
