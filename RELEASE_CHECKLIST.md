# Release Checklist for v2.0.0

> **Note**: Container images are hosted on GitHub Container Registry (GHCR). See [GHCR Migration Guide](docs/GHCR_MIGRATION.md) for details.

## Pre-Release Verification

### Code Quality
- [x] All tests pass (189 backend + 76 devices frontend + 23 mentor frontend + 15 integration)
- [x] Linting clean (ruff, golangci-lint, ESLint)
- [x] Type checking passes (mypy with relaxed settings)
- [x] No TODOs or FIXMEs in production code
- [x] Code review completed

### Documentation
- [ ] README.md up to date
- [ ] CHANGELOG.md updated with v2.0.0 changes
- [ ] API documentation current
- [ ] All docs reviewed for accuracy
- [ ] CONTRIBUTING.md reviewed
- [ ] LICENSE file present (MIT)
- [ ] Migration guide created (if breaking changes)

### Version Consistency
- [ ] VERSION file set to 2.0.0
- [ ] devices/frontend/package.json version 2.0.0
- [ ] mentor/frontend/package.json version 2.0.0
- [ ] All Helm charts version 2.0.0
- [ ] CHANGELOG.md dated correctly

### Build & Packaging
- [x] setup.py created
- [x] pyproject.toml created
- [x] MANIFEST.in created
- [x] .dockerignore updated
- [x] All build artifacts cleaned

### Repository Cleanup
- [x] Temporary files removed
- [x] __pycache__ directories cleaned
- [x] .gitignore updated
- [x] No sensitive data in repo

### CI/CD
- [x] GitHub Actions workflow verified
- [x] Docker builds successful
- [x] All CI checks pass

### GitHub Setup
- [x] Issue templates created
- [x] PR template created
- [x] Branch protection configured (if applicable)

## Release Process

### 1. Final Commit
```bash
git add .
git commit -m "Release v2.0.0

- Update version to 2.0.0 across all components
- Update CHANGELOG.md with v2.0.0 release notes
- Prepare release documentation
- Major version bump to reflect platform maturity
"
```

### 2. Create Git Tag
```bash
git tag -a v2.0.0 -m "Release v2.0.0 - Major version release"
git push origin master
git push origin v2.0.0
```

### 3. GitHub Release
- Go to https://github.com/mj-nehme/raqeem/releases/new
- Tag: v2.0.0
- Title: "Raqeem v2.0.0"
- Description: Copy from CHANGELOG.md
- Upload any release artifacts

### 4. Docker Images
```bash
# Build and push Docker images to GHCR
docker build -t ghcr.io/mj-nehme/raqeem/devices-backend:2.0.0 ./devices/backend
docker build -t ghcr.io/mj-nehme/raqeem/mentor-backend:2.0.0 ./mentor/backend

# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Push images
docker push ghcr.io/mj-nehme/raqeem/devices-backend:2.0.0
docker push ghcr.io/mj-nehme/raqeem/mentor-backend:2.0.0

# Or use the automated script:
./scripts/tag-release.sh v2.0.0
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
- [ ] Create v2.1.0 milestone
- [ ] Plan next features
- [ ] Address any release feedback

## Rollback Plan

If issues are discovered:
1. Remove GitHub release
2. Delete git tag: `git tag -d v2.0.0 && git push origin :refs/tags/v2.0.0`
3. Remove Docker images from GHCR registry
4. Fix issues and restart release process

---
**Release Date**: TBD
**Release Manager**: Raqeem Team
