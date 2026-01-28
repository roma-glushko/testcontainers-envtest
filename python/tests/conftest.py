"""Pytest configuration and fixtures."""

import os

import pytest


def pytest_configure(config):
    """Configure pytest markers."""
    config.addinivalue_line(
        "markers", "integration: marks tests as integration tests (require Docker)"
    )


def pytest_collection_modifyitems(config, items):
    """Skip integration tests if SKIP_INTEGRATION is set."""
    if os.environ.get("SKIP_INTEGRATION"):
        skip_integration = pytest.mark.skip(reason="SKIP_INTEGRATION is set")
        for item in items:
            if "integration" in item.keywords:
                item.add_marker(skip_integration)
