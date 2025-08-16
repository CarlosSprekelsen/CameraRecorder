# tests/utils/test_helpers.py
"""
Test helper utilities for MediaMTX Camera Service.

Requirements Traceability:
- REQ-UTIL-006: Test helpers shall provide common test utility functions for real component testing
- REQ-UTIL-014: Test helpers shall support async test operations and timeouts
- REQ-UTIL-015: Test helpers shall provide WebSocket client testing utilities
- REQ-UTIL-016: Test helpers shall support test data generation and validation
- REQ-UTIL-017: Test helpers shall provide condition waiting and polling utilities

Story Coverage: All test stories requiring utility support
IV&V Control Point: Test utility validation and real component testing support

This module provides:
1. Async test utilities for real component testing
2. WebSocket client testing helpers
3. Test data generation and validation functions
4. Condition waiting and polling utilities
5. Common test patterns for real system validation
"""

import asyncio
import time
from typing import Any, Dict, Optional
