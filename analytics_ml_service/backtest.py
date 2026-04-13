"""Walk-forward validation logic and experiment logging."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import numpy as np
import pandas as pd

from analytics_ml_service.config import ServiceConfig
from analytics_ml_service.metrics import mape


@dataclass(slots=True)
class FoldSplit:
    """Index ranges for one walk-forward fold."""

    train_start: int
    train_end: int
    test_start: int
    test_end: int


def build_walk_forward_splits(
    n_rows: int,
    min_train_size: int,
    test_size: int,
    step_size: int,
) -> list[FoldSplit]:
    """Build expanding-window walk-forward folds."""
    if min_train_size <= 0 or test_size <= 0 or step_size <= 0:
        raise ValueError("min_train_size, test_size and step_size must be positive")
    if n_rows < min_train_size + test_size:
        raise ValueError("Not enough rows for one fold")

    splits: list[FoldSplit] = []
    test_start = min_train_size
    while test_start + test_size <= n_rows:
        splits.append(
            FoldSplit(
                train_start=0,
                train_end=test_start,
                test_start=test_start,
                test_end=test_start + test_size,
            )
        )
        test_start += step_size
    return splits


def run_walk_forward_backtest(
    feature_df: pd.DataFrame,
    feature_columns: list[str],
    config: ServiceConfig,
) -> tuple[pd.DataFrame, dict[str, float]]:
    """Run baseline vs XGBoost walk-forward validation."""
    from analytics_ml_service.train import train_xgboost

    splits = build_walk_forward_splits(
        n_rows=len(feature_df),
        min_train_size=config.min_train_size,
        test_size=config.test_size,
        step_size=config.step_size,
    )
    fold_rows: list[dict[str, Any]] = []
    for fold_no, split in enumerate(splits, start=1):
        train_slice = feature_df.iloc[split.train_start : split.train_end]
        test_slice = feature_df.iloc[split.test_start : split.test_end]

        X_train = train_slice[feature_columns].to_numpy()
        y_train = train_slice[config.target_col].to_numpy()
        X_test = test_slice[feature_columns].to_numpy()
        y_test = test_slice[config.target_col].to_numpy()

        model = train_xgboost(X_train, y_train, config)
        xgb_pred = model.predict(X_test)

        lag_1_col = f"{config.target_col}_lag_1"
        if lag_1_col in test_slice.columns:
            baseline_pred = test_slice[lag_1_col].to_numpy()
        else:
            baseline_pred = np.full_like(y_test, fill_value=float(np.mean(y_train)))

        baseline_score = mape(y_test, baseline_pred)
        xgb_score = mape(y_test, xgb_pred)
        fold_rows.append(
            {
                "fold": fold_no,
                "train_start": split.train_start,
                "train_end": split.train_end,
                "test_start": split.test_start,
                "test_end": split.test_end,
                "baseline_mape": round(baseline_score, 6),
                "xgboost_mape": round(xgb_score, 6),
            }
        )

    fold_df = pd.DataFrame(fold_rows)
    baseline_avg = float(fold_df["baseline_mape"].mean())
    xgb_avg = float(fold_df["xgboost_mape"].mean())
    delta_abs = baseline_avg - xgb_avg
    delta_pct = (delta_abs / baseline_avg * 100.0) if baseline_avg else 0.0
    summary = {
        "baseline_mape": round(baseline_avg, 6),
        "xgboost_mape": round(xgb_avg, 6),
        "delta_mape_abs": round(delta_abs, 6),
        "delta_mape_pct": round(delta_pct, 6),
    }
    return fold_df, summary


def write_backtest_logs(
    fold_df: pd.DataFrame,
    summary: dict[str, float],
    logs_dir: Path,
) -> tuple[Path, Path]:
    """Write fold-level and summary metrics to CSV/JSON."""
    logs_dir.mkdir(parents=True, exist_ok=True)
    fold_csv_path = logs_dir / "walk_forward_folds.csv"
    summary_json_path = logs_dir / "mape_summary.json"

    fold_df.to_csv(fold_csv_path, index=False)
    summary_json_path.write_text(json.dumps(summary, indent=2), encoding="utf-8")
    return fold_csv_path, summary_json_path

