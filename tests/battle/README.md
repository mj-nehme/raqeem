# Battle Testing Suite for v0.2.0

This directory contains comprehensive end-to-end battle testing for the Raqeem IoT platform, designed to ensure production-readiness through stress, load, and chaos testing.

## Overview

The battle testing suite validates that the Raqeem platform can handle:
- **1000+ concurrent device connections** with sustained telemetry
- **High-volume data ingestion** (metrics, activities, alerts, screenshots)
- **Network failures** and service disruptions
- **Database connection exhaustion** and recovery
- **Storage failures** (MinIO/S3)
- **Concurrent API access** from multiple clients
- **Large payload handling** (screenshots, bulk operations)

## Test Categories

### 1. Stress Tests (`stress_test.py`)
High-volume testing to find breaking points and validate performance under load:
- 1000+ device registration (burst and sustained)
- Continuous telemetry ingestion (multiple metrics per second per device)
- Concurrent alert generation and forwarding
- Bulk screenshot uploads
- Database query performance under load

### 2. Chaos Engineering Tests (`chaos_test.py`)
Failure scenario validation to ensure graceful degradation:
- Network partition simulation
- Service crash and recovery
- Database connection failures
- MinIO/S3 unavailability
- Partial service degradation
- Message queue overflow

### 3. Load Tests (`load_test.py`)
Sustained load testing to validate normal and peak operations:
- Continuous device operation simulation
- Frontend API load (concurrent dashboard access)
- Alert forwarding pipeline under load
- Query performance degradation testing
- Resource utilization monitoring

### 4. Performance Benchmarks (`benchmark_test.py`)
Baseline performance metrics and regression detection:
- Device registration latency (p50, p95, p99)
- Telemetry ingestion throughput
- Alert forwarding latency
- Database query response times
- Screenshot upload/download times
- API endpoint response times

### 5. Data Consistency Tests (`consistency_test.py`)
Validate data integrity under stress:
- Concurrent writes to same device
- Alert forwarding reliability
- Database transaction isolation
- Eventually consistent operations
- Data loss detection

### 6. Recovery Tests (`recovery_test.py`)
Backup, restore, and migration scenarios:
- Database backup under load
- Service restart with in-flight requests
- Data migration scenarios
- Connection pool exhaustion and recovery
- Graceful shutdown validation

## Prerequisites

### Required Services
All battle tests require a running Docker environment with:
```bash
docker-compose -f .github/docker-compose.test.yml up -d
```

Or use the full Kubernetes stack:
```bash
./start.sh
```

### Required Python Packages
```bash
pip install -r requirements.txt
```

Includes:
- `requests` - HTTP client for API testing
- `pytest` - Test framework
- `psutil` - Resource monitoring (optional)

## Running Tests

### Quick Start (All Tests)
```bash
./run_battle_tests.sh
```

This runs all test categories and generates a comprehensive report.

### Individual Test Suites

**Stress Test (30-60 minutes)**
```bash
python3 tests/battle/stress_test.py --devices 1000 --duration 300
```

**Chaos Test (15-30 minutes)**
```bash
python3 tests/battle/chaos_test.py --scenarios all
```

**Load Test (5-10 minutes)**
```bash
python3 tests/battle/load_test.py --concurrent-users 100 --duration 300
```

**Benchmark Test (5 minutes)**
```bash
python3 tests/battle/benchmark_test.py --samples 1000
```

**Consistency Test (10 minutes)**
```bash
python3 tests/battle/consistency_test.py --iterations 100
```

**Recovery Test (15 minutes)**
```bash
python3 tests/battle/recovery_test.py --scenarios all
```

## Test Configuration

Configuration is done via command-line arguments or environment variables:

### Environment Variables
```bash
export DEVICES_BACKEND_URL=http://localhost:8081
export MENTOR_BACKEND_URL=http://localhost:8080
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export MINIO_ENDPOINT=localhost:9000
```

### Command-Line Options
```bash
--devices N           # Number of devices to simulate (default: 100)
--duration N          # Test duration in seconds (default: 60)
--concurrent-users N  # Concurrent API users (default: 10)
--samples N           # Number of samples for benchmarks (default: 100)
--scenarios NAME      # Specific scenario or 'all' (default: all)
--verbose            # Enable detailed logging
--report-file PATH   # Save results to file (default: stdout)
```

## Expected Results

### Acceptance Criteria (from Issue)
- ✅ Platform handles 1000+ concurrent device connections
- ✅ No data loss during failure scenarios
- ✅ All services recover gracefully from outages
- ✅ Frontend remains responsive under load
- ✅ Complete test automation in CI/CD pipeline

### Performance Targets
| Metric | Target | Acceptable |
|--------|--------|------------|
| Device registration | <100ms p95 | <200ms p95 |
| Telemetry ingestion | >1000 msg/sec | >500 msg/sec |
| Alert forwarding | <500ms p95 | <1s p95 |
| Database queries | <50ms p95 | <100ms p95 |
| Screenshot upload | <2s for 1MB | <5s for 1MB |
| API response time | <200ms p95 | <500ms p95 |

### Reliability Targets
- **Availability**: 99.9% uptime during normal operations
- **Recovery Time**: <30s for service restart
- **Data Loss**: 0% under normal failures
- **Graceful Degradation**: Continue operations with reduced capacity

## CI/CD Integration

Battle tests run in CI on:
- Pre-release branches (manual trigger)
- Scheduled nightly runs
- Release candidate validation

See `.github/workflows/battle-test.yml` for configuration.

## Troubleshooting

### "Too many database connections"
```bash
# Increase PostgreSQL connection limit
docker exec -it <postgres-container> psql -U monitor -d monitoring_db -c \
  "ALTER SYSTEM SET max_connections = 500;"
docker restart <postgres-container>
```

### "Services become unresponsive"
- Reduce `--devices` or `--concurrent-users`
- Increase `--duration` to spread load over time
- Check resource limits (Docker/Kubernetes)

## See Also

- [../integration/README.md](../integration/README.md) - Integration tests
- [../../docs/TESTING.md](../../docs/TESTING.md) - Testing guide
- [../../docs/DEPLOYMENT.md](../../docs/DEPLOYMENT.md) - Deployment guide
