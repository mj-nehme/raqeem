# Release v2.0.0 Planning Document

## Overview

This document tracks the planning and preparation for release v2.0.0.

**Target Release Date**: TBD  
**Version**: 2.0.0  
**Status**: ✅ **VERSION UPDATES COMPLETE** - Ready for feature development and testing

**Completed**: 2025-11-17  
**Next Steps**: Define features, test, and finalize release

---

## What's Been Completed

✅ **Version Synchronization**: All components updated to version 2.0.0
- VERSION file: 2.0.0
- mentor/frontend/package.json: 2.0.0
- devices/frontend/package.json: 2.0.0
- charts/devices-backend/Chart.yaml: 2.0.0
- charts/mentor-backend/Chart.yaml: 2.0.0
- charts/postgres/Chart.yaml: 2.0.0
- charts/minio/Chart.yaml: 2.0.0

✅ **Documentation**: Release planning documents created
- CHANGELOG.md updated with v2.0.0 section
- RELEASE_CHECKLIST.md updated for v2.0.0
- RELEASE_v2.0.0.md created for tracking

✅ **Validation**: All configuration files validated
- JSON files (package.json): Syntax verified
- YAML files (Chart.yaml): Syntax verified
- Python setup.py: Version correctly reads from VERSION file

---

## Release Scope

Version 2.0.0 represents a major release following the semantic versioning specification. This version bump indicates:

- Potential breaking changes to APIs or behavior
- Significant new features or architectural changes
- Major improvements to existing functionality

### Key Considerations for v2.0.0

1. **Breaking Changes**: Any breaking changes must be clearly documented
2. **Migration Guide**: Users upgrading from v0.2.0 need clear migration instructions
3. **Deprecation Notices**: Features to be removed should be marked as deprecated
4. **Backward Compatibility**: Document any compatibility breaks with previous versions

---

## Proposed Changes for v2.0.0

### 1. Version Management
- [x] Update VERSION file from 0.2.0 to 2.0.0
- [x] Update CHANGELOG.md with v2.0.0 section
- [x] Update package.json files (mentor-frontend, devices-frontend)
- [x] Update Helm Chart.yaml files (devices-backend, mentor-backend, postgres, minio)
- [x] Verify version consistency across all components

### 2. Breaking Changes (if any)
- [ ] Document API breaking changes
- [ ] Update API version if needed
- [ ] Create migration guide
- [ ] Update OpenAPI specs

### 3. New Features
- [ ] List new features to be included
- [ ] Ensure features are tested
- [ ] Update documentation for new features

### 4. Improvements
- [ ] Performance improvements
- [ ] Security enhancements
- [ ] Developer experience improvements

### 5. Deprecations
- [ ] Mark deprecated features
- [ ] Provide alternatives
- [ ] Set timeline for removal

---

## Version Update Checklist

### Core Version Files
- [x] `/VERSION` - Main version file (2.0.0)
- [x] `/pyproject.toml` - Python package version (dynamic from VERSION)
- [x] `/setup.py` - Setup.py reads from VERSION

### Frontend Packages
- [x] `/mentor/frontend/package.json` - Mentor frontend version
- [x] `/devices/frontend/package.json` - Devices frontend version

### Helm Charts
- [x] `/charts/devices-backend/Chart.yaml` - Devices backend chart
- [x] `/charts/mentor-backend/Chart.yaml` - Mentor backend chart
- [x] `/charts/postgres/Chart.yaml` - Postgres chart
- [x] `/charts/minio/Chart.yaml` - MinIO chart

### Documentation
- [x] `/CHANGELOG.md` - Add v2.0.0 section
- [ ] `/README.md` - Update version badges and references
- [ ] `/RELEASE_CHECKLIST.md` - Update for v2.0.0 specifics

---

## Pre-Release Verification

### Code Quality
- [ ] All tests pass (backend + frontend + integration)
- [ ] Linting clean (ruff, golangci-lint, ESLint)
- [ ] Type checking passes (mypy)
- [ ] No critical TODOs or FIXMEs
- [ ] Code review completed

