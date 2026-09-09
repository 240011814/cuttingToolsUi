from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QuerySz50StocksParams:
    date: str | None


@dataclass(frozen=True)
class QuerySz50StocksResult:
    date: str | None
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "date": self.date,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QuerySz50StocksParams:
    return QuerySz50StocksParams(date=get_optional_query_param(query, "date"))


def execute(params: QuerySz50StocksParams) -> dict[str, object]:
    query_kwargs: dict[str, str] = {}

    if params.date is not None:
        query_kwargs["date"] = params.date

    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_sz50_stocks(**query_kwargs)

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_sz50_stocks failed")

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QuerySz50StocksResult(
        date=params.date,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
