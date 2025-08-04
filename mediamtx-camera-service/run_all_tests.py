#!/usr/bin/env python3
"""
Comprehensive test automation script for MediaMTX Camera Service.

Runs all quality gates in sequence: formatting, linting, type checking, 
unit tests, and integration tests with coverage measurement and reporting.

Usage:
    python3 run_all_tests.py                    # Run all stages
    python3 run_all_tests.py --no-lint          # Skip linting
    python3 run_all_tests.py --only-unit        # Unit tests only
    python3 run_all_tests.py --threshold=85     # Custom coverage threshold
    python3 run_all_tests.py --help             # Show all options

Exit codes:
    0 - All stages passed
    1 - One or more stages failed
    2 - Environment/setup error
"""

import sys
import subprocess
import argparse
import os
import platform
import shutil
import time
import json
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Optional, Tuple, Any


class TestStage:
    """Represents a single test stage with execution details."""
    
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.start_time: Optional[float] = None
        self.end_time: Optional[float] = None
        self.return_code: Optional[int] = None
        self.output: str = ""
        self.error_output: str = ""
        self.skipped: bool = False
        
    @property
    def duration(self) -> float:
        """Get stage duration in seconds."""
        if self.start_time and self.end_time:
            return self.end_time - self.start_time
        return 0.0
        
    @property
    def status(self) -> str:
        """Get stage status string."""
        if self.skipped:
            return "SKIPPED"
        elif self.return_code is None:
            return "NOT_RUN"
        elif self.return_code == 0:
            return "PASSED"
        else:
            return "FAILED"


