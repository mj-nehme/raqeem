# Battle Test Runner

This repository includes a battle test runner that exercises benchmark, stress, and load suites against the Devices and Mentor backends. Logs are saved under `logs/battle` and a brief summary is printed after each run.

## Quick Start

Run all tests quickly:

```
bash scripts/run-battle-tests.sh --quick
```

Select tests:

```
bash scripts/run-battle-tests.sh --tests benchmark,stress
```

## Options

- `--devices-url` / `--mentor-url`: Override backend URLs.
- `--tests`: Comma-separated list: `benchmark,stress,load,chaos`.
- `--benchmark` / `--stress` / `--load`: Shorthand selectors.
- `--include-chaos`: Run chaos tests (disruptive).
- `--quick`: Reduce devices/users and durations; sets `BATTLE_QUERY_LIMIT` and skips long mentor queries.
- `--scale <float>`: Scale devices/users/duration (e.g., `0.25`, `0.5`).
- `--keep-logs`: Do not delete previous logs. By default, old logs are cleaned before each run.

## Logs & Report

- Detailed per-test JSON files under `logs/battle`.
- A brief report is printed summarizing success rates per test and overall.
- To keep previous logs, pass `--keep-logs`.

## CI Integration

The CI pipeline includes a job to run the battle tests in quick mode and uploads logs as artifacts. See `.github/workflows/ci.yml`.

## Notes

- On macOS, Python runs unbuffered for live streaming if `stdbuf` is unavailable.
- Chaos tests should be run in isolated environments only.
