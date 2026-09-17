from __future__ import annotations

import logging

import akshare as ak
import pandas as pd


def parse_params(query: dict[str, list[str]]) -> dict[str, str]:
    params: dict[str, str] = {}
    if "stock" in query:
        params["stock"] = query["stock"][0]
    if "market" in query:
        params["market"] = query["market"][0]
    return params


def execute(params: dict[str, str]) -> dict[str, object]:
    stock = params.get("stock", "000001")
    market = params.get("market", "sh")

    df: pd.DataFrame = ak.stock_individual_fund_flow(stock=stock, market=market)

    if df.empty:
        return {"fields": [], "total": 0, "items": []}

    fields = [
        "日期", "收盘价", "涨跌幅",
        "主力净流入-净额", "主力净流入-净占比",
        "超大单净流入-净额", "超大单净流入-净占比",
        "大单净流入-净额", "大单净流入-净占比",
        "中单净流入-净额", "中单净流入-净占比",
        "小单净流入-净额", "小单净流入-净占比",
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

    logging.info("query_fund_flow: fetched %d rows for %s", len(items), stock)
    return {"fields": available_fields, "total": len(items), "items": items}
