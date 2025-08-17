#!/usr/bin/env python3
"""
MediaMTX Camera Service Python SDK

A Python SDK for interacting with the MediaMTX Camera Service via WebSocket JSON-RPC.
Provides high-level client interface with JWT and API key authentication support.
"""

from setuptools import setup, find_packages
import os

# Read the README file
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), 'README.md')
    if os.path.exists(readme_path):
        with open(readme_path, 'r', encoding='utf-8') as f:
            return f.read()
    return "MediaMTX Camera Service Python SDK"

setup(
    name="mediamtx-camera-sdk",
    version="1.0.0",
    description="Python SDK for MediaMTX Camera Service",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    author="MediaMTX Camera Service Team",
    author_email="team@mediamtx-camera-service.com",
    url="https://github.com/mediamtx/camera-service",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Multimedia :: Video :: Capture",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
    python_requires=">=3.8",
    install_requires=[
        "websockets>=10.0",
        "asyncio",
        "typing-extensions>=4.0.0",
    ],
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-asyncio>=0.18.0",
            "black>=22.0",
            "flake8>=4.0",
            "mypy>=0.950",
        ],
    },
    entry_points={
        "console_scripts": [
            "mediamtx-camera-cli=mediamtx_camera_sdk.cli:cli_main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
)
