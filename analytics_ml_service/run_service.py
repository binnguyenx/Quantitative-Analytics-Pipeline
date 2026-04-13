"""CLI entry point for analytics ML service."""

from __future__ import annotations

import argparse
from pathlib import Path

from analytics_ml_service.backtest import run_walk_forward_backtest, write_backtest_logs
from analytics_ml_service.config import ServiceConfig
from analytics_ml_service.data import build_synthetic_dataset, clean_timeseries_data, load_timeseries_csv
from analytics_ml_service.features import build_feature_frame, get_feature_columns
from analytics_ml_service.train import save_model_and_metadata, train_xgboost


def parse_args() -> argparse.Namespace:
    """Parse CLI arguments."""
    parser = argparse.ArgumentParser(description="Run time-series forecasting analytics service.")
    parser.add_argument(
        "--input",
        type=Path,
        default=Path("analytics_ml_service/data/sample_timeseries.csv"),
        help="Path to input CSV. Must include target column and optional timestamp.",
    )
    parser.add_argument(
        "--generate-sample",
        action="store_true",
        help="Generate synthetic sample CSV if input does not exist.",
    )
    return parser.parse_args()


def main() -> None:
    """Execute feature pipeline, backtesting, and artifact logging."""
    args = parse_args()
    config = ServiceConfig()

    input_path: Path = args.input
    if args.generate_sample and not input_path.exists():
        build_synthetic_dataset(output_path=input_path, seed=config.random_seed)
        print(f"[INFO] Synthetic dataset generated at: {input_path}")

    if not input_path.exists():
        raise FileNotFoundError(
            f"Input data not found: {input_path}. Use --generate-sample or provide --input."
        )

    raw_df = load_timeseries_csv(input_path)
    clean_df = clean_timeseries_data(raw_df, config.target_col, config.timestamp_col)
    feature_df = build_feature_frame(
        clean_df,
        target_col=config.target_col,
        lags=config.lags,
        rolling_windows=config.rolling_windows,
        timestamp_col=config.timestamp_col,
    )
    feature_columns = get_feature_columns(feature_df, config.target_col, config.timestamp_col)

    fold_df, summary = run_walk_forward_backtest(feature_df, feature_columns, config)

    # Train final model on full available feature frame for serving.
    final_model = train_xgboost(
        feature_df[feature_columns].to_numpy(),
        feature_df[config.target_col].to_numpy(),
        config,
    )
    model_path, metadata_path = save_model_and_metadata(final_model, config, feature_columns, summary)
    fold_csv_path, summary_json_path = write_backtest_logs(fold_df, summary, config.logs_dir)

    print("[INFO] Walk-forward MAPE (before/after):")
    print(f"  baseline_mape  : {summary['baseline_mape']:.4f}")
    print(f"  xgboost_mape   : {summary['xgboost_mape']:.4f}")
    print(f"  delta_mape_abs : {summary['delta_mape_abs']:.4f}")
    print(f"  delta_mape_pct : {summary['delta_mape_pct']:.2f}%")
    print("[INFO] Saved artifacts/logs:")
    print(f"  model          : {model_path}")
    print(f"  metadata       : {metadata_path}")
    print(f"  fold_log_csv   : {fold_csv_path}")
    print(f"  summary_json   : {summary_json_path}")


if __name__ == "__main__":
    main()

