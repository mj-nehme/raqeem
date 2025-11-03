# 🚀 First Time Setup Guide

Welcome to Raqeem! This guide will help you get started from scratch, even if you're new to the project or returning after a long time.

## Prerequisites

Before you begin, ensure you have these tools installed:

### Required Tools
- **Docker Desktop** (with Kubernetes enabled) or any Kubernetes cluster
  - [Download Docker Desktop](https://www.docker.com/products/docker-desktop)
  - Enable Kubernetes in Docker Desktop settings
- **kubectl** - Kubernetes command-line tool
  ```bash
  # macOS
  brew install kubectl
  
  # Verify installation
  kubectl version --client
  ```
- **Helm 3.x** - Kubernetes package manager
  ```bash
  # macOS
  brew install helm
  
  # Verify installation
  helm version
  ```
- **Node.js & npm** - For running frontend applications
  ```bash
  # macOS
  brew install node
  
  # Verify installation
  node --version
  npm --version
  ```

### Optional Tools (for development)
- **Go 1.23+** - For mentor backend development
- **Python 3.11+** - For devices backend development

## 📋 Step-by-Step Setup

### 1️⃣ Clone the Repository
```bash
git clone <repository-url>
cd raqeem
```

### 2️⃣ Verify Kubernetes Cluster
Make sure your Kubernetes cluster is running:

```bash
kubectl cluster-info
```

You should see output showing your cluster is running. If not:
- **Docker Desktop**: Open Docker Desktop → Settings → Kubernetes → Enable Kubernetes
- **Other**: Start your Kubernetes cluster (minikube, kind, k3d, etc.)

### 3️⃣ Start Everything!
**No configuration needed!** Just run:

```bash
./start.sh
```

The smart discovery system uses intelligent defaults:
- Frontend ports: Auto-detected starting from 4000, 5000
- Backend ports: Stable Kubernetes NodePort (30080, 30081)  
- Namespace: `default`
- All other settings: Auto-configured

The script will automatically:
1. ✅ Validate your environment
2. ✅ Deploy PostgreSQL and wait for it to be ready
3. ✅ Deploy MinIO (S3-compatible storage) and wait for it
4. ✅ Deploy backends with stable NodePort services (no port-forwarding)
5. ✅ Auto-detect available frontend ports
6. ✅ Install npm dependencies and start frontends
7. ✅ Register all services in discovery registry

**First-time run will take 3-5 minutes** as it:
- Pulls Docker images from Docker Hub
- Installs npm dependencies
- Initializes databases

**Subsequent runs are much faster** (30-60 seconds)!

### 4️⃣ Access Your Applications

Once the start script completes, you'll see:

```
🎉 Smart Service Discovery Ready!

📱 Discovered Services:
  - Devices Backend:   http://localhost:30080/docs
  - Mentor Backend:    http://localhost:30081/health
  - Mentor Dashboard:  http://localhost:5001
  - Device Simulator:  http://localhost:4000
```

Check actual discovered services:
```bash
./scripts/discover.sh list
```

Open these URLs in your browser:
- **Mentor Dashboard** (`http://localhost:15000`) - Main monitoring interface
- **Device Simulator** (`http://localhost:14000`) - Simulate device data
- **API Documentation** - Interactive API docs at the `/docs` endpoints

## 🛑 Stopping Everything

When you're done, stop all services:

```bash
./stop.sh
```

This will:
- Stop frontend dev servers
- Stop port-forwards
- Uninstall Helm releases
- **Preserve your data** (PostgreSQL data persists in Kubernetes volumes)

### Delete All Data (Fresh Start)
If you want to completely reset and delete all data:

```bash
./stop.sh
kubectl delete pvc --all -n default
```

## 🔧 Troubleshooting

### "Port already in use"
If you see port conflicts:

```bash
# Check what's using your ports
lsof -i :14100  # or whichever port
lsof -i :15100

# Kill the process using the port
kill -9 <PID>

# Or change the port in .env
```

### "kubectl cannot reach cluster"
```bash
# Check cluster status
kubectl cluster-info

# If using Docker Desktop, restart it
# Settings → Kubernetes → Restart

# Verify nodes are ready
kubectl get nodes
```

### "Helm chart failed to install"
```bash
# Check pod status
kubectl get pods -n default

# Check pod logs for errors
kubectl logs <pod-name> -n default

# Describe pod to see events
kubectl describe pod <pod-name> -n default
```

### "Frontend won't start"
```bash
# Clear npm cache and reinstall
cd mentor/frontend  # or devices/frontend
rm -rf node_modules package-lock.json
npm install
```

### "Service not ready after 60s"
This usually means:
1. Docker image failed to pull → Check internet connection
2. Container crashed → Check pod logs: `kubectl logs <pod-name>`
3. Resource constraints → Check Docker Desktop has enough memory (4GB+ recommended)

### Complete Reset
If something is really broken, do a complete reset:

```bash
# Stop everything
./stop.sh

# Delete all Kubernetes resources in default namespace
kubectl delete all --all -n default
kubectl delete pvc --all -n default

# Start fresh
./start.sh
```

## 📊 What's Running?

After starting, you have:

### Kubernetes Pods (in `default` namespace)
- **postgres** - PostgreSQL database
- **minio** - S3-compatible object storage
- **devices-backend** - FastAPI service (Python)
- **mentor-backend** - Gin service (Go)

### Local Processes
- **Port-forward to devices-backend** - Tunnels K8s service to localhost
- **Port-forward to mentor-backend** - Tunnels K8s service to localhost  
- **Devices Frontend** - Vite dev server
- **Mentor Frontend** - Vite dev server

### Service Discovery (Already Configured!)
Services in Kubernetes talk to each other using DNS names:
- `postgres-service.default.svc.cluster.local`
- `minio-service.default.svc.cluster.local`
- `devices-backend.default.svc.cluster.local`
- `mentor-backend.default.svc.cluster.local`

You don't need to configure this - it's built into Kubernetes! 🎉

## 📚 Next Steps

Now that everything is running:

1. **Explore the Mentor Dashboard** - View device monitoring data
2. **Use the Device Simulator** - Generate test data
3. **Check API Docs** - Interactive docs at `/docs` endpoints
4. **Read the main README** - Learn about development workflow
5. **Check the Architecture section** - Understand the system design

## 🆘 Need Help?

- Check the main [README.md](../README.md) for detailed documentation
- Look at [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for common issues
- Review logs: `kubectl logs <pod-name> -n default`
- Check pod status: `kubectl get pods -n default`

## 🎓 Understanding the Stack

### Why Kubernetes?
- **Consistency**: Same setup works on any machine (Mac, Linux, Windows)
- **Service Discovery**: Services find each other automatically
- **Scalability**: Easy to scale services up/down
- **Production-like**: Local environment mirrors production

### Why Helm?
- **Templating**: Reusable Kubernetes configs
- **Versioning**: Track changes to infrastructure
- **Easy Updates**: Single command to update services

### Why Port-Forwarding?
- **Security**: Services stay inside K8s cluster
- **Simplicity**: No need for Ingress controllers locally
- **Flexibility**: Each developer can use their own ports

---

**You're all set!** 🎉 Run `./start.sh` and start building!
