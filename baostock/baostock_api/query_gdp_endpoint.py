from __future__ import annotations

import logging

import akshare as ak
import pandas as pd


def parse_params(query: dict[str, list[str]]) -> dict[str, str]:
    return {}


def execute(params: dict[str, str]) -> dict[str, object]:
    df: pd.DataFrame = ak.macro_china_gdp()

    if df.empty:
        return {"fields": [], "total": 0, "items": []}

    fields = ["季度", "国内生产总值-绝对值", "国内生产总值-同比增长"]

    available_fields = [f for f in fields if f in df.columns]
    items: list[dict[str, str]] = []

    for _, row in df.iterrows():
        item: dict[str, str] = {}
        for col in available_fields:
            val = row.get(col)
            if pd.isna(val):
                item[col] = ""
            elif isinstance(val, pd.Timestamp):
                item[col] = val.strftime("%Y-%m-%d")
            else:
                item[col] = str(val)
        items.append(item)

    logging.info("query_gdp: fetched %d rows", len(items))
    return {"fields": available_fields, "total": len(items), "items": items}
