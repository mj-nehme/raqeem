#!/usr/bin/env bash
# Common preflight checks for local dev scripts

# Fail fast on errors in subshells
set -e

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "❌ ERROR: Missing required command: $cmd"
    return 1
  fi
}

ensure_docker_running() {
  require_cmd docker || return 1
  # 'docker info' fails if daemon isn't running
  if ! docker info >/dev/null 2>&1; then
    echo "❌ ERROR: Docker daemon is not running. Start Docker Desktop and retry."
    return 1
  fi
}

ensure_helm_ready() {
  require_cmd helm || return 1
  if ! helm version --short >/dev/null 2>&1; then
    echo "❌ ERROR: Helm not responding. Verify Helm installation and kube config."
    return 1
  fi
}

ensure_kube_ready() {
  require_cmd kubectl || return 1
  if ! kubectl cluster-info >/dev/null 2>&1; then
    echo "❌ ERROR: kubectl cannot reach Kubernetes cluster. Check Docker Desktop Kubernetes or your kube context."
    return 1
  fi
}
