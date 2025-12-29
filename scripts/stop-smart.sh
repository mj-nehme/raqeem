#!/usr/bin/env bash
set -e

echo "🛑 Stopping Smart Service Discovery Environment..."
echo ""

NAMESPACE=${NAMESPACE:-default}
CLEAN_DATA=${1:-""}

# Stop frontend processes
  if [[ -f ".deploy/smart.pids" ]]; then
    echo "🌐 Stopping frontend processes..."
  source .deploy/smart.pids
  
    if [[ -n "$MENTOR_FE_PID" ]] && kill -0 "$MENTOR_FE_PID" 2>/dev/null; then
      echo "  - Stopping Dashboard (PID $MENTOR_FE_PID)"
    kill "$MENTOR_FE_PID" 2>/dev/null || true
  fi
  
  if [[ -n "$DEVICES_FE_PID" ]] && kill -0 "$DEVICES_FE_PID" 2>/dev/null; then
    echo "  - Stopping Devices Frontend (PID $DEVICES_FE_PID)"
    kill "$DEVICES_FE_PID" 2>/dev/null || true
  fi
  
  rm -f .deploy/smart.pids
fi

# Clean up service registry
if [[ -d ".deploy/registry" ]]; then
  echo "🗂️  Cleaning up service registry..."
  rm -rf .deploy/registry
fi

# Uninstall Helm releases
echo "📦 Uninstalling Kubernetes services..."
helm uninstall devices-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ devices-backend not found"
helm uninstall mentor-backend -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ mentor-backend not found"
helm uninstall minio -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ minio not found"
helm uninstall postgres -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ postgres not found"

# Optionally stop local PostgreSQL if it's occupying port 5432
stop_local_postgres() {
  echo ""
  echo "🧪 Checking for local PostgreSQL on port 5432..."

  # Detect listeners on 5432 (macOS/Linux)
  local LISTENERS
  if command -v lsof >/dev/null 2>&1; then
    LISTENERS=$(lsof -nP -iTCP:5432 -sTCP:LISTEN 2>/dev/null || true)
  else
    LISTENERS=""
  fi

  if [[ -z "$LISTENERS" ]]; then
    echo "  ℹ️ No local process is listening on 5432."
    return 0
  fi

  echo "  ⚠️  Detected the following processes on 5432:"
  echo "$LISTENERS" | sed 's/^/    /'

  # Prompt for confirmation
  read -r -p "❓ Stop PostgreSQL/listeners on 5432 now? [y/N] " REPLY
  case "$REPLY" in
    [yY][eE][sS]|[yY]) ;;
    *)
      echo "  ⏭️  Skipping PostgreSQL shutdown."
      return 0
      ;;
  esac

  echo ""
  echo "🛑 Attempting to stop PostgreSQL (best-effort)..."

  # 1) Stop Docker containers exposing 5432 or based on postgres image
  if command -v docker >/dev/null 2>&1; then
    # Find containers exposing host port 5432 or running postgres image (portable for bash 3.x)
    DOCKER_CIDS=$(docker ps --format '{{.ID}} {{.Image}} {{.Ports}}' 2>/dev/null | awk '/:5432->/ || tolower($2) ~ /postgres/ {print $1}')
    if [[ -n "$DOCKER_CIDS" ]]; then
      echo "  🐳 Stopping Docker containers using 5432: $DOCKER_CIDS"
      for cid in $DOCKER_CIDS; do
        docker stop "$cid" >/dev/null 2>&1 || true
      done
    fi
  fi

  # 2) Stop Homebrew services (common versions)
  if [[ "$(uname -s)" == "Darwin" ]] && command -v brew >/dev/null 2>&1; then
    for svc in postgresql postgresql@16 postgresql@15 postgresql@14; do
      if brew services list 2>/dev/null | awk '{print $1" "$2}' | grep -q "^${svc} started"; then
        echo "  🍺 Stopping brew service: ${svc}"
        brew services stop "${svc}" >/dev/null 2>&1 || true
      fi
    done
  fi

  # Re-check; if still listening, offer to kill remaining PIDs
  local REMAIN
  REMAIN=$(lsof -nP -iTCP:5432 -sTCP:LISTEN 2>/dev/null || true)
  if [[ -n "$REMAIN" ]]; then
    echo "  ⚠️  5432 still in use after Docker/brew stop attempts:"
    echo "$REMAIN" | sed 's/^/    /'
    read -r -p "❓ Force-kill remaining 5432 listener PIDs? [y/N] " REPLY2
    case "$REPLY2" in
      [yY][eE][sS]|[yY])
        # Extract PIDs from lsof output (2nd column)
        PIDS=$(echo "$REMAIN" | awk 'NR>1 {print $2}' | sort -u)
        if [[ -n "$PIDS" ]]; then
          echo "  🗡️  Killing PIDs: $PIDS"
          for p in $PIDS; do
            kill "$p" 2>/dev/null || true
          done
          sleep 1
          # SIGKILL if still present
          for p in $PIDS; do
            if kill -0 "$p" 2>/dev/null; then
              kill -9 "$p" 2>/dev/null || true
            fi
          done
        fi
        ;;
      *) echo "  ⏭️  Skipping force kill." ;;
    esac
  fi

  # Final status
  if lsof -nP -iTCP:5432 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "  ❌ Port 5432 still in use. Consider manual investigation."
  else
    echo "  ✅ Port 5432 is now free."
  fi
}

# Handle data cleanup based on flag
if [[ "$CLEAN_DATA" == "--clean" ]] || [[ "$CLEAN_DATA" == "-c" ]]; then
  echo ""
  echo "🗑️  Deleting persistent volumes (fresh database on next start)..."
  kubectl delete pvc --all -n "$NAMESPACE" 2>/dev/null || echo "  ℹ️ No PVCs to delete"

  # Wait for PVCs to be fully removed to avoid Helm ownership errors on next start
  echo "⏳ Waiting for PVCs to be fully removed..."
  for i in $(seq 1 30); do
    REMAINING=$(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$REMAINING" == "0" ]]; then
      break
    fi
    sleep 2
  done

  # If postgres-pvc is still present (Terminating/finalizer), attempt force removal
  if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
    echo "⚠️  postgres-pvc still present; attempting to remove finalizers and force delete..."
    kubectl patch pvc postgres-pvc -n "$NAMESPACE" -p '{"metadata":{"finalizers":null}}' --type=merge >/dev/null 2>&1 || true
    kubectl delete pvc postgres-pvc -n "$NAMESPACE" --grace-period=0 >/dev/null 2>&1 || true
  fi

  echo "✅ Database will be recreated fresh on next start"
else
  echo ""
  # Run local PostgreSQL shutdown flow before final message
  stop_local_postgres

  echo "✅ Smart Service Discovery Environment Stopped!"
  echo ""
  echo "💾 Data volumes preserved - database will persist on restart"
  echo "💡 To start fresh with clean database, run: ./stop.sh --clean"
fi
