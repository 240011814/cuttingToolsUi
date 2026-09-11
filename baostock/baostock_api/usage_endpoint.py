from __future__ import annotations

from .shared import usage_counter


def handle() -> dict[str, object]:
    stats = usage_counter.get_stats().to_dict()
    return {
        "ok": True,
        "data": stats,
        "usage": stats,
    }
