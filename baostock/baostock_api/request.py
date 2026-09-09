from __future__ import annotations


def get_required_query_param(query: dict[str, list[str]], name: str) -> str:
    value = query.get(name, [None])[0]
    if value is None or value == "":
        raise ValueError(f"missing required query parameter: {name}")
    return value


def get_optional_query_param(query: dict[str, list[str]], name: str) -> str | None:
    value = query.get(name, [None])[0]
    if value is None or value == "":
        return None
    return value
