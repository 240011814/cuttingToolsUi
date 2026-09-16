from __future__ import annotations

import json
import logging
import os
import socket
import threading
import time
import zlib
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

import baostock.common.contants as bs_cons
import baostock.common.context as bs_context
import baostock.util.socketutil as bs_socketutil


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


# --- baostock 收包兜底 ---
# SDK socketutil.send_msg 的收包循环有个致命缺陷: recv() 返回 b"" (对端断开/半关闭) 时,
# receive 永远不增长, 尾部分隔符判断恒为假 -> 死循环空转(单核 100%, 4 核即 25% CPU),
# 而它是在持有 baostock_lock 时被调用的 -> 锁永不释放, 所有请求连锁挂死(只有 /health 能响应)。
# socket.setdefaulttimeout 救不了: 对端 EOF 后 socket 立即可读且返回 0 字节, recv 不再阻塞。
# 旧版 baostock 服务端重置长连接时若为半关闭(FIN)而非 RST, 就会走到这里, 即本次卡死成因。
# 这里替换为带 EOF 检测 + 总超时上限的版本, 并把异常抛出(而非 SDK 那样静默返回 None),
# 让 _execute_with_retry 能感知瞬时断连并 force_disconnect 后重试。
SEND_MSG_DEADLINE_SECONDS = 60
_RECV_CHUNK = 8192
_MESSAGE_END = b"<![CDATA[]]>\n"


def _patched_send_msg(msg: str) -> str:
    with baostock_lock:
        default_socket = getattr(bs_context, "default_socket", None)
        if default_socket is None:
            raise ConnectionError("baostock socket not connected (login required)")

        default_socket.send(bytes(msg + "\n", encoding="utf-8"))

        receive = b""
        deadline = time.monotonic() + SEND_MSG_DEADLINE_SECONDS
        while True:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError("baostock response deadline exceeded")
            default_socket.settimeout(remaining)
            recv = default_socket.recv(_RECV_CHUNK)
            if not recv:
                # 对端关闭连接: recv 会一直秒返回 b"", 必须在此终止, 否则空转占锁
                raise ConnectionError("baostock connection closed by peer")
            receive += recv
            if receive[-len(_MESSAGE_END) :] == _MESSAGE_END:
                break

        head_bytes = receive[0 : bs_cons.MESSAGE_HEADER_LENGTH]
        head_str = bytes.decode(head_bytes)
        head_arr = head_str.split(bs_cons.MESSAGE_SPLIT)
        if head_arr[1] in bs_cons.COMPRESSED_MESSAGE_TYPE_TUPLE:
            head_inner_length = int(head_arr[2])
            body_str = bytes.decode(
                zlib.decompress(
                    receive[
                        bs_cons.MESSAGE_HEADER_LENGTH : bs_cons.MESSAGE_HEADER_LENGTH + head_inner_length
                    ]
                )
            )
            return head_str + body_str
        return bytes.decode(receive)


bs_socketutil.send_msg = _patched_send_msg


def make_error_payload(message: str, code: str, usage: UsageStats) -> dict[str, Any]:
    return {
        "ok": False,
        "error": {
            "code": code,
            "message": message,
        },
        "usage": usage.to_dict(),
    }
