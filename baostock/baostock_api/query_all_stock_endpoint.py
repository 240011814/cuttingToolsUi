from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryAllStockParams:
    day: str | None


@dataclass(frozen=True)
class BaostockAllStockItem:
    code: str
    code_name: str
    trade_status: str

    def to_dict(self) -> dict[str, str]:
        return {
            "code": self.code,
            "code_name": self.code_name,
            "trade_status": self.trade_status,
        }


def parse_params(query: dict[str, list[str]]) -> QueryAllStockParams:
    return QueryAllStockParams(day=get_optional_query_param(query, "day"))


def execute(params: QueryAllStockParams) -> dict[str, object]:
    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_all_stock(day=params.day)

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_all_stock failed")

            items: list[BaostockAllStockItem] = []
            while result.next():
                row = dict(zip(result.fields, result.get_row_data()))
                items.append(
                    BaostockAllStockItem(
                        code=row.get("code", ""),
                        code_name=row.get("code_name", ""),
                        trade_status=row.get("tradeStatus", ""),
                    )
                )
        finally:
            bs.logout()

    return {
        "day": params.day,
        "total": len(items),
        "items": [item.to_dict() for item in items],
    }
