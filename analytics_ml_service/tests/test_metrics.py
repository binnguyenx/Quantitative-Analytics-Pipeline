"""Unit tests for metrics."""

import unittest

import numpy as np

from analytics_ml_service.metrics import mape


class TestMetrics(unittest.TestCase):
    """Validate MAPE implementation."""

    def test_mape_basic_case(self) -> None:
        y_true = np.array([100.0, 200.0, 400.0])
        y_pred = np.array([90.0, 220.0, 380.0])
        score = mape(y_true, y_pred)
        self.assertAlmostEqual(score, 8.333333, places=5)

    def test_mape_handles_zero_with_epsilon(self) -> None:
        y_true = np.array([0.0, 50.0])
        y_pred = np.array([10.0, 40.0])
        score = mape(y_true, y_pred)
        self.assertTrue(np.isfinite(score))


if __name__ == "__main__":
    unittest.main()

