# Migration Guide: DockerHub to GitHub Container Registry (GHCR)

## Overview

As of v0.2.0, Raqeem has migrated from DockerHub to GitHub Container Registry (GHCR) for all container images. This guide will help you transition your deployments.

## Why We Migrated

- **Better Integration**: Seamless integration with GitHub ecosystem
- **Enhanced Security**: Built-in vulnerability scanning and image signing
- **Cost Control**: No rate limits or unexpected costs
- **Improved Reliability**: Better availability and performance

## What Changed

### Old Image Names (DockerHub)
```
jaafarn/raqeem-devices-backend:v0.1.1
jaafarn/raqeem-mentor-backend:v0.1.1
```

### New Image Names (GHCR)
```
ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0
ghcr.io/mj-nehme/raqeem/mentor-backend:v0.2.0
```

## Migration Steps

### For Local Development

If you're using the provided scripts, no action is needed:

```bash
# These scripts now use GHCR automatically
./start.sh
./stop.sh
```

### For Kubernetes Deployments

#### 1. Update Helm Chart Values

If you have custom values files, update the image repository:

**Before:**
```yaml
image:
  repository: jaafarn/raqeem-devices-backend
  tag: v0.1.1
```

**After:**
```yaml
image:
  repository: ghcr.io/mj-nehme/raqeem/devices-backend
  tag: v0.2.0
```

#### 2. Create Image Pull Secret (if needed)

For private images, create a GitHub token and configure Kubernetes:

```bash
# Create GitHub Personal Access Token with 'read:packages' scope
# Visit: https://github.com/settings/tokens

# Create Kubernetes secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<your-github-username> \
  --docker-password=<your-github-token> \
  -n <namespace>

# Update deployment to use the secret
kubectl patch serviceaccount default \
  -p '{"imagePullSecrets": [{"name": "ghcr-secret"}]}' \
  -n <namespace>
```

#### 3. Upgrade Your Deployment

```bash
# Update using Helm
helm upgrade devices-backend ./charts/devices-backend \
  --set image.tag=v0.2.0 \
  -n <namespace>

helm upgrade mentor-backend ./charts/mentor-backend \
  --set image.tag=v0.2.0 \
  -n <namespace>
```

### For Docker Compose

Update your `docker-compose.yml`:

**Before:**
```yaml
services:
  devices-backend:
    image: jaafarn/raqeem-devices-backend:v0.1.1
```

**After:**
```yaml
services:
  devices-backend:
    image: ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0
```

Then pull and restart:
```bash
docker-compose pull
docker-compose up -d
```

### For Manual Docker Commands

**Before:**
```bash
docker pull jaafarn/raqeem-devices-backend:latest
docker run jaafarn/raqeem-devices-backend:latest
```

**After:**
```bash
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest
docker run ghcr.io/mj-nehme/raqeem/devices-backend:latest
```

## Image Availability

All images are publicly available on GHCR:

- **Devices Backend**: https://github.com/mj-nehme/raqeem/pkgs/container/raqeem%2Fdevices-backend
- **Mentor Backend**: https://github.com/mj-nehme/raqeem/pkgs/container/raqeem%2Fmentor-backend

## Tagging Strategy

Images are tagged with multiple identifiers:

- **Version tags**: `v0.2.0`, `v0.2.1`, etc.
- **Branch tags**: `main`, `develop`
- **Commit tags**: `main-<short-sha>`
- **Latest tag**: `latest` (always points to the latest release)

Example:
```bash
# Pull specific version
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0

# Pull latest
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest

# Pull from main branch
docker pull ghcr.io/mj-nehme/raqeem/devices-backend:main
```

## CI/CD Integration

### GitHub Actions

Images are automatically built and pushed on every merge to main:

```yaml
- name: Log in to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    push: true
    tags: ghcr.io/${{ github.repository }}/devices-backend:latest
```

### Manual Release Process

Use the updated `tag-release.sh` script:

```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Create release
./scripts/tag-release.sh v0.2.0

# Push changes
git push origin main
git push origin v0.2.0
```

## Troubleshooting

### Cannot Pull Images

**Error**: `manifest unknown` or `unauthorized`

**Solutions**:

1. **Check image name**:
   ```bash
   docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest
   ```

2. **Login to GHCR**:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
   ```

3. **Verify image exists**:
   Visit: https://github.com/mj-nehme/raqeem/pkgs/container/raqeem%2Fdevices-backend

### Kubernetes Pod ImagePullBackOff

**Check events**:
```bash
kubectl describe pod <pod-name> -n <namespace>
```

**Create image pull secret**:
```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token> \
  -n <namespace>
```

**Update deployment**:
```yaml
spec:
  template:
    spec:
      imagePullSecrets:
      - name: ghcr-secret
```

## Legacy Support

### DockerHub Images Deprecated

As of v0.2.0, DockerHub images are no longer updated. The last available versions are:
- `jaafarn/raqeem-devices-backend:v0.1.1`
- `jaafarn/raqeem-mentor-backend:v0.1.1`

**Important**: Please migrate to GHCR images to receive updates and security patches.

## Rollback Plan

If you need to rollback temporarily:

```bash
# Use previous DockerHub image
kubectl set image deployment/devices-backend \
  devices-backend=jaafarn/raqeem-devices-backend:v0.1.1 \
  -n <namespace>
```

However, this is not recommended for long-term use as these images won't receive updates.

## Security Considerations

### Image Signing and Verification

GHCR images include provenance and SBOM (Software Bill of Materials):

```bash
# Verify image signature (requires cosign)
cosign verify ghcr.io/mj-nehme/raqeem/devices-backend:latest
```

### Vulnerability Scanning

GitHub automatically scans images for vulnerabilities. View reports at:
- https://github.com/mj-nehme/raqeem/security

### Access Control

Images are public by default. For private deployments:

1. Fork the repository
2. Configure private packages in your fork's settings
3. Use GitHub tokens for authentication

## Benefits Summary

| Feature | DockerHub | GHCR |
|---------|-----------|------|
| Rate Limits | ✗ Yes (100 pulls/6hrs) | ✓ None |
| Cost | ✗ Potential charges | ✓ Free |
| Security Scanning | ✗ Limited | ✓ Built-in |
| GitHub Integration | ✗ External | ✓ Native |
| Image Signing | ✗ Manual | ✓ Automatic |
| SBOM Support | ✗ No | ✓ Yes |

## Support

For issues or questions:
1. Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
2. Review [GitHub Issues](https://github.com/mj-nehme/raqeem/issues)
3. Create a new issue with details about your migration problem

## Related Documentation

- [Deployment Guide](DEPLOYMENT.md) - Production deployment with GHCR
- [Development Guide](DEVELOPMENT.md) - Local development setup
- [Architecture Documentation](ARCHITECTURE.md) - System design overview
