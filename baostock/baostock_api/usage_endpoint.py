from __future__ import annotations

from http import HTTPStatus
from http.server import BaseHTTPRequestHandler

from .shared import json_response, usage_counter


def handle(handler: BaseHTTPRequestHandler) -> None:
    json_response(
        handler,
        HTTPStatus.OK,
        {
            "ok": True,
            "data": usage_counter.get_stats().to_dict(),
            "usage": usage_counter.get_stats().to_dict(),
        },
    )
