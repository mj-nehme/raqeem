# Planning for v2.0.0 Release

## Overview

This document outlines the planning for the v2.0.0 release of Raqeem.

**Current Version**: 0.2.0 (Released 2025-11-16)  
**Target Version**: 2.0.0  
**Status**: 🔵 **PLANNING PHASE**

---

## Background

Version 0.2.0 was successfully released on 2025-11-16 with the following features:
- ✅ Removed DockerHub integration
- ✅ Consistent Swagger/OpenAPI documentation for both backends
- ✅ 90% test coverage target
- ✅ Enhanced reliability and battle-tested end-to-end flows

---

## Requirements for v2.0.0

Based on milestone requirements and user feedback, v2.0.0 should include:

### 1. ✅ Boost Reliability
**Status**: Already implemented in v0.2.0
- Comprehensive test suite (189 backend + 99 frontend + 15 integration tests)
- CI/CD pipeline with automated testing
- Coverage reporting to Codecov

### 2. ✅ Battle Tested Version
**Status**: Already implemented in v0.2.0
- End-to-end testing infrastructure
- Integration tests with PostgreSQL
- Smoke testing capabilities

### 3. ✅ Add Consistent Swagger Support
**Status**: Already implemented in v0.2.0
- **Devices Backend**: FastAPI auto-generated docs at `/docs` and `/redoc`
- **Mentor Backend**: Swagger UI at `/swagger/index.html` and `/docs` redirect

### 4. ✅ Coverage 90%
**Status**: Already maintained in v0.2.0
- Codecov configuration with 90% target
- Coverage uploaded for all four components
- Threshold set to 5% variance

### 5. ✅ Remove Connection to DockerHub
**Status**: Already completed in v0.2.0
- DockerHub push job removed from CI pipeline
- Images now hosted on GitHub Container Registry (GHCR)
- Migration guide documented in `docs/GHCR_MIGRATION.md`

### 6. 🔴 Let Commands Appear on the Frontend
**Status**: NOT IMPLEMENTED - Key new feature for v2.0.0

This is the primary new feature that differentiates v2.0.0 from v0.2.0.

**Requirements**:
- Display remote commands in the frontend UI
- Show command history and status
- Allow viewing command results
- Provide command execution feedback

**Technical Considerations**:
- Need to determine which frontend (Devices or Mentor)
- Command data model and API endpoints
- UI/UX design for command display
- Real-time updates vs polling
- Command filtering and search

**Implementation Steps**:
1. Design command display UI/UX
2. Implement API endpoints for command retrieval
3. Add frontend components for command display
4. Implement real-time updates (if needed)
5. Add tests for command display functionality
6. Update documentation

---

## Additional Considerations for v2.0.0

### Semantic Versioning
Since v2.0.0 is a major version bump, we should consider:
- Any breaking changes to APIs
- Migration path from v0.2.0
- Deprecation notices for removed features

### Breaking Changes (if any)
- Document any API changes
- Provide migration guide
- Update all affected documentation

### New Features Beyond Commands
- Are there other features to include?
- Performance improvements?
- Security enhancements?

---

## Release Checklist (To Be Completed)

### Pre-Development
- [ ] Finalize command frontend feature scope
- [ ] Create detailed technical design for command feature
- [ ] Identify any other features for v2.0.0
- [ ] Document breaking changes (if any)

### Development Phase
- [ ] Implement command display on frontend
- [ ] Add necessary API endpoints
- [ ] Write tests for new features
- [ ] Update documentation
- [ ] Review code quality and coverage

### Pre-Release Phase
- [ ] Update VERSION to 2.0.0
- [ ] Update CHANGELOG.md with v2.0.0 changes
- [ ] Update all package.json and Chart.yaml files
- [ ] Run full test suite
- [ ] Perform end-to-end testing
- [ ] Build Docker images
- [ ] Update RELEASE_CHECKLIST.md

### Release Phase
- [ ] Create git tag v2.0.0
- [ ] Push images to GHCR
- [ ] Create GitHub release
- [ ] Update documentation
- [ ] Announce release

---

## Questions to Resolve

1. **Command Frontend**: Which frontend should display commands?
   - Mentor Dashboard (for monitoring all devices)
   - Devices Simulator (for individual device commands)
   - Both?

2. **Command Data**: What command information should be displayed?
   - Command text/content
   - Execution status
   - Timestamps
   - Results/output
   - Device association

3. **Real-time**: Should commands appear in real-time?
   - WebSocket connection
   - Server-Sent Events (SSE)
   - Polling mechanism

4. **Scope**: Are there other features for v2.0.0?
   - Performance improvements
   - New monitoring capabilities
   - Additional integrations

---

## Timeline

**Planning Phase**: Current  
**Development Phase**: TBD  
**Testing Phase**: TBD  
**Release Date**: TBD

---

## References

- [v0.2.0 Release Report](./RELEASE_v0.2.0.md) - Previous release details
- [Milestone 2](https://github.com/mj-nehme/raqeem/milestone/2) - v0.2.0 requirements
- [CHANGELOG.md](./CHANGELOG.md) - Project changelog
- [RELEASE_CHECKLIST.md](./RELEASE_CHECKLIST.md) - Release process

---

**Document Created**: 2025-11-17  
**Status**: Planning in progress  
**Next Review**: After scope clarification
