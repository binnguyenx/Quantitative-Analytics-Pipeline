"""Evaluation metrics for forecasting."""

from __future__ import annotations

import numpy as np


def mape(y_true: np.ndarray, y_pred: np.ndarray, epsilon: float = 1e-8) -> float:
    """Compute Mean Absolute Percentage Error in percentage points."""
    y_true_arr = np.asarray(y_true, dtype=float)
    y_pred_arr = np.asarray(y_pred, dtype=float)
    denom = np.clip(np.abs(y_true_arr), epsilon, None)
    return float(np.mean(np.abs((y_true_arr - y_pred_arr) / denom)) * 100.0)

