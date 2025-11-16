# Container Registry Strategy

## Overview

As of v0.2.0, Raqeem uses a multi-registry strategy to eliminate dependencies on DockerHub while maintaining security, reliability, and transparency. This document outlines our container registry usage and rationale.

## Registry Usage by Image Type

### Application Images (Raqeem Services)

**Registry**: GitHub Container Registry (GHCR) - `ghcr.io`

All Raqeem application images are published to GHCR:

```yaml
# Devices Backend
image: ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0

# Mentor Backend  
image: ghcr.io/mj-nehme/raqeem/mentor-backend:v0.2.0
```

**Benefits**:
- ✅ Native GitHub integration
- ✅ Built-in vulnerability scanning
- ✅ No rate limits
- ✅ Free for public repositories
- ✅ Automatic CI/CD publishing

### Base Images (Build Dependencies)

**Registry**: Docker Official Images - `docker.io/library/`

Base images for building our applications use explicit registry prefixes:

```dockerfile
# Go applications
FROM docker.io/library/golang:1.25-alpine

# Python applications
FROM docker.io/library/python:3.10-slim
```

**Why Docker Official Images**:
- ✅ Maintained by Docker and upstream vendors
- ✅ Regular security updates
- ✅ Widely trusted and tested
- ✅ Available without authentication
- ⚠️ Subject to rate limits (but our CI/CD handles this)

### Third-Party Services

#### PostgreSQL

**Registry**: Docker Official Images - `docker.io/library/`

```yaml
image: docker.io/library/postgres:16
```

**Why Official Images**:
- ✅ Directly maintained by PostgreSQL team via Docker
- ✅ Regular security patches
- ✅ Most widely used and tested
- ✅ Compatible with all deployment tools

#### MinIO

**Registry**: Quay.io - `quay.io`

```yaml
image: quay.io/minio/minio:latest
```

**Why Quay.io**:
- ✅ Official MinIO registry
- ✅ Recommended by MinIO team
- ✅ No rate limits
- ✅ Better performance than DockerHub

## Why We Use Explicit Registry Prefixes

All container images now use explicit registry prefixes (e.g., `docker.io/library/postgres:16` instead of just `postgres:16`).

**Benefits**:
1. **Transparency**: Clear source of every image
2. **Security**: No ambiguity about image origin
3. **Reproducibility**: Explicit source prevents surprises
4. **Compliance**: Easier to audit image sources
5. **Documentation**: Self-documenting in manifests

## Rate Limits and Authentication

### DockerHub Rate Limits

Docker Hub enforces rate limits:
- **Anonymous users**: 100 pulls per 6 hours per IP
- **Authenticated users**: 200 pulls per 6 hours
- **Pro/Team accounts**: No limits

**Our Strategy**:
- Use Docker Official Images (`docker.io/library/*`) which have higher limits
- CI/CD caches images between builds
- Alternative registries (GHCR, Quay.io) have no rate limits

### Authentication Requirements

| Registry | Public Images | Private Images |
|----------|---------------|----------------|
| GHCR (`ghcr.io`) | No auth needed | GitHub token required |
| Docker Official (`docker.io/library/*`) | No auth needed | N/A (all public) |
| Quay.io (`quay.io`) | No auth needed | Account required |

## CI/CD Configuration

### GitHub Actions

Our CI/CD workflows automatically:
1. Build application images
2. Push to GHCR on main branch
3. Tag with version, branch, and commit SHA
4. Use Docker buildx caching to minimize pulls

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
    context: ./devices/backend
    push: true
    tags: ghcr.io/${{ github.repository }}/devices-backend:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Docker Buildx Caching

We use GitHub Actions caching to minimize base image pulls:
- `cache-from: type=gha` - Reuse cached layers
- `cache-to: type=gha,mode=max` - Save all layers for next build

This reduces build times and avoids rate limits.

## Security Considerations

### Image Scanning

All images are automatically scanned for vulnerabilities:

**GHCR Images** (our applications):
- GitHub automatically scans on push
- Vulnerabilities visible in Security tab
- SBOM (Software Bill of Materials) generated

**Base Images** (golang, python):
- Scanned by Docker Scout
- Regular updates from upstream
- CVE notifications available

### Image Signing (Future)

We plan to implement:
- Cosign for image signing
- Attestations for build provenance
- SLSA compliance

### Supply Chain Security

Our strategy ensures:
1. **Trusted Sources**: Only official registries
2. **Explicit References**: No implicit DockerHub lookups
3. **Version Pinning**: Specific tags, not just `latest`
4. **Automated Updates**: Dependabot for base image updates

## Migration from DockerHub

### What Changed

**Before (v0.1.x)**:
```yaml
# Old DockerHub images
image: jaafarn/raqeem-devices-backend:v0.1.1
image: postgres:16
image: minio/minio:latest
```

