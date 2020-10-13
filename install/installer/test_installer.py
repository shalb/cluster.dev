#!/usr/bin/env python3
"""
Non-interactive tests for installer.

pytest files should start from `test_`
"""
import cluster_dev as installer  # pylint: disable=unused-import # noqa: WPS110, F401

# Example usage:


def func(test_param):
    """Test something."""
    return test_param + 1


def test_func():
    """Function for test func().

    Test function should start from `test_`
    """
    assert func(3) == 5  # noqa: S101
