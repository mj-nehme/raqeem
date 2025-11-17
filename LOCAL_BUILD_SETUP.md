# Local Build Setup - No External Registry Required

## Overview

The deployment now uses **local Docker images** instead of pulling from GitHub Container Registry (ghcr.io). This eliminates external dependencies and authentication issues during development.

## Changes Made

### 1. Helm Chart Updates
- **charts/mentor-backend/values.yaml**: Changed image from `ghcr.io/mj-nehme/raqeem/mentor-backend:v0.2.0` to `raqeem/mentor-backend:latest`
- **charts/devices-backend/values.yaml**: Changed image from `ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0` to `raqeem/devices-backend:latest`
- Set `pullPolicy: Never` to force Kubernetes to use local images only

### 2. New Build Script
- **scripts/build-local-images.sh**: Builds both backend images locally with tags:
  - `raqeem/devices-backend:latest`
  - `raqeem/mentor-backend:latest`

### 3. Start Script Integration
- **scripts/start-smart.sh**: Now automatically builds local images before deploying to Kubernetes

## Usage

### Automatic (Recommended)
Simply run the start script as usual:
```bash
./start.sh
```

The script will:
1. Build local images automatically
2. Deploy to Kubernetes using the local images

### Manual Build
If you need to rebuild images without redeploying:
```bash
./scripts/build-local-images.sh
```

## Benefits

1. **No External Dependencies**: No need for GitHub Container Registry authentication
2. **Faster Iteration**: Changes are reflected immediately in local builds
3. **Offline Development**: Works without internet connection (after base images are cached)
4. **Simpler Workflow**: No need to push images to remote registry during development

## Production/CI Considerations

For production deployments or CI/CD pipelines that need to use remote registries:
- The existing **scripts/tag-release.sh** still handles pushing to ghcr.io
- Update Helm values to use specific version tags (e.g., `v0.2.0`) instead of `latest`
- Set `pullPolicy: IfNotPresent` or `pullPolicy: Always` as appropriate

## Troubleshooting

### Image Not Found Error
If you see `ImagePullBackOff` errors:
```bash
# Verify local images exist
docker images | grep raqeem

# Rebuild if needed
./scripts/build-local-images.sh
```

### Old Deployments
If upgrading from ghcr.io images, you may need to delete old deployments:
```bash
kubectl delete deployment mentor-backend devices-backend
./start.sh
```
