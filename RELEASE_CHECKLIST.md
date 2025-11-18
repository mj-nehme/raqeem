# Release Checklist Template

> **Note**: As of v0.2.0, container images are hosted on GitHub Container Registry (GHCR). See [GHCR Migration Guide](docs/GHCR_MIGRATION.md) for details.

This is a template checklist for creating new releases. Copy and customize for each release.

## Pre-Release Verification

### Code Quality
- [ ] All tests pass
- [ ] Linting clean (ruff, golangci-lint, ESLint)
- [ ] Type checking passes (mypy)
- [ ] No TODOs or FIXMEs in production code
- [ ] Code review completed

### Documentation
- [ ] README.md up to date
- [ ] CHANGELOG.md updated with new version changes
- [ ] API documentation current
- [ ] All docs reviewed for accuracy
- [ ] VERSION file updated

### Version Consistency
- [ ] VERSION file updated to new version
- [ ] devices/frontend/package.json version updated
- [ ] mentor/frontend/package.json version updated
- [ ] All Helm charts version updated
- [ ] CHANGELOG.md dated correctly

### Build & Packaging
- [ ] All build artifacts cleaned
- [ ] Docker images build successfully

### Repository Cleanup
- [ ] Temporary files removed
- [ ] __pycache__ directories cleaned
- [ ] No sensitive data in repo

### CI/CD
- [ ] GitHub Actions workflow verified
- [ ] Docker builds successful
- [ ] All CI checks pass

## Release Process

### 1. Final Commit
```bash
git add .
git commit -m "Release v<VERSION>

- <Summary of changes>
- <Key features>
- <Important updates>
"
```

### 2. Create Git Tag
```bash
git tag -a v<VERSION> -m "Release v<VERSION> - <Brief description>"
git push origin master
git push origin v<VERSION>
```

### 3. GitHub Release
- Go to https://github.com/mj-nehme/raqeem/releases/new
- Tag: v<VERSION>
- Title: "Raqeem v<VERSION> - <Release Name>"
- Description: Copy from CHANGELOG.md
- Upload any release artifacts

### 4. Docker Images
```bash
# Build and push Docker images to GHCR
docker build -t ghcr.io/mj-nehme/raqeem/devices-backend:<VERSION> ./devices/backend
docker build -t ghcr.io/mj-nehme/raqeem/mentor-backend:<VERSION> ./mentor/backend

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Push images
docker push ghcr.io/mj-nehme/raqeem/devices-backend:<VERSION>
docker push ghcr.io/mj-nehme/raqeem/mentor-backend:<VERSION>

# Or use the automated script:
./scripts/tag-release.sh v<VERSION>
```

### 5. Python Package (Optional)
```bash
# Build and publish to PyPI
python -m build
twine check dist/*
twine upload dist/*
```

## Post-Release

### Verification
- [ ] Release appears on GitHub
- [ ] Docker images available on GitHub Container Registry
- [ ] Package available on PyPI (if published)
- [ ] Documentation links work
- [ ] Download and test release artifacts

### Communication
- [ ] Announce release (if applicable)
- [ ] Update project status badges
- [ ] Close related milestone

### Next Steps
- [ ] Create next version milestone
- [ ] Plan next features
- [ ] Address any release feedback

## Rollback Plan

If issues are discovered:
1. Remove GitHub release
2. Delete git tag: `git tag -d v<VERSION> && git push origin :refs/tags/v<VERSION>`
3. Remove Docker images from GHCR registry
4. Fix issues and restart release process

---
**Last Updated**: 2025-11-18