### Documentation
- [ ] README.md updated
- [ ] CHANGELOG.md complete
- [ ] API documentation current
- [ ] Migration guide created (if breaking changes)
- [ ] All docs reviewed

### Build & Test
- [ ] Python backend builds successfully
- [ ] Go backend compiles
- [ ] Frontend builds complete
- [ ] Docker images build
- [ ] Helm charts validate
- [ ] Integration tests pass

### Security
- [ ] Dependency audit completed
- [ ] No known vulnerabilities
- [ ] Security review done
- [ ] Secrets properly managed

---

## Release Process

### 1. Pre-Release Testing
```bash
# Start environment with current code
./start.sh

# Run manual tests
# - Test device registration
# - Test telemetry ingestion
# - Test dashboard display
# - Test API endpoints

# Run automated tests
./scripts/health-check.sh

# Stop environment
./stop.sh
```

### 2. Create Release Tag
```bash
# Validate and tag release
./scripts/tag-release.sh v2.0.0

# Push to GitHub
git push origin v2.0.0
git push origin master
```

### 3. Build and Push Container Images
```bash
# Images will be tagged as:
# - ghcr.io/mj-nehme/raqeem/devices-backend:2.0.0
# - ghcr.io/mj-nehme/raqeem/devices-backend:2.0.0-<commit>
# - ghcr.io/mj-nehme/raqeem/devices-backend:latest
# - ghcr.io/mj-nehme/raqeem/mentor-backend:2.0.0
# - ghcr.io/mj-nehme/raqeem/mentor-backend:2.0.0-<commit>
# - ghcr.io/mj-nehme/raqeem/mentor-backend:latest
```

### 4. Create GitHub Release
- Go to https://github.com/mj-nehme/raqeem/releases/new
- Tag: v2.0.0
- Title: "Raqeem v2.0.0"
- Description: Copy from CHANGELOG.md
- Mark as major release

### 5. Verify Release
```bash
# Deploy with version tag
echo "IMAGE_TAG=v2.0.0" > .deploy/tag.env
./start.sh

# Verify everything works
./scripts/health-check.sh

# Check services
kubectl get pods
kubectl get services
```

---

## Post-Release Tasks

### Verification
- [ ] Release appears on GitHub
- [ ] Container images available on GHCR
- [ ] Documentation links work
- [ ] Release artifacts downloadable
- [ ] CI/CD pipeline passes

### Communication
- [ ] Update project status
- [ ] Announce release (if applicable)
- [ ] Close v2.0.0 milestone
- [ ] Update badges

### Next Steps
- [ ] Create v2.1.0 milestone
- [ ] Plan next features
- [ ] Address release feedback

---

## Rollback Plan

If critical issues are discovered post-release:

1. **Assess severity**: Determine if rollback is necessary
2. **Remove GitHub release**: If needed
3. **Delete git tag**: 
   ```bash
   git tag -d v2.0.0
   git push origin :refs/tags/v2.0.0
   ```
4. **Revert to v0.2.0**: Deploy previous stable version
5. **Fix issues**: Address problems
6. **Re-release as v2.0.1**: After fixes

---

## Breaking Changes Documentation

### API Changes
- Document any endpoint changes
- Note parameter modifications
- List removed endpoints
- Describe new requirements

### Configuration Changes
- Environment variable changes
- Configuration file format changes
- Default value changes

### Behavior Changes
- Changed functionality
- Modified defaults
- Different error handling

---

## Migration Guide (v0.2.0 → v2.0.0)

### For Users
1. Review breaking changes
2. Update configuration
3. Test in development
4. Deploy to production

### For Developers
1. Update dependencies
2. Modify code for API changes
3. Update tests
4. Update documentation

---

## Notes

- This is a major version bump (0.2.0 → 2.0.0)
- Skipping version 1.x.x indicates significant changes
- Follow semantic versioning: MAJOR.MINOR.PATCH
- Major version 2 allows for breaking changes

---

**Document Created**: 2025-11-17  
**Status**: Planning phase  
**Next Review**: After scope definition
