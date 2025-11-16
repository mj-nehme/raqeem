# 🔧 Troubleshooting Guide

## Overview

This guide helps you diagnose and fix common issues with the Raqeem IoT monitoring platform. Issues are organized by category for quick reference.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Service Startup Issues](#service-startup-issues)
- [Kubernetes Issues](#kubernetes-issues)
- [Database Issues](#database-issues)
- [Networking and CORS](#networking-and-cors)
- [Frontend Issues](#frontend-issues)
- [Backend Issues](#backend-issues)
- [Performance Issues](#performance-issues)
- [Data Issues](#data-issues)
- [Docker and Image Issues](#docker-and-image-issues)
- [Development Environment](#development-environment)

## Quick Diagnostics

### Check System Status

```bash
# Check all pods
kubectl get pods -n default

# Check services
kubectl get svc -n default

# Check service discovery
./scripts/discover.sh list

# Check health of backends
curl http://localhost:30080/health
curl http://localhost:30081/health

# View recent events
kubectl get events --sort-by=.metadata.creationTimestamp -n default | tail -20
```

### Common Quick Fixes

```bash
# Restart everything
./stop.sh && ./start.sh

# Delete all data and restart fresh
./stop.sh
kubectl delete pvc --all -n default
./start.sh

# Restart a specific service
kubectl rollout restart deployment/<service-name> -n default

# Clear and reinstall frontend dependencies
cd <frontend-directory>
rm -rf node_modules package-lock.json
npm install
```

## Service Startup Issues

### Issue: "Service not ready after 60s"

**Symptoms**: Start script times out waiting for a service

**Possible Causes**:
1. Image pull failure
2. Container crash
3. Resource constraints
4. Configuration error

**Diagnostic Steps**:
```bash
# Check pod status
kubectl get pods -n default

# View pod details
kubectl describe pod <pod-name> -n default

# Check logs
kubectl logs <pod-name> -n default

# Check events
kubectl get events -n default | grep <pod-name>
```

**Solutions**:

**If image pull failed**:
```bash
# Check internet connection
ping github.com

# Manually pull image
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest

# Check if image exists
docker images | grep raqeem
```

**If container crashed**:
```bash
# View crash logs
kubectl logs <pod-name> -n default --previous

# Common fixes:
# 1. Database not ready - wait longer, then restart
kubectl rollout restart deployment/<service-name> -n default

# 2. Configuration error - check environment variables
kubectl describe pod <pod-name> -n default | grep -A 20 "Environment:"

# 3. Missing dependencies - rebuild image
```

**If resource constraints**:
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n default

# Increase Docker Desktop memory
# Docker Desktop → Settings → Resources → Memory: 4GB+ recommended
```

### Issue: Port Already in Use

**Symptoms**: Frontend won't start with "EADDRINUSE" error

**Diagnostic Steps**:
```bash
# Check what's using common ports
lsof -i :4000
lsof -i :5000
lsof -i :5001

# Or use netstat
netstat -an | grep LISTEN | grep -E '(4000|5000|5001)'
```

**Solutions**:

```bash
# Option 1: Kill the process
kill -9 <PID>

# Option 2: Use different ports
DEVICES_FRONTEND_START_PORT=6000 MENTOR_FRONTEND_START_PORT=7000 ./start.sh

# Option 3: Let auto-discovery find available ports
./stop.sh && ./start.sh
```

### Issue: Helm Release Already Exists

**Symptoms**: "Error: cannot re-use a name that is still in use"

**Solutions**:
```bash
# List all releases
helm list -n default

# Uninstall specific release
helm uninstall <release-name> -n default

# Or use stop script
./stop.sh

# Then start again
./start.sh
```

## Kubernetes Issues

### Issue: kubectl Cannot Reach Cluster

**Symptoms**: "The connection to the server localhost:8080 was refused"

**Diagnostic Steps**:
```bash
# Check cluster info
kubectl cluster-info

# Check context
kubectl config current-context

# List contexts
kubectl config get-contexts
```

**Solutions**:

**If using Docker Desktop**:
1. Open Docker Desktop
2. Go to Settings → Kubernetes
3. Click "Enable Kubernetes"
4. Wait for initialization
5. Restart Docker Desktop if needed

**If context is wrong**:
```bash
# Switch to correct context
kubectl config use-context docker-desktop

# Or for other clusters
kubectl config use-context <context-name>
```

### Issue: Pods Stuck in Pending

**Symptoms**: Pod stays in "Pending" state

**Diagnostic Steps**:
```bash
# Describe pod to see why
kubectl describe pod <pod-name> -n default
```

**Common Causes & Solutions**:

**PVC not bound**:
```bash
# Check PVC status
kubectl get pvc -n default

# If PVC is pending, check storage class
kubectl get sc

# For Docker Desktop, should have "hostpath" storage class
# If missing, restart Docker Desktop
```

**Insufficient resources**:
```bash
# Check node resources
kubectl describe nodes

# Solution: Increase Docker Desktop resources
# Docker Desktop → Settings → Resources
# CPU: 2+ cores, Memory: 4GB+
```

### Issue: Pods Stuck in ImagePullBackOff

**Symptoms**: Pod status shows "ImagePullBackOff" or "ErrImagePull"

**Diagnostic Steps**:
```bash
# Check events
kubectl describe pod <pod-name> -n default | grep -A 10 "Events:"
```

**Solutions**:

**If image doesn't exist**:
```bash
# Check image name in deployment
kubectl get deployment <deployment-name> -n default -o yaml | grep image:

# Verify image exists on Docker Hub
docker pull <image-name>

# If using local images, ensure they're built
docker images | grep raqeem
```

**If authentication required**:
```bash
# Create GitHub Container Registry secret
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token> \
  -n default

# Update deployment to use secret
# Add imagePullSecrets to pod spec
```

### Issue: CrashLoopBackOff

**Symptoms**: Pod continuously restarts

**Diagnostic Steps**:
```bash
# View current logs
kubectl logs <pod-name> -n default

# View previous crash logs
kubectl logs <pod-name> -n default --previous

# Check restart count
kubectl get pods -n default
```

**Common Causes & Solutions**:

**Database connection failure**:
```bash
# Check if postgres is running
kubectl get pods -n default | grep postgres

# Check postgres logs
kubectl logs <postgres-pod> -n default

# Verify connection string
kubectl describe pod <backend-pod> -n default | grep DATABASE_URL

# Test database connection
kubectl exec -it <backend-pod> -n default -- sh
# Inside pod:
psql $DATABASE_URL  # or telnet postgres-service 5432
```

**Application error**:
```bash
# Check logs for stack trace
kubectl logs <pod-name> -n default --tail=100

# Common fixes:
# 1. Missing environment variable
# 2. Configuration error
# 3. Database migration needed
```

## Database Issues

### Issue: Cannot Connect to PostgreSQL

**Symptoms**: Backend logs show "connection refused" or "could not connect"

**Diagnostic Steps**:
```bash
# Check if postgres pod is running
kubectl get pods -n default | grep postgres

# Check postgres service
kubectl get svc postgres-service -n default

# Check if postgres is listening
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c "SELECT 1"
```

**Solutions**:

**If postgres pod not running**:
```bash
# Check pod status
kubectl describe pod <postgres-pod> -n default

# Restart postgres
kubectl rollout restart statefulset/postgres -n default

# Or redeploy
helm uninstall postgres -n default
helm install postgres ./charts/postgres -n default
```

**If service not found**:
```bash
# Check service exists
kubectl get svc -n default

# Recreate service
kubectl delete svc postgres-service -n default
helm upgrade postgres ./charts/postgres -n default
```

**Connection string issues**:
```bash
# Correct format:
# postgresql://username:password@host:port/database

# From inside cluster:
postgresql://monitor:password@postgres-service:5432/monitoring_db

# From local machine (with port-forward):
postgresql://monitor:password@localhost:5432/monitoring_db
```

### Issue: "Too Many Connections"

**Symptoms**: Backend fails with "remaining connection slots are reserved"

**Solutions**:
```bash
# Check current connections
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Increase max_connections (edit postgres config)
kubectl edit statefulset postgres -n default
# Add environment variable:
# - name: POSTGRES_MAX_CONNECTIONS
#   value: "200"

# Or reduce connection pool size in backends
```

### Issue: Database Locked or Slow Queries

**Diagnostic Steps**:
```bash
# Check for locks
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c \
  "SELECT * FROM pg_locks WHERE NOT granted;"

# Check slow queries
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c \
  "SELECT query, state, wait_event FROM pg_stat_activity WHERE state != 'idle';"
```

**Solutions**:
```bash
# Kill blocking queries
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND pid != pg_backend_pid();"

# Add indexes for slow queries
# See ARCHITECTURE.md for recommended indexes

# Analyze query performance
EXPLAIN ANALYZE <your-query>;
```

## Networking and CORS

### Issue: CORS Errors in Browser

**Symptoms**: "Access-Control-Allow-Origin" error in browser console

**Diagnostic Steps**:
```bash
# Check CORS headers
curl -I -X OPTIONS \
  -H "Origin: http://localhost:5001" \
  -H "Access-Control-Request-Method: GET" \
  http://localhost:30081/devices
```

**Solutions**:

**For Mentor Backend (Go)**:
```bash
# Check FRONTEND_ORIGIN environment variable
kubectl describe pod <mentor-backend-pod> -n default | grep FRONTEND_ORIGIN

# Update if needed
kubectl set env deployment/mentor-backend \
  FRONTEND_ORIGIN=http://localhost:5001,http://localhost:5002 \
  -n default
```

**For Devices Backend (Python)**:
```python
# Check app/main.py CORS configuration
# Should include frontend origin in allow_origins
```

**Quick fix for development**:
```bash
# Allow all origins (NOT for production!)
kubectl set env deployment/mentor-backend \
  FRONTEND_ORIGIN=* \
  -n default
```

### Issue: Frontend Cannot Reach Backend

**Symptoms**: Network errors when frontend tries to call API

**Diagnostic Steps**:
```bash
# Check backend is accessible
curl http://localhost:30080/health
curl http://localhost:30081/health

# Check from inside cluster
kubectl run curl-test -it --rm --image=curlimages/curl -n default -- \
  curl http://devices-backend:8080/health
```

**Solutions**:

**If NodePort not accessible**:
```bash
# Check service type
kubectl get svc devices-backend -n default -o yaml | grep type:

# Verify NodePort is configured
kubectl get svc devices-backend -n default -o yaml | grep nodePort:

# Test NodePort
curl http://localhost:30080/health
```

**If frontend using wrong URL**:
```bash
# Check frontend environment variables
cd mentor/frontend
cat .env* 

# Should be:
# VITE_API_URL=http://localhost:30081

# Restart frontend with correct URL
VITE_API_URL=http://localhost:30081 npm run dev
```

### Issue: Service Discovery Not Working

**Symptoms**: Services can't find each other

**Diagnostic Steps**:
```bash
# List registered services
./scripts/discover.sh list

# Check registry files
ls -la .deploy/registry/

# Test DNS resolution
kubectl run dns-test -it --rm --image=busybox -n default -- \
  nslookup devices-backend.default.svc.cluster.local
```

**Solutions**:
```bash
# Regenerate registry
rm -rf .deploy/registry/
./start.sh

# Verify Kubernetes DNS
kubectl get svc -n kube-system | grep kube-dns

# Restart CoreDNS if needed
kubectl rollout restart deployment/coredns -n kube-system
```

## Frontend Issues

### Issue: Blank Page or White Screen

**Diagnostic Steps**:
```bash
# Check browser console for errors
# Open browser DevTools (F12) → Console tab

# Check if frontend is running
lsof -i :5001  # or your frontend port

# Check frontend logs in terminal where npm run dev is running
```

**Solutions**:

**If build errors**:
```bash
cd mentor/frontend  # or devices/frontend

# Clear and reinstall
rm -rf node_modules package-lock.json dist
npm install

# Restart dev server
npm run dev
```

**If API connection issue**:
```bash
# Check API URL in browser console
# Should see network requests to correct backend

# Verify backend URL
echo $VITE_API_URL

# Set if needed
export VITE_API_URL=http://localhost:30081
npm run dev
```

### Issue: Module Not Found Errors

**Symptoms**: "Cannot find module" or "Module not found: Error: Can't resolve"

**Solutions**:
```bash
cd <frontend-directory>

# Clear node_modules
rm -rf node_modules package-lock.json

# Clear npm cache
npm cache clean --force

# Reinstall
npm install

# If specific module missing
npm install <module-name>
```

### Issue: Hot Module Replacement Not Working

**Symptoms**: Changes don't appear without manual refresh

**Solutions**:
```bash
# Restart dev server
# Ctrl+C then npm run dev

# Check file watchers limit (Linux)
echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Use polling mode
npm run dev -- --force
```

## Backend Issues

### Issue: Python Import Errors

**Symptoms**: "ModuleNotFoundError" or "ImportError"

**Solutions**:
```bash
cd devices/backend/src

# Install dependencies
pip install -r ../requirements.txt

# Or use virtual environment
python -m venv venv
source venv/bin/activate  # or venv\Scripts\activate on Windows
pip install -r ../requirements.txt

# Run with correct PYTHONPATH
export PYTHONPATH=/home/runner/work/raqeem/raqeem/devices/backend/src:$PYTHONPATH
python -m pytest
```

### Issue: Go Build Failures

**Symptoms**: "undefined: <symbol>" or "cannot find package"

**Solutions**:
```bash
cd mentor/backend/src

# Download dependencies
go mod download

# Update go.mod
go mod tidy

# Clear cache and rebuild
go clean -cache
go build

# Verify go version
go version  # Should be 1.25+
```

### Issue: FastAPI Server Won't Start

**Symptoms**: uvicorn crashes or port already in use

**Solutions**:
```bash
# Check if port is in use
lsof -i :8081

# Kill process if needed
kill -9 <PID>

# Start on different port
uvicorn app.main:app --reload --port 8082

# Check for syntax errors
python -m py_compile app/main.py
```

### Issue: Gin Server Panic

**Symptoms**: Go backend crashes with panic

**Diagnostic Steps**:
```bash
# View stack trace in logs
kubectl logs <mentor-backend-pod> -n default --tail=100

# Common causes:
# 1. Nil pointer dereference
# 2. Database connection nil
# 3. Uncaught error
```

**Solutions**:
```bash
# Add nil checks
# Add error handling
# Check database connection before use

# Restart with more logging
kubectl set env deployment/mentor-backend GIN_MODE=debug -n default
```

## Performance Issues

### Issue: Slow API Responses

**Diagnostic Steps**:
```bash
# Check response time
time curl http://localhost:30080/api/v1/devices

# Check database query performance
# Enable slow query log in PostgreSQL

# Check resource usage
kubectl top pods -n default
```

**Solutions**:

**If database queries slow**:
```sql
-- Add indexes (see ARCHITECTURE.md)
CREATE INDEX idx_device_metrics_device_timestamp 
  ON device_metrics(device_id, timestamp DESC);

-- Analyze query plans
EXPLAIN ANALYZE SELECT * FROM device_metrics WHERE device_id = 'dev1';
```

**If high CPU/memory**:
```bash
# Scale up replicas
kubectl scale deployment devices-backend --replicas=3 -n default

# Increase resource limits
kubectl edit deployment devices-backend -n default
# Update resources.limits
```

**If connection pool exhausted**:
```python
# Increase pool size (Python)
# In database connection setup:
engine = create_engine(url, pool_size=20, max_overflow=10)
```

```go
// Increase pool size (Go)
db.DB().SetMaxOpenConns(100)
db.DB().SetMaxIdleConns(10)
```

### Issue: High Memory Usage

**Diagnostic Steps**:
```bash
# Check memory usage
kubectl top pods -n default

# Get detailed metrics
kubectl exec -it <pod-name> -n default -- top
```

**Solutions**:
```bash
# Check for memory leaks
# Use profiling tools:
# - Python: memory_profiler
# - Go: pprof

# Restart pods to free memory
kubectl rollout restart deployment/<deployment-name> -n default

# Increase memory limits
kubectl set resources deployment/<deployment-name> \
  --limits=memory=1Gi \
  -n default
```

## Data Issues

### Issue: No Devices Showing Up

**Diagnostic Steps**:
```bash
# Check if devices in database
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c "SELECT * FROM devices;"

# Check backend logs for errors
kubectl logs <devices-backend-pod> -n default | grep error
```

**Solutions**:
```bash
# Register a device manually
curl -X POST http://localhost:30080/api/v1/devices/register \
  -H "Content-Type: application/json" \
  -d '{"id": "test-001", "name": "Test Device", "device_type": "laptop"}'

# Check device simulator is running
# Open http://localhost:<auto-detected-port>
# Click "Register Device"
```

### Issue: Metrics Not Updating

**Diagnostic Steps**:
```bash
# Check if metrics being submitted
kubectl logs <devices-backend-pod> -n default | grep metrics

# Check database
kubectl exec -it <postgres-pod> -n default -- \
  psql -U monitor -d monitoring_db -c \
  "SELECT COUNT(*) FROM device_metrics;"
```

**Solutions**:
```bash
# Submit test metrics
curl -X POST http://localhost:30080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "deviceid": "test-001",
    "cpu_usage": 45.5,
    "memory_used": 8589934592,
    "disk_used": 107374182400
  }'

# Restart device simulator
# Use auto-simulation mode
```

### Issue: Alerts Not Forwarding

**Symptoms**: Alerts appear in devices backend but not in mentor backend

**Diagnostic Steps**:
```bash
# Check devices backend logs
kubectl logs <devices-backend-pod> -n default | grep "Forwarding alert"

# Check mentor backend logs
kubectl logs <mentor-backend-pod> -n default | grep alert

# Verify MENTOR_API_URL
kubectl describe pod <devices-backend-pod> -n default | grep MENTOR_API_URL
```

**Solutions**:
```bash
# Set correct MENTOR_API_URL
kubectl set env deployment/devices-backend \
  MENTOR_API_URL=http://mentor-backend:8080 \
  -n default

# Test connection from devices backend
kubectl exec -it <devices-backend-pod> -n default -- \
  curl http://mentor-backend:8080/health
```

## Docker and Image Issues

### Issue: Image Build Fails

**Diagnostic Steps**:
```bash
# Try building manually
cd devices/backend
docker build -t test-image .

# Check Docker daemon
docker ps

# Check disk space
df -h
```

**Solutions**:
```bash
# Clear Docker cache
docker system prune -a

# Rebuild with no cache
docker build --no-cache -t test-image .

# Check Dockerfile syntax
docker build --check -t test-image .
```

### Issue: Cannot Pull Images

**Symptoms**: "manifest unknown" or "repository does not exist"

**Solutions**:
```bash
# Check image name
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest

# Login to GitHub Container Registry if private
echo $GITHUB_TOKEN | docker login ghcr.io -u <github-username> --password-stdin

# Check internet connection
ping github.com

# Ensure image is public or you have access
# Visit: https://github.com/mj-nehme/raqeem/pkgs/container/raqeem%2Fdevices-backend
```

## Development Environment

### Issue: Environment Variables Not Loading

**Symptoms**: Application can't find configuration

**Solutions**:
```bash
# Check environment variables
echo $DATABASE_URL

# Set manually
export DATABASE_URL="postgresql://monitor:password@localhost:5432/monitoring_db"

# Load from file
source .env

# For Python: use python-dotenv
# For Go: use godotenv
```

### Issue: Tests Failing

**Diagnostic Steps**:
```bash
# Run tests with verbose output
pytest -v  # Python
go test -v ./...  # Go
npm test  # JavaScript

# Check test database connection
# Run specific test
pytest tests/api/test_alerts_forwarding.py::test_post_alerts_is_saved_and_forwarded -v
```

**Solutions**:
```bash
# Ensure test database is running
docker run -d --name test-db \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:16

# Update test configuration
# Check conftest.py or test setup

# Clear test cache
pytest --cache-clear
go clean -testcache
```

## Getting More Help

### Enable Debug Logging

**Python (FastAPI)**:
```bash
# Set log level
export LOG_LEVEL=DEBUG
uvicorn app.main:app --reload --log-level debug
```

**Go (Gin)**:
```bash
# Set Gin mode
export GIN_MODE=debug
go run main.go
```

**Kubernetes**:
```bash
# Increase log verbosity
kubectl logs <pod-name> -n default --tail=1000 -f
```

### Collect Diagnostic Information

```bash
# Create support bundle
cat > support-bundle.txt <<EOF
=== Pod Status ===
$(kubectl get pods -n default)

=== Service Status ===
$(kubectl get svc -n default)

=== Recent Events ===
$(kubectl get events -n default --sort-by=.metadata.creationTimestamp | tail -20)

=== Logs ===
$(kubectl logs <pod-name> -n default --tail=100)
EOF
```

### Resources

- [First Time Setup](FIRST_TIME_SETUP.md)
- [Development Guide](DEVELOPMENT.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Testing Guide](TESTING.md)
- [GitHub Issues](https://github.com/mj-nehme/raqeem/issues)

## Still Stuck?

1. Check existing [GitHub Issues](https://github.com/mj-nehme/raqeem/issues)
2. Create a new issue with:
   - Problem description
   - Steps to reproduce
   - Expected vs actual behavior
   - Logs and diagnostic output
   - Environment details (OS, Docker version, Kubernetes version)
3. Include output from diagnostic commands above
