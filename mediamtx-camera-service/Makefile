.PHONY: help install dev-install test lint format clean run

help:
@echo "Available commands:"
@echo "  install     - Install production dependencies"
@echo "  dev-install - Install development dependencies" 
@echo "  test        - Run test suite"
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
pytest

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
python -m camera_service.main
