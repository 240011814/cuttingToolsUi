from __future__ import annotations

import logging

import akshare as ak
import pandas as pd


def parse_params(query: dict[str, list[str]]) -> dict[str, str]:
    params: dict[str, str] = {}
    if "symbol" in query:
        params["symbol"] = query["symbol"][0]
    return params


def execute(params: dict[str, str]) -> dict[str, object]:
    symbol = params.get("symbol", "沪股通")

    df: pd.DataFrame = ak.stock_hsgt_north_net_flow_in_em(symbol=symbol)

    if df.empty:
        return {"fields": [], "total": 0, "items": []}

    fields = [
        "date", "value",
    ]

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

    logging.info("query_north_flow: fetched %d rows for %s", len(items), symbol)
    return {"fields": available_fields, "total": len(items), "items": items}
