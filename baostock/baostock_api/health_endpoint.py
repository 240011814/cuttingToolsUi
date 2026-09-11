from __future__ import annotations

from .shared import usage_counter


def handle() -> dict[str, object]:
    return {
        "ok": True,
        "data": {
            "status": "ok",
            "service": "baostock-api",
        },
        "usage": usage_counter.get_stats().to_dict(),
    }
