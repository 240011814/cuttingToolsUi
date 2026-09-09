from __future__ import annotations

from dataclasses import dataclass

from .request import get_optional_query_param
from .shared import baostock_lock, bs


@dataclass(frozen=True)
class QueryRequiredReserveRatioDataParams:
    start_date: str | None
    end_date: str | None
    year_type: str | None


@dataclass(frozen=True)
class QueryRequiredReserveRatioDataResult:
    start_date: str | None
    end_date: str | None
    year_type: str | None
    fields: list[str]
    total: int
    items: list[dict[str, str]]

    def to_dict(self) -> dict[str, object]:
        return {
            "start_date": self.start_date,
            "end_date": self.end_date,
            "year_type": self.year_type,
            "fields": self.fields,
            "total": self.total,
            "items": self.items,
        }


def parse_params(query: dict[str, list[str]]) -> QueryRequiredReserveRatioDataParams:
    return QueryRequiredReserveRatioDataParams(
        start_date=get_optional_query_param(query, "start_date"),
        end_date=get_optional_query_param(query, "end_date"),
        year_type=get_optional_query_param(query, "year_type"),
    )


def execute(params: QueryRequiredReserveRatioDataParams) -> dict[str, object]:
    query_kwargs: dict[str, str] = {}

    if params.start_date is not None:
        query_kwargs["start_date"] = params.start_date
    if params.end_date is not None:
        query_kwargs["end_date"] = params.end_date
    if params.year_type is not None:
        query_kwargs["yearType"] = params.year_type

    with baostock_lock:
        login_result = bs.login()

        if login_result.error_code != "0":
            raise RuntimeError(f"baostock login failed: {login_result.error_msg}")

        try:
            result = bs.query_required_reserve_ratio_data(**query_kwargs)

            if result.error_code != "0":
                raise RuntimeError(
                    result.error_msg or "query_required_reserve_ratio_data failed"
                )

            items: list[dict[str, str]] = []
            while result.next():
                items.append(dict(zip(result.fields, result.get_row_data())))
        finally:
            bs.logout()

    payload = QueryRequiredReserveRatioDataResult(
        start_date=params.start_date,
        end_date=params.end_date,
        year_type=params.year_type,
        fields=list(result.fields),
        total=len(items),
        items=items,
    )
    return payload.to_dict()
