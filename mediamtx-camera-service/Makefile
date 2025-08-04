.PHONY: help install dev-install test test-unit test-integration test-coverage test-quick lint format clean run

help:
	@echo "Available commands:"
	@echo "  install     - Install production dependencies"
	@echo "  dev-install - Install development dependencies" 
	@echo "  test        - Run full test suite"
	@echo "  test-unit   - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-quick  - Run quick smoke tests"
	@echo "  lint        - Run linting checks"
	@echo "  format      - Format code with black"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run development server"

install:
	pip install -r requirements.txt

dev-install:
	pip install -r requirements-dev.txt
	pip install -e .

test:
	python3 run_tests.py --coverage

test-unit:
	python3 run_tests.py --unit

test-integration:
	python3 run_tests.py --integration

test-coverage:
	python3 run_tests.py --coverage --verbose

test-quick:
	python3 run_tests.py --specific "test_placeholder or test_frame_rate_extraction"

lint:
	flake8 src/ tests/
	mypy src/

format:
	black src/ tests/

clean:
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete
	rm -rf build/ dist/ *.egg-info/ .coverage htmlcov/

run:
	python3 -m camera_service.main

# New test setup targets
test-setup:
	python3 run_tests.py --create-files
	pip install pytest pytest-asyncio pytest-cov

test-hybrid-monitor:
	python3 -m pytest tests/test_hybrid_monitor_comprehensive.py -v

# Continuous testing during development
test-watch:
	python3 -m pytest tests/ --tb=short -q --disable-warnings --looponfail