from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryStockBasicParams:
    code: str | None
    code_name: str | None


@dataclass(frozen=True)
class QueryStockBasicResult:
    code: str | None
    code_name: str | None
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "code": self.code,
            "code_name": self.code_name,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QueryStockBasicParams:
    return QueryStockBasicParams(
        code=get_optional_query_param(query, "code"),
        code_name=get_optional_query_param(query, "code_name"),
    )


def execute(params: QueryStockBasicParams) -> dict[str, object]:
    query_kwargs: dict[str, str] = {}

    if params.code is not None:
        query_kwargs["code"] = params.code
    if params.code_name is not None:
        query_kwargs["code_name"] = params.code_name

    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_stock_basic(**query_kwargs)

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_stock_basic failed")

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QueryStockBasicResult(
        code=params.code,
        code_name=params.code_name,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
