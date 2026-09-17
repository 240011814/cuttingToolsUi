from __future__ import annotations

import logging

import akshare as ak
import pandas as pd


def parse_params(query: dict[str, list[str]]) -> dict[str, str]:
    params: dict[str, str] = {}
    if "symbol" in query:
        params["symbol"] = query["symbol"][0]
    if "start_date" in query:
        params["start_date"] = query["start_date"][0]
    if "end_date" in query:
        params["end_date"] = query["end_date"][0]
    return params


def execute(params: dict[str, str]) -> dict[str, object]:
    symbol = params.get("symbol", "全部")
    start_date = params.get("start_date", "")
    end_date = params.get("end_date", "")

    if not start_date or not end_date:
        raise ValueError("start_date and end_date are required")

    df: pd.DataFrame = ak.stock_lhb_detail_em(
        symbol=symbol,
        start_date=start_date,
        end_date=end_date,
    )

    if df.empty:
        return {"fields": [], "total": 0, "items": []}

    fields = [
        "序号", "代码", "名称", "收盘价", "涨跌幅", "龙虎榜净买额",
        "龙虎榜买入额", "龙虎榜卖出额", "龙虎榜成交额", "市场总成交额",
        "净买额占总成交比", "成交额占总成交比", "换手率", "流通市值",
        "上榜原因", "上榜日期",
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

    logging.info("query_lhb_detail: fetched %d rows", len(items))
    return {"fields": available_fields, "total": len(items), "items": items}
