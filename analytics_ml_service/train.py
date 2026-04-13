"""Model training and artifact persistence utilities."""

from __future__ import annotations

import json
from dataclasses import asdict
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

import numpy as np
from xgboost import XGBRegressor

from analytics_ml_service.config import ServiceConfig


def train_xgboost(
    X_train: np.ndarray,
    y_train: np.ndarray,
    config: ServiceConfig,
) -> XGBRegressor:
    """Train an XGBoost regressor with reproducible seed."""
    model = XGBRegressor(
        **config.model.as_dict(),
        random_state=config.random_seed,
        n_jobs=-1,
    )
    model.fit(X_train, y_train)
    return model


def save_model_and_metadata(
    model: XGBRegressor,
    config: ServiceConfig,
    feature_columns: list[str],
    summary: dict[str, Any],
) -> tuple[Path, Path]:
    """Save model artifact and metadata JSON."""
    config.artifacts_dir.mkdir(parents=True, exist_ok=True)
    model_path = config.artifacts_dir / "xgboost_model.json"
    metadata_path = config.artifacts_dir / "model_metadata.json"

    model.save_model(model_path)
    metadata = {
        "saved_at_utc": datetime.now(UTC).isoformat(),
        "target_col": config.target_col,
        "timestamp_col": config.timestamp_col,
        "lags": list(config.lags),
        "rolling_windows": list(config.rolling_windows),
        "random_seed": config.random_seed,
        "model_config": asdict(config.model),
        "feature_columns": feature_columns,
        "evaluation_summary": summary,
    }
    metadata_path.write_text(json.dumps(metadata, indent=2), encoding="utf-8")
    return model_path, metadata_path

