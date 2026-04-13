"""Configuration objects for analytics ML service."""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass(slots=True)
class ModelConfig:
    """Hyperparameters for XGBoost regressor."""

    n_estimators: int = 300
    learning_rate: float = 0.05
    max_depth: int = 4
    subsample: float = 0.9
    colsample_bytree: float = 0.9
    reg_lambda: float = 1.0
    objective: str = "reg:squarederror"

    def as_dict(self) -> dict[str, Any]:
        """Return model parameters as dict for estimator init."""
        return {
            "n_estimators": self.n_estimators,
            "learning_rate": self.learning_rate,
            "max_depth": self.max_depth,
            "subsample": self.subsample,
            "colsample_bytree": self.colsample_bytree,
            "reg_lambda": self.reg_lambda,
            "objective": self.objective,
        }


@dataclass(slots=True)
class ServiceConfig:
    """Top-level configuration for pipeline, backtesting and outputs."""

    target_col: str = "target"
    timestamp_col: str | None = "timestamp"
    lags: tuple[int, ...] = (1, 2, 3, 7, 14)
    rolling_windows: tuple[int, ...] = (3, 7, 14)
    min_train_size: int = 48
    test_size: int = 12
    step_size: int = 12
    random_seed: int = 42
    artifacts_dir: Path = field(default_factory=lambda: Path("analytics_ml_service/artifacts"))
    logs_dir: Path = field(default_factory=lambda: Path("analytics_ml_service/logs"))
    model: ModelConfig = field(default_factory=ModelConfig)

