"""Unit tests for walk-forward split logic."""

import unittest

from analytics_ml_service.backtest import build_walk_forward_splits


class TestWalkForwardSplits(unittest.TestCase):
    """Validate fold generation behavior."""

    def test_split_generation_expected_ranges(self) -> None:
        splits = build_walk_forward_splits(
            n_rows=30,
            min_train_size=12,
            test_size=6,
            step_size=6,
        )
        self.assertEqual(len(splits), 3)
        self.assertEqual((splits[0].train_start, splits[0].train_end), (0, 12))
        self.assertEqual((splits[0].test_start, splits[0].test_end), (12, 18))
        self.assertEqual((splits[2].test_start, splits[2].test_end), (24, 30))

    def test_invalid_row_count_raises(self) -> None:
        with self.assertRaises(ValueError):
            build_walk_forward_splits(
                n_rows=10,
                min_train_size=8,
                test_size=4,
                step_size=2,
            )


if __name__ == "__main__":
    unittest.main()

