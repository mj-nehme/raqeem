# Release Checklist for v0.1.0

## Pre-Release Verification

### Code Quality
- [x] All tests pass (189 backend + 76 devices frontend + 23 mentor frontend + 15 integration)
- [x] Linting clean (ruff, golangci-lint, ESLint)
- [x] Type checking passes (mypy with relaxed settings)
- [x] No TODOs or FIXMEs in production code
- [x] Code review completed

### Documentation
- [x] README.md up to date
- [x] CHANGELOG.md updated with v0.1.0 changes
- [x] API documentation current
- [x] All docs reviewed for accuracy
- [x] CONTRIBUTING.md created
- [x] LICENSE file present (MIT)

### Version Consistency
- [x] VERSION file set to 0.1.0
- [x] devices/frontend/package.json version 0.1.0
- [x] mentor/frontend/package.json version 0.1.0
- [x] All Helm charts version 0.1.0
- [x] CHANGELOG.md dated correctly

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
git commit -m "Release v0.1.0

- Clean repository for first stable release
- Add packaging configuration
- Update all documentation
- Ensure test coverage across all components
"
```

### 2. Create Git Tag
```bash
git tag -a v0.1.0 -m "Release v0.1.0 - Initial stable release"
git push origin master
git push origin v0.1.0
```

### 3. GitHub Release
- Go to https://github.com/mj-nehme/raqeem/releases/new
- Tag: v0.1.0
- Title: "Raqeem v0.1.0 - Initial Release"
- Description: Copy from CHANGELOG.md
- Upload any release artifacts

### 4. Docker Images
```bash
# Build and push Docker images
docker build -t jaafarn/raqeem-devices-backend:0.1.0 ./devices/backend
docker build -t jaafarn/raqeem-mentor-backend:0.1.0 ./mentor/backend
docker push jaafarn/raqeem-devices-backend:0.1.0
docker push jaafarn/raqeem-mentor-backend:0.1.0
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
- [ ] Docker images available on Docker Hub
- [ ] Package available on PyPI (if published)
- [ ] Documentation links work
- [ ] Download and test release artifacts

### Communication
- [ ] Announce release (if applicable)
- [ ] Update project status badges
- [ ] Close related milestone

### Next Steps
- [ ] Create v0.2.0 milestone
- [ ] Plan next features
- [ ] Address any release feedback

## Rollback Plan

If issues are discovered:
1. Remove GitHub release
2. Delete git tag: `git tag -d v0.1.0 && git push origin :refs/tags/v0.1.0`
3. Remove Docker images from registry
4. Fix issues and restart release process

---
**Release Date**: 2025-11-15
**Release Manager**: Raqeem Team
