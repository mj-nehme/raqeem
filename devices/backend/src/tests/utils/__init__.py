"""Test utilities package.

Provides a "formatters" attribute expected by service-layer tests that patch
these functions. The real implementations here are intentionally minimal; the
tests replace them with MagicMock, but presence avoids AttributeError.
"""

class _Formatters:
	def bytes_to_human(self, value: int) -> str:  # pragma: no cover
		for unit in ["B","KB","MB","GB","TB"]:
			if value < 1024 or unit == "TB":
				return f"{value/ (1024 ** (['B','KB','MB','GB','TB'].index(unit))):.1f} {unit}" if unit != "B" else f"{value} B"
			value /= 1024
		return f"{value:.1f} TB"

	def format_timestamp(self, dt) -> str:  # pragma: no cover
		return dt.strftime("%Y-%m-%dT%H:%M:%S")

	def format_percentage(self, val: float) -> str:  # pragma: no cover
		return f"{val:.1f}%"


formatters = _Formatters()

__all__ = ["formatters"]
