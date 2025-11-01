# Running CI Locally

This guide shows you how to run the GitHub Actions CI pipeline on your local machine using [act](https://github.com/nektos/act).

## Prerequisites

### Install act

**macOS (with Homebrew):**
```bash
brew install act
```

**Linux:**
```bash
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

**Windows (with Chocolatey):**
```bash
choco install act-cli
```

### Install Docker

Act requires Docker to run the CI containers. Make sure Docker Desktop is installed and running.

## Running the CI Pipeline

### Run the full CI workflow

```bash
act -j build-and-test
```

This will:
- Start a Postgres container
- Run Python tests (devices backend)
- Run Go tests (mentor backend)
- Run frontend tests (devices and mentor)

### Run specific steps

**Run only Python tests:**
```bash
act -j build-and-test -s "Run Python tests (devices backend)"
```

**Run only Go tests:**
```bash
act -j build-and-test -s "Go test (mentor backend)"
```

**Run only frontend tests:**
```bash
act -j build-and-test -s "Install and test mentor frontend"
```

### Dry run (see what would execute)

```bash
act -j build-and-test --dryrun
```

### Use a specific event

```bash
# Simulate a push to main
act push -j build-and-test

# Simulate a pull request
act pull_request -j build-and-test
```

## Troubleshooting

### act is slow or uses too much disk space

Act downloads large Docker images. Use a medium-sized runner:

```bash
act -j build-and-test --container-architecture linux/amd64
```

### Services not starting properly

Act's service containers may behave differently than GitHub Actions. If you encounter issues with the Postgres service:

1. Start Postgres manually:
   ```bash
   docker run -d --name act-postgres \
     -e POSTGRES_USER=monitor \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=monitoring_db \
     -p 5432:5432 \
     postgres:16
   ```

2. Run act without services:
   ```bash
   act -j build-and-test --bind
   ```

### Secrets and environment variables

If you need to pass secrets or env vars:

```bash
act -j build-and-test -s GITHUB_TOKEN=your_token_here
```

Or create a `.secrets` file:
```
GITHUB_TOKEN=your_token_here
OTHER_SECRET=value
```

Then run:
```bash
act -j build-and-test --secret-file .secrets
```

## Alternative: Run Integration Tests

For a more reliable local test that starts the full stack, use our integration test suite:

```bash
./tests/integration/run_integration_tests.sh
```

This uses docker-compose to start all services and runs an end-to-end test of the alert flow.

## Quick Comparison

| Method | Pros | Cons |
|--------|------|------|
| **act** | - Matches GitHub Actions exactly<br>- Tests CI workflow itself<br>- No code changes needed | - Slower (downloads images)<br>- Some service quirks<br>- Requires Docker |
| **Integration Tests** | - Fast and reliable<br>- Tests real service interactions<br>- Easy to customize | - Doesn't test CI workflow<br>- Requires docker-compose |
| **Unit Tests** | - Very fast<br>- No external dependencies<br>- Good for TDD | - Limited coverage<br>- Mocked dependencies |

## Best Practice

Use all three approaches:

1. **During development**: Run unit tests frequently
   ```bash
   cd devices/backend/src && pytest tests/api/test_alerts_forwarding.py -v
   ```

2. **Before pushing**: Run integration tests
   ```bash
   ./tests/integration/run_integration_tests.sh
   ```

3. **To debug CI failures**: Run act
   ```bash
   act -j build-and-test
   ```

4. **On every push/PR**: GitHub Actions runs automatically
