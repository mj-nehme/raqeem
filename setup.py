"""Setup configuration for Raqeem IoT Monitoring Platform."""
from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("VERSION", "r", encoding="utf-8") as fh:
    version = fh.read().strip()

setup(
    name="raqeem",
    version=version,
    author="Raqeem Contributors",
    description="Full-stack IoT device monitoring platform with real-time telemetry, alerts, and analytics",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/mj-nehme/raqeem",
    project_urls={
        "Bug Tracker": "https://github.com/mj-nehme/raqeem/issues",
        "Documentation": "https://github.com/mj-nehme/raqeem/tree/master/docs",
        "Source Code": "https://github.com/mj-nehme/raqeem",
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Intended Audience :: System Administrators",
        "Topic :: System :: Monitoring",
        "Topic :: Internet",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Python :: 3.13",
        "Programming Language :: Go",
        "Framework :: FastAPI",
    ],
    packages=find_packages(where="devices/backend/src") + find_packages(where="mentor"),
    python_requires=">=3.11",
    install_requires=[
        "fastapi>=0.115.0",
        "uvicorn[standard]>=0.32.0",
        "sqlalchemy[asyncio]>=2.0.0",
        "asyncpg>=0.30.0",
        "pydantic>=2.10.0",
        "pydantic-settings>=2.6.0",
        "httpx>=0.28.0",
        "boto3>=1.35.0",
        "python-jose[cryptography]>=3.3.0",
        "passlib[bcrypt]>=1.7.4",
        "python-multipart>=0.0.20",
    ],
    extras_require={
        "dev": [
            "pytest>=8.3.0",
            "pytest-asyncio>=0.25.0",
            "pytest-cov>=6.0.0",
            "respx>=0.22.0",
            "ruff>=0.9.0",
            "mypy>=1.14.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "raqeem-devices=devices.backend.src.app.main:app",
        ],
    },
    include_package_data=True,
    package_data={
        "": ["*.yaml", "*.yml", "*.md"],
    },
)