class TestRunner:
    """Main test runner orchestrating all quality gates."""
    
    def __init__(self, args: argparse.Namespace):
        self.args = args
        self.project_root = Path(__file__).parent
        self.artifacts_dir = self._create_artifacts_dir()
        self.stages: List[TestStage] = []
        self.overall_start_time = time.time()
        
        # Detect virtual environment
        self.venv_active = self._detect_virtual_environment()
        
    def _create_artifacts_dir(self) -> Path:
        """Create timestamped artifacts directory."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        artifacts_dir = self.project_root / "artifacts" / timestamp
        artifacts_dir.mkdir(parents=True, exist_ok=True)
        return artifacts_dir
        
    def _detect_virtual_environment(self) -> bool:
        """Detect if running in virtual environment."""
        return (
            hasattr(sys, 'real_prefix') or
            (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix) or
            os.environ.get('VIRTUAL_ENV') is not None
        )
        
    def _setup_test_environment(self) -> bool:
        """Setup test environment and validate dependencies."""
        # Add project root to Python path
        sys.path.insert(0, str(self.project_root))
        
        # Create test directory structure if needed
        test_dirs = [
            "tests",
            "tests/unit",
            "tests/integration", 
            "tests/mocks"
        ]
        
        for dir_path in test_dirs:
            test_dir = self.project_root / dir_path
            test_dir.mkdir(exist_ok=True)
            
            # Create __init__.py files
            init_file = test_dir / "__init__.py"
            if not init_file.exists():
                init_file.write_text("# Test package\n")
        
        # Validate required tools
        required_tools = ["python", "black", "flake8", "mypy", "pytest"]
        missing_tools = []
        
        for tool in required_tools:
            if not shutil.which(tool):
                missing_tools.append(tool)
        
        if missing_tools:
            print(f"ERROR: Missing required tools: {', '.join(missing_tools)}")
            print("Install with: pip install -r requirements-dev.txt")
            return False
            
        return True
        
    def _run_command(
        self, 
        cmd: List[str], 
        cwd: Optional[Path] = None,
        capture_output: bool = True
    ) -> subprocess.CompletedProcess:
        """Run command with proper error handling and logging."""
        if cwd is None:
            cwd = self.project_root
            
        if self.args.verbose:
            print(f"Running: {' '.join(cmd)}")
            print(f"Working directory: {cwd}")
            
        try:
            # Use shell=True on Windows for better compatibility
            use_shell = platform.system() == "Windows"
            
            result = subprocess.run(
                cmd,
                cwd=cwd,
                capture_output=capture_output,
                text=True,
                shell=use_shell,
                timeout=300  # 5 minute timeout
            )
            return result
            
        except subprocess.TimeoutExpired:
            print(f"ERROR: Command timed out: {' '.join(cmd)}")
            # Return failed result
            return subprocess.CompletedProcess(cmd, 124, "", "Command timed out")
        except FileNotFoundError as e:
            print(f"ERROR: Command not found: {' '.join(cmd)} - {e}")
            return subprocess.CompletedProcess(cmd, 127, "", str(e))
        except Exception as e:
            print(f"ERROR: Failed to run command: {' '.join(cmd)} - {e}")
            return subprocess.CompletedProcess(cmd, 1, "", str(e))
            
    def _run_stage(self, stage: TestStage, cmd: List[str]) -> bool:
        """Run a single test stage and record results."""
        if self.args.verbose:
            print(f"\n{'='*60}")
            print(f"STAGE: {stage.name}")
            print(f"DESCRIPTION: {stage.description}")
            print(f"{'='*60}")
            
        stage.start_time = time.time()
        
        # Save command to artifacts
        cmd_file = self.artifacts_dir / f"{stage.name.lower().replace(' ', '_')}_command.txt"
        cmd_file.write_text(' '.join(cmd) + '\n')
        
        result = self._run_command(cmd)
        
        stage.end_time = time.time()
        stage.return_code = result.returncode
        stage.output = result.stdout
        stage.error_output = result.stderr
        
        # Save outputs to artifacts
        if result.stdout:
            output_file = self.artifacts_dir / f"{stage.name.lower().replace(' ', '_')}_output.txt"
            output_file.write_text(result.stdout)
            
        if result.stderr:
            error_file = self.artifacts_dir / f"{stage.name.lower().replace(' ', '_')}_errors.txt"
            error_file.write_text(result.stderr)
            
        # Print stage result
        status_symbol = "PASS" if result.returncode == 0 else "FAIL"
        print(f"{stage.name}: {status_symbol} ({stage.duration:.1f}s)")
        
        if result.returncode != 0 and not self.args.quiet:
            print(f"  Exit code: {result.returncode}")
            if result.stderr:
                print(f"  Error: {result.stderr.strip()}")
                
        return result.returncode == 0
        
    def run_formatting_check(self) -> bool:
        """Run black formatting check."""
        if self.args.no_format:
            stage = TestStage("Formatting", "Code formatting check with black")
            stage.skipped = True
            self.stages.append(stage)
            return True
            
        stage = TestStage("Formatting", "Code formatting check with black")
        cmd = ["black", "--check", "--diff", "src/", "tests/"]
        
        success = self._run_stage(stage, cmd)
        self.stages.append(stage)
        return success
        
    def run_linting(self) -> bool:
        """Run flake8 linting."""
        if self.args.no_lint:
            stage = TestStage("Linting", "Code linting with flake8")  
            stage.skipped = True
            self.stages.append(stage)
            return True
            
        stage = TestStage("Linting", "Code linting with flake8")
        cmd = ["flake8", "src/", "tests/"]
        
        success = self._run_stage(stage, cmd)
        self.stages.append(stage)
        return success
        
    def run_type_checking(self) -> bool:
        """Run mypy type checking."""
        if self.args.no_type_check:
            stage = TestStage("Type Checking", "Static type checking with mypy")
            stage.skipped = True  
            self.stages.append(stage)
            return True
            
        stage = TestStage("Type Checking", "Static type checking with mypy")
        cmd = ["mypy", "src/"]
        
        success = self._run_stage(stage, cmd)
        self.stages.append(stage)
        return success
        
    def run_unit_tests(self) -> bool:
        """Run unit tests with coverage."""
        stage = TestStage("Unit Tests", "Unit tests with coverage measurement")
        
        cmd = ["python", "-m", "pytest", "tests/unit/", "-v"]
        
        # Add coverage options if not disabled
        if not self.args.no_coverage:
            coverage_threshold = self.args.threshold
            cmd.extend([
                f"--cov=src/camera_discovery",
                f"--cov=src/camera_service", 
                f"--cov-report=term-missing",
                f"--cov-report=html:{self.artifacts_dir}/htmlcov",
                f"--cov-fail-under={coverage_threshold}"
            ])
            
        # Add unit test marker
        cmd.extend(["-m", "unit"])
        
        success = self._run_stage(stage, cmd)
        self.stages.append(stage)
        return success
        
    def run_integration_tests(self) -> bool:
        """Run integration/smoke tests."""
        if self.args.only_unit:
            stage = TestStage("Integration Tests", "Integration and smoke tests")
            stage.skipped = True
            self.stages.append(stage)
            return True
            
        stage = TestStage("Integration Tests", "Integration and smoke tests")
        
        # Look for integration tests or smoke test patterns
        cmd = ["python", "-m", "pytest", "-v"]
        
        # Try integration directory first, fall back to smoke marker
        integration_dir = self.project_root / "tests" / "integration"
        if integration_dir.exists() and any(integration_dir.glob("test_*.py")):
            cmd.append("tests/integration/")
        else:
            # Use smoke test marker if available
            cmd.extend(["-k", "smoke", "tests/"])
            
        # Add integration marker if available
        cmd.extend(["-m", "integration"])
        
        success = self._run_stage(stage, cmd)
        self.stages.append(stage)
        return success
        
    def generate_summary_report(self) -> None:
        """Generate comprehensive summary report."""
        overall_duration = time.time() - self.overall_start_time
        
        # Summary table
        print(f"\n{'='*80}")
        print("TEST EXECUTION SUMMARY")
        print(f"{'='*80}")
        print(f"{'Stage':<20} {'Status':<10} {'Duration':<10} {'Details':<30}")
        print(f"{'-'*80}")
        
        overall_success = True
        
        for stage in self.stages:
            status = stage.status
            if status == "FAILED":
                overall_success = False
                
            duration_str = f"{stage.duration:.1f}s" if stage.duration > 0 else "-"
            details = ""
            
            if status == "FAILED" and stage.error_output:
                details = stage.error_output.split('\n')[0][:30]
            elif status == "SKIPPED":
                details = "Skipped per command line option"
                
            print(f"{stage.name:<20} {status:<10} {duration_str:<10} {details:<30}")
            
        print(f"{'-'*80}")
        print(f"{'OVERALL':<20} {'PASSED' if overall_success else 'FAILED':<10} {overall_duration:.1f}s")
        print(f"{'='*80}")
        
        # Environment info
        print(f"\nENVIRONMENT:")
        print(f"  Python: {sys.version.split()[0]}")
        print(f"  Platform: {platform.system()} {platform.release()}")
        print(f"  Virtual Environment: {'Yes' if self.venv_active else 'No'}")
        print(f"  Working Directory: {self.project_root}")
        
        # Artifacts info  
        print(f"\nARTIFACTS:")
        print(f"  Directory: {self.artifacts_dir}")
        print(f"  Coverage Report: {self.artifacts_dir}/htmlcov/index.html")
        print(f"  Log Files: {len(list(self.artifacts_dir.glob('*.txt')))} files")
        
        # Next steps
        if not overall_success:
            print(f"\nNEXT STEPS:")
            failed_stages = [s for s in self.stages if s.status == "FAILED"]
            for stage in failed_stages:
                print(f"  - Fix {stage.name.lower()} issues")
                error_file = self.artifacts_dir / f"{stage.name.lower().replace(' ', '_')}_errors.txt"
                if error_file.exists():
                    print(f"    See: {error_file}")
                    
        # Save JSON report for CI integration
        self._save_json_report(overall_success, overall_duration)
        
    def _save_json_report(self, overall_success: bool, overall_duration: float) -> None:
        """Save machine-readable JSON report."""
        report = {
            "timestamp": datetime.now().isoformat(),
            "overall_success": overall_success,
            "overall_duration": overall_duration,
            "environment": {
                "python_version": sys.version.split()[0],
                "platform": f"{platform.system()} {platform.release()}",
                "virtual_env": self.venv_active,
                "working_directory": str(self.project_root)
            },
            "stages": [
                {
                    "name": stage.name,
                    "status": stage.status,
                    "duration": stage.duration,
                    "return_code": stage.return_code,
                    "skipped": stage.skipped
                }
                for stage in self.stages
            ],
            "artifacts_directory": str(self.artifacts_dir),
            "coverage_threshold": self.args.threshold
        }
        
        report_file = self.artifacts_dir / "test_report.json"
        report_file.write_text(json.dumps(report, indent=2))
        
    def run_all(self) -> int:
        """Run all test stages and return exit code."""
        if not self._setup_test_environment():
            return 2
            
        print("MediaMTX Camera Service - Test Automation")
        print(f"Starting test execution at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Artifacts will be saved to: {self.artifacts_dir}")
        
        if not self.venv_active:
            print("WARNING: Not running in virtual environment")
            
        # Run all stages
        stages_to_run = [
            self.run_formatting_check,
            self.run_linting, 
            self.run_type_checking,
            self.run_unit_tests
        ]
        
        # Add integration tests unless unit-only mode
        if not self.args.only_unit:
            stages_to_run.append(self.run_integration_tests)
            
        success = True
        for stage_func in stages_to_run:
            try:
                if not stage_func():
                    success = False
                    if self.args.fail_fast:
                        print("\nFailing fast due to stage failure")
                        break
            except Exception as e:
                print(f"ERROR: Stage failed with exception: {e}")
                success = False
                if self.args.fail_fast:
                    break
                    
        self.generate_summary_report()
        return 0 if success else 1


def parse_arguments() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Run all quality gates for MediaMTX Camera Service",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                          # Run all stages
  %(prog)s --only-unit              # Unit tests only  
  %(prog)s --no-lint --no-format    # Skip linting and formatting
  %(prog)s --threshold=85           # Custom coverage threshold
  %(prog)s --fail-fast              # Stop on first failure
        """
    )
    
    # Stage control
    parser.add_argument(
        "--no-format", action="store_true",
        help="Skip code formatting check"
    )
    parser.add_argument(
        "--no-lint", action="store_true", 
        help="Skip linting check"
    )
    parser.add_argument(
        "--no-type-check", action="store_true",
        help="Skip type checking"
    )
    parser.add_argument(
        "--no-coverage", action="store_true",
        help="Skip coverage measurement"
    )
    parser.add_argument(
        "--only-unit", action="store_true",
        help="Run unit tests only (skip integration tests)"
    )
    
    # Configuration
    parser.add_argument(
        "--threshold", type=int, default=80,
        help="Coverage threshold percentage (default: 80)"
    )
    parser.add_argument(
        "--fail-fast", action="store_true",
        help="Stop execution on first stage failure"
    )
    
    # Output control
    parser.add_argument(
        "-v", "--verbose", action="store_true",
        help="Verbose output with command details"
    )
    parser.add_argument(
        "-q", "--quiet", action="store_true", 
        help="Minimal output (errors only)"
    )
    
    return parser.parse_args()


def main() -> int:
    """Main entry point."""
    args = parse_arguments()
    
    try:
        runner = TestRunner(args)
        return runner.run_all()
    except KeyboardInterrupt:
        print("\nTest execution interrupted by user")
        return 1
    except Exception as e:
        print(f"FATAL ERROR: {e}")
        return 2


if __name__ == "__main__":
    sys.exit(main())