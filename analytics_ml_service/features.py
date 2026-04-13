"""Feature engineering utilities for time-series forecasting."""

from __future__ import annotations

import pandas as pd


def build_feature_frame(
    df: pd.DataFrame,
    target_col: str,
    lags: tuple[int, ...],
    rolling_windows: tuple[int, ...],
    timestamp_col: str | None = None,
) -> pd.DataFrame:
    """
    Build leakage-safe features.

    Note: rolling statistics are computed on shifted target (t-1) to avoid
    peeking into current timestep.
    """
    frame = df.copy()
    for lag in lags:
        frame[f"{target_col}_lag_{lag}"] = frame[target_col].shift(lag)

    shifted_target = frame[target_col].shift(1)
    for window in rolling_windows:
        frame[f"{target_col}_roll_mean_{window}"] = shifted_target.rolling(window=window).mean()
        frame[f"{target_col}_roll_std_{window}"] = shifted_target.rolling(window=window).std()

    if timestamp_col and timestamp_col in frame.columns:
        ts = pd.to_datetime(frame[timestamp_col], errors="coerce")
        frame["month"] = ts.dt.month
        frame["day"] = ts.dt.day
        frame["dayofweek"] = ts.dt.dayofweek
        frame["dayofyear"] = ts.dt.dayofyear
        frame["is_month_start"] = ts.dt.is_month_start.astype(int)
        frame["is_month_end"] = ts.dt.is_month_end.astype(int)

    return frame.dropna().reset_index(drop=True)


def get_feature_columns(df: pd.DataFrame, target_col: str, timestamp_col: str | None = None) -> list[str]:
    """Return model feature columns by excluding target and timestamp."""
    exclude = {target_col}
    if timestamp_col:
        exclude.add(timestamp_col)
    return [col for col in df.columns if col not in exclude]

