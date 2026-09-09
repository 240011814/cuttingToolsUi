from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryStockIndustryParams:
    code: str | None
    date: str | None


@dataclass(frozen=True)
class QueryStockIndustryResult:
    code: str | None
    date: str | None
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "code": self.code,
            "date": self.date,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QueryStockIndustryParams:
    return QueryStockIndustryParams(
        code=get_optional_query_param(query, "code"),
        date=get_optional_query_param(query, "date"),
    )


def execute(params: QueryStockIndustryParams) -> dict[str, object]:
    query_kwargs: dict[str, str] = {}

    if params.code is not None:
        query_kwargs["code"] = params.code
    if params.date is not None:
        query_kwargs["date"] = params.date

    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_stock_industry(**query_kwargs)

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_stock_industry failed")

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QueryStockIndustryResult(
        code=params.code,
        date=params.date,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
