"""baostock-api package."""

from __future__ import annotations

from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover
    tomllib = None


def _load_version() -> str:
    pyproject_path = Path(__file__).resolve().parent.parent / "pyproject.toml"

    if tomllib is None or not pyproject_path.exists():
        return "1.0.0"

    try:
        payload = tomllib.loads(pyproject_path.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return "1.0.0"

    project = payload.get("project")
    if not isinstance(project, dict):
        return "1.0.0"

    version = project.get("version")
    return str(version) if version else "1.0.0"


__version__ = _load_version()
