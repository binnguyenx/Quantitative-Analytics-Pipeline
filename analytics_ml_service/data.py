"""Data loading and cleaning utilities."""

from __future__ import annotations

from pathlib import Path

import numpy as np
import pandas as pd


def load_timeseries_csv(input_path: Path) -> pd.DataFrame:
    """Load a CSV file into a DataFrame."""
    return pd.read_csv(input_path)


def clean_timeseries_data(
    df: pd.DataFrame,
    target_col: str,
    timestamp_col: str | None = None,
) -> pd.DataFrame:
    """Clean and sort timeseries data with deterministic ordering."""
    cleaned = df.copy()
    if timestamp_col and timestamp_col in cleaned.columns:
        cleaned[timestamp_col] = pd.to_datetime(cleaned[timestamp_col], errors="coerce")
        cleaned = cleaned.dropna(subset=[timestamp_col])
        cleaned = cleaned.sort_values(timestamp_col)
        cleaned = cleaned.drop_duplicates(subset=[timestamp_col], keep="last")
    else:
        cleaned = cleaned.reset_index(drop=True)

    cleaned[target_col] = pd.to_numeric(cleaned[target_col], errors="coerce")
    cleaned = cleaned.dropna(subset=[target_col])
    return cleaned.reset_index(drop=True)


def build_synthetic_dataset(
    output_path: Path,
    n_rows: int = 180,
    seed: int = 42,
) -> Path:
    """Generate a synthetic daily dataset for quick end-to-end runs."""
    rng = np.random.default_rng(seed)
    dates = pd.date_range("2023-01-01", periods=n_rows, freq="D")
    trend = np.linspace(100.0, 145.0, n_rows)
    seasonality = 5 * np.sin(np.arange(n_rows) * 2 * np.pi / 30)
    noise = rng.normal(0.0, 1.8, n_rows)
    values = trend + seasonality + noise

    sample_df = pd.DataFrame(
        {
            "timestamp": dates,
            "target": values.round(4),
        }
    )

    output_path.parent.mkdir(parents=True, exist_ok=True)
    sample_df.to_csv(output_path, index=False)
    return output_path

