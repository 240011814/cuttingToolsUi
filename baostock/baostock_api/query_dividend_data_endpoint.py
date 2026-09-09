from __future__ import annotations

from dataclasses import dataclass

from .request import get_required_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryDividendDataParams:
    code: str
    year: str
    year_type: str


@dataclass(frozen=True)
class QueryDividendDataResult:
    code: str
    year: str
    year_type: str
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "code": self.code,
            "year": self.year,
            "year_type": self.year_type,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QueryDividendDataParams:
    return QueryDividendDataParams(
        code=get_required_query_param(query, "code"),
        year=get_required_query_param(query, "year"),
        year_type=get_required_query_param(query, "year_type"),
    )


def execute(params: QueryDividendDataParams) -> dict[str, object]:
    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_dividend_data(
                code=params.code,
                year=params.year,
                yearType=params.year_type,
            )

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_dividend_data failed")

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QueryDividendDataResult(
        code=params.code,
        year=params.year,
        year_type=params.year_type,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