**After (v0.2.0+)**:
```yaml
# New GHCR and explicit registry images
image: ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0
image: docker.io/library/postgres:16
image: quay.io/minio/minio:latest
```

### Migration Guide

See [GHCR_MIGRATION.md](./GHCR_MIGRATION.md) for detailed migration instructions.

## Production Deployment

### Kubernetes

Update your Helm values:

```yaml
# Application images
devices-backend:
  image:
    repository: ghcr.io/mj-nehme/raqeem/devices-backend
    tag: v0.2.0

# Infrastructure images
postgres:
  image:
    repository: docker.io/library/postgres
    tag: "16"

minio:
  image:
    repository: quay.io/minio/minio
    tag: latest
```

### Docker Compose

```yaml
services:
  devices-backend:
    image: ghcr.io/mj-nehme/raqeem/devices-backend:v0.2.0
  
  postgres:
    image: docker.io/library/postgres:16
  
  minio:
    image: quay.io/minio/minio:latest
```

## Image Lifecycle Management

### Tagging Strategy

**Application Images**:
- `v0.2.0` - Specific release
- `v0.2` - Minor version tracking
- `latest` - Latest release
- `main` - Latest from main branch
- `main-abc1234` - Specific commit

**Infrastructure Images**:
- `16` - Major version (PostgreSQL)
- `16.1` - Minor version
- `latest` - Latest stable (use with caution)

### Retention Policy

**GHCR**:
- Keep all tagged releases indefinitely
- Keep `main` branch images for 90 days
- Keep PR images for 7 days

**Local Development**:
- Prune unused images regularly:
  ```bash
  docker image prune -a --filter "until=720h"
  ```

## Troubleshooting

### Cannot Pull Images

**Error**: `manifest unknown` or `unauthorized`

**Solutions**:

1. **Check image exists**:
   ```bash
   docker manifest inspect ghcr.io/mj-nehme/raqeem/devices-backend:latest
   ```

2. **Login to GHCR** (if private):
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
   ```

3. **Use correct registry**:
   ```bash
   # Correct
   docker pull ghcr.io/mj-nehme/raqeem/devices-backend:latest
   
   # Incorrect (old)
   docker pull jaafarn/raqeem-devices-backend:latest
   ```

### Rate Limit Exceeded

**Error**: `toomanyrequests: You have reached your pull rate limit`

**Solutions**:

1. **Login to Docker Hub**:
   ```bash
   docker login
   ```

2. **Use alternative registry**:
   - For application images: Use GHCR (no limits)
   - For infrastructure: Images cached by CI/CD

3. **Implement local registry mirror**:
   ```bash
   docker run -d -p 5000:5000 --restart=always --name registry registry:2
   ```

### Kubernetes ImagePullBackOff

**Check pod events**:
```bash
kubectl describe pod <pod-name>
```

**Common causes**:
1. Wrong image name/tag
2. Missing imagePullSecret (for private images)
3. Network issues

**Fix**:
```bash
# Create GHCR secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token>

# Patch service account
kubectl patch serviceaccount default \
  -p '{"imagePullSecrets": [{"name": "ghcr-secret"}]}'
```

## Best Practices

### Development

✅ **DO**:
- Use specific version tags in production
- Use explicit registry prefixes
- Test with the same images as production
- Keep local images up to date

❌ **DON'T**:
- Use `latest` tag in production
- Rely on implicit DockerHub lookups
- Mix registry sources unnecessarily
- Ignore image security warnings

### Production

✅ **DO**:
- Pin specific versions
- Enable automatic vulnerability scanning
- Set up image pull secrets
- Monitor registry availability
- Have fallback/mirror registries

❌ **DON'T**:
- Use `latest` tag without pinning
- Skip security scans
- Ignore CVE notifications
- Deploy without testing images
- Forget to update documentation

## Related Documentation

- [GHCR Migration Guide](./GHCR_MIGRATION.md) - Migration from DockerHub
- [Deployment Guide](./DEPLOYMENT.md) - Production deployment
- [Development Guide](./DEVELOPMENT.md) - Local development setup
- [CI/CD Documentation](./.github/workflows/README.md) - Build pipeline

## Support

For registry-related issues:
1. Check this document first
2. Review [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
3. Search [GitHub Issues](https://github.com/mj-nehme/raqeem/issues)
4. Create new issue with `registry` label

## Changelog

### v0.2.0
- ✅ Migrated from DockerHub to GHCR for application images
- ✅ Added explicit registry prefixes for all images
- ✅ Updated documentation with registry strategy
- ✅ Configured automated GHCR publishing in CI/CD
- ✅ Implemented image vulnerability scanning
