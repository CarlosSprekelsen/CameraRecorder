"""
Integration test for Python SDK CLI functionality.

Requirements Coverage:
- REQ-CLIENT-044: Command-line interface for camera operations
"""

import pytest
import subprocess
import sys
from pathlib import Path


@pytest.mark.integration
class TestSDKCLI:
    """Integration test for SDK CLI functionality."""
    
    @pytest.fixture
    def cli_path(self):
        """Path to the CLI module."""
        sdk_path = Path(__file__).parent.parent.parent / "sdk" / "python"
        return sdk_path / "mediamtx_camera_sdk" / "cli.py"
    
    def test_cli_help_displays_correctly(self, cli_path):
        """Test that CLI help displays correctly without async errors."""
        try:
            result = subprocess.run(
                [sys.executable, "-m", "mediamtx_camera_sdk.cli", "--help"],
                capture_output=True,
                text=True,
                cwd=cli_path.parent.parent,  # sdk/python directory
                timeout=10
            )
            
            # Should exit with success (0)
            assert result.returncode == 0, f"CLI help failed: {result.stderr}"
            
            # Should display help information
            assert "MediaMTX Camera Service CLI" in result.stdout
            assert "usage:" in result.stdout
            assert "Examples:" in result.stdout
            
            # Should not have any RuntimeWarning about coroutines
            assert "RuntimeWarning" not in result.stderr
            assert "coroutine" not in result.stderr.lower()
            
        except subprocess.TimeoutExpired:
            pytest.fail("CLI help command timed out")
        except Exception as e:
            pytest.fail(f"CLI help test failed: {e}")
    
    def test_cli_without_arguments_shows_help(self, cli_path):
        """Test that CLI without arguments shows help."""
        try:
            result = subprocess.run(
                [sys.executable, "-m", "mediamtx_camera_sdk.cli"],
                capture_output=True,
                text=True,
                cwd=cli_path.parent.parent,  # sdk/python directory
                timeout=10
            )
            
            # Should exit with error (2) for missing required arguments
            assert result.returncode == 2, f"CLI should exit with error: {result.stdout}"
            
            # Should display help information
            assert "usage:" in result.stderr
            assert "error:" in result.stderr
            
            # Should not have any RuntimeWarning about coroutines
            assert "RuntimeWarning" not in result.stderr
            assert "coroutine" not in result.stderr.lower()
            
        except subprocess.TimeoutExpired:
            pytest.fail("CLI command timed out")
        except Exception as e:
            pytest.fail(f"CLI test failed: {e}")
    
    def test_cli_invalid_command_shows_error(self, cli_path):
        """Test that CLI with invalid command shows appropriate error."""
        try:
            result = subprocess.run(
                [sys.executable, "-m", "mediamtx_camera_sdk.cli", "invalid_command"],
                capture_output=True,
                text=True,
                cwd=cli_path.parent.parent,  # sdk/python directory
                timeout=10
            )
            
            # Should exit with error (1)
            assert result.returncode == 1, f"CLI should exit with error: {result.stdout}"
            
            # Should show error about missing token (since no auth provided)
            assert "token required" in result.stderr or "error:" in result.stderr
            
            # Should not have any RuntimeWarning about coroutines
            assert "RuntimeWarning" not in result.stderr
            assert "coroutine" not in result.stderr.lower()
            
        except subprocess.TimeoutExpired:
            pytest.fail("CLI command timed out")
        except Exception as e:
            pytest.fail(f"CLI test failed: {e}")
    
    def test_cli_missing_auth_token_shows_error(self, cli_path):
        """Test that CLI with missing auth token shows appropriate error."""
        try:
            result = subprocess.run(
                [sys.executable, "-m", "mediamtx_camera_sdk.cli", "list"],
                capture_output=True,
                text=True,
                cwd=cli_path.parent.parent,  # sdk/python directory
                timeout=10
            )
            
            # Should exit with error (1)
            assert result.returncode == 1, f"CLI should exit with error: {result.stdout}"
            
            # Should show error about missing token
            assert "token required" in result.stderr or "error:" in result.stderr
            
            # Should not have any RuntimeWarning about coroutines
            assert "RuntimeWarning" not in result.stderr
            assert "coroutine" not in result.stderr.lower()
            
        except subprocess.TimeoutExpired:
            pytest.fail("CLI command timed out")
        except Exception as e:
            pytest.fail(f"CLI test failed: {e}")
    
    def test_cli_sync_wrapper_function_exists(self, cli_path):
        """Test that the synchronous wrapper function exists and is callable."""
        import sys
        sys.path.insert(0, str(cli_path.parent.parent))
        
        try:
            from mediamtx_camera_sdk.cli import cli_main
            
            # Verify the function exists and is callable
            assert callable(cli_main), "cli_main should be callable"
            
            # Verify it's not an async function
            import inspect
            assert not inspect.iscoroutinefunction(cli_main), "cli_main should not be async"
            
        except ImportError as e:
            pytest.fail(f"Could not import cli_main: {e}")
    
    def test_cli_entry_point_configuration(self, cli_path):
        """Test that the entry point configuration is correct."""
        setup_py_path = cli_path.parent.parent / "setup.py"
        
        if not setup_py_path.exists():
            pytest.skip("setup.py not found")
        
        # Read setup.py and check entry point configuration
        with open(setup_py_path, 'r') as f:
            setup_content = f.read()
        
        # Should reference cli_main, not main
        assert "mediamtx-camera-cli=mediamtx_camera_sdk.cli:cli_main" in setup_content
        assert "mediamtx-camera-cli=mediamtx_camera_sdk.cli:main" not in setup_content
    
    def test_cli_imports_correctly(self, cli_path):
        """Test that the CLI module imports correctly."""
        import sys
        sys.path.insert(0, str(cli_path.parent.parent))
        
        try:
            # Should be able to import the CLI module
            import mediamtx_camera_sdk.cli
            
            # Should have both main and cli_main functions
            assert hasattr(mediamtx_camera_sdk.cli, 'main'), "main function should exist"
            assert hasattr(mediamtx_camera_sdk.cli, 'cli_main'), "cli_main function should exist"
            
            # main should be async, cli_main should be sync
            import inspect
            assert inspect.iscoroutinefunction(mediamtx_camera_sdk.cli.main), "main should be async"
            assert not inspect.iscoroutinefunction(mediamtx_camera_sdk.cli.cli_main), "cli_main should be sync"
            
        except ImportError as e:
            pytest.fail(f"Could not import CLI module: {e}")
    
    def test_cli_direct_execution_works(self, cli_path):
        """Test that CLI can be executed directly as a module."""
        try:
            result = subprocess.run(
                [sys.executable, "-m", "mediamtx_camera_sdk.cli", "--help"],
                capture_output=True,
                text=True,
                cwd=cli_path.parent.parent,  # sdk/python directory
                timeout=10
            )
            
            # Should exit with success (0)
            assert result.returncode == 0, f"Direct CLI execution failed: {result.stderr}"
            
            # Should display help information
            assert "MediaMTX Camera Service CLI" in result.stdout
            
            # Should not have any RuntimeWarning about coroutines
            assert "RuntimeWarning" not in result.stderr
            assert "coroutine" not in result.stderr.lower()
            
        except subprocess.TimeoutExpired:
            pytest.fail("Direct CLI execution timed out")
        except Exception as e:
            pytest.fail(f"Direct CLI execution test failed: {e}")
