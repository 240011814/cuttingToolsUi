from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param, get_required_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryGrowthDataParams:
    code: str
    year: str | None
    quarter: str | None


@dataclass(frozen=True)
class QueryGrowthDataResult:
    code: str
    year: str | None
    quarter: str | None
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "code": self.code,
            "year": self.year,
            "quarter": self.quarter,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QueryGrowthDataParams:
    return QueryGrowthDataParams(
        code=get_required_query_param(query, "code"),
        year=get_optional_query_param(query, "year"),
        quarter=get_optional_query_param(query, "quarter"),
    )


def execute(params: QueryGrowthDataParams) -> dict[str, object]:
    query_kwargs = {
        "code": params.code,
    }

    if params.year is not None:
        query_kwargs["year"] = params.year
    if params.quarter is not None:
        query_kwargs["quarter"] = params.quarter

    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_growth_data(**query_kwargs)

            if result.error_code != "0":
                raise RuntimeError(result.error_msg or "query_growth_data failed")

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QueryGrowthDataResult(
        code=params.code,
        year=params.year,
        quarter=params.quarter,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
