from __future__ import annotations

import json
import logging
import os
import socket
import threading
from dataclasses import asdict, dataclass
from datetime import date
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


class DailyLimitExceeded(RuntimeError):
    """每日调用限额已用尽"""


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
                raise DailyLimitExceeded("daily_limit_exceeded")

            self.stats.count += 1
            self._save_stats(self.stats)
            return UsageStats(
                date=self.stats.date,
                count=self.stats.count,
                limit=self.stats.limit,
            )


class BaostockQueryError(RuntimeError):
    """baostock 服务端返回的确定性错误(参数/限额等), 连接本身正常, 重试无意义"""


usage_counter = UsageCounter(USAGE_FILE, DAILY_LIMIT)

# 单一全局锁: 串行化所有 baostock SDK 调用(login/query/logout)。
# 旧实现用 baostock_lock + _session_lock 两把锁: force_disconnect 在持 _session_lock 时做
# 真实网络 logout, 而等待 login 的线程在持 baostock_lock 时等 _session_lock, 一旦 logout
# 卡住就形成锁序倒置, 全服务僵死。合并为一把可重入锁后, 全链路只有一种获取顺序, 消除死锁。
baostock_lock = threading.RLock()

# --- baostock 会话保持 ---
# endpoints 每个请求都 login/logout, 两次网络握手既慢又容易触发服务端卡顿;
# 包装 login/logout: 已登录时 login 复用会话, logout 不再真正登出, 查询异常时强制断开重连
_real_login = bs.login
_real_logout = bs.logout
_logged_in = False
_login_result: Any = None


def _patched_login() -> Any:
    global _logged_in, _login_result
    # 调用方(endpoint)已持有 baostock_lock, RLock 可重入; 独立调用时也能自我串行化
    with baostock_lock:
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
    """查询异常后调用: 强制断开会话, 下一个请求重新登录

    logout 仍在本锁内执行: SDK 只有一条全局连接, 若在锁外 logout, 会与锁内正在
    login/query 的线程并发操作同一连接, 制造新的状态错乱。死锁已由"合并为单锁"消除;
    若 logout 自身卡死, 交由健康检查看门狗(后续项)自愈。
    """
    global _logged_in, _login_result
    with baostock_lock:
        if not _logged_in and _login_result is None:
            return
        _logged_in = False
        _login_result = None
        try:
            _real_logout()
        except Exception as error:
            logging.error("force disconnect logout failed: %s", error)


bs.login = _patched_login
bs.logout = _patched_logout


def make_error_payload(message: str, code: str, usage: UsageStats) -> dict[str, Any]:
    return {
        "ok": False,
        "error": {
            "code": code,
            "message": message,
        },
        "usage": usage.to_dict(),
    }
