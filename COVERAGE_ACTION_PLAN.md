# Test Coverage Action Plan - Path to 90%

## Current Status (as of 2025-11-18)

### Infrastructure ✅ COMPLETE
- **Codecov Integration**: Fully configured and operational
- **CI Pipeline**: Coverage uploaded for all 4 components with proper flags
- **Coverage Enforcement**: CI now fails if coverage drops >5% from baseline
- **Coverage Badges**: Visible in README.md
- **Target**: 90% set in codecov.yml

### Current Coverage by Component

| Component | Current | Target | Gap | Status |
|-----------|---------|--------|-----|--------|
| Mentor Frontend | ~75% | 90% | +15% | ⚠️ Tests Failing |
| Go Backend | ~41% | 90% | +49% | ⚠️ Requires PostgreSQL |
| Python Backend | ~71% | 90% | +19% | ⚠️ Requires PostgreSQL |
| Devices Frontend | 62.31% | 90% | +27.69% | ✅ Tests Passing |

**Note**: Backend tests require PostgreSQL database connection to run locally.

## Blocking Issues

### 1. Mentor Frontend Test Failures
**Impact**: 4 tests failing in `DeviceDashboard.extended.test.jsx`
**Failing Tests**:
- `displays commands tab and allows sending commands`
- `sends command on Enter key press`
- `displays no commands message when commands array is empty`
- `renders different device icons based on device type`

**Cause**: Likely component implementation changes or mock setup issues

**Fix Required**: Debug and fix test mocks or update tests to match current component behavior

### 2. Backend Tests Require Database
**Impact**: Cannot run Go or Python backend tests locally without PostgreSQL
**Solution**: 
- Tests run successfully in CI with PostgreSQL service
- For local development, use Docker: `docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:16`

## Detailed Coverage Gaps

### Devices Frontend (62.31% → 90%)

**Effort Estimate**: 8-10 hours
**Uncovered Areas**:
- Simulation interval logic (lines 368-380)
- Auto-polling mechanism (lines 383-390)
- Random data generation paths
- Cleanup and unmount scenarios

**Recommended Tests**:
1. Test `startSimulation()` and `stopSimulation()` functions
2. Test command polling interval with `vi.useFakeTimers()`
3. Test simulation cleanup on component unmount
4. Test error handling in async operations
5. Mock `Math.random()` for deterministic random data tests

**Files to Modify**:
- `devices/frontend/src/components/DeviceSimulator.test.jsx`

### Python Backend (71% → 90%)

**Effort Estimate**: 12-15 hours
**Major Gaps**:
```
app/api/v1/endpoints/devices.py: 51% (128/262 lines uncovered)
app/services/metrics_service.py: 43%
app/services/security_service.py: 40%
app/services/activity_service.py: 45%
app/services/device_service.py: 57%
```

**Recommended Tests**:
1. Add comprehensive endpoint tests for `devices.py`:
   - POST/PUT/DELETE operations
   - Error handling paths
   - Validation logic
   - Concurrent operations
2. Expand service layer tests:
   - All validation methods
   - Error scenarios
   - Edge cases (negative values, boundaries)
3. Integration tests:
   - Database operations
   - Alert forwarding to Mentor API
   - Screenshot handling with MinIO

**Files to Create/Modify**:
- `devices/backend/src/tests/api/v1/endpoints/test_devices_extended.py`
- `devices/backend/src/tests/services/test_metrics_service.py`
- `devices/backend/src/tests/services/test_security_service.py`
- `devices/backend/src/tests/services/test_activity_service.py`

### Go Backend (41% → 90%)

**Effort Estimate**: 4-6 hours
**Note**: Current 41% is misleading - controllers show 0% because tests are skipped without database

**Actual Status**: When tests run in CI with PostgreSQL:
- Controllers: Various percentages (60-88%)
- Database: 38.2%
- Logging: 46.9%
- Models: 92.9%
- Reliability: 85.4%
- Router: 71.3%
- S3: 70.0%
- Util: 83.8%

**Recommended Tests**:
1. Database error scenario tests
2. Pagination/filtering edge cases
3. Validation tests for models
4. Error path coverage in controllers
5. Concurrent access patterns

**Files to Modify**:
- Add tests to existing `mentor/backend/src/controllers/*_test.go` files
- Add tests to `mentor/backend/src/database/*_test.go`

### Mentor Frontend (75% → 90%)

**Effort Estimate**: 2-4 hours (after fixing failing tests)

**Current Blocker**: 4 failing tests must be fixed first

**Uncovered Areas**:
- Lines 132, 167-188, 335, 440-477 in DeviceDashboard.jsx
- Error handling paths
- Edge cases in device selection
- Command tab functionality

**Recommended Tests**:
1. Fix existing failing tests
2. Add error handling tests
3. Test edge cases in device list rendering
4. Test command execution flow
5. Test screenshot display functionality

## Implementation Strategy

### Phase 1: Quick Wins (1 week)
**Priority**: Fix blocking issues and low-hanging fruit

1. **Fix Mentor Frontend Tests** (4-6 hours)
   - Debug and fix 4 failing tests
   - Restore coverage measurement capability

2. **Add Devices Frontend Tests** (4-6 hours)
   - Focus on simulation start/stop
   - Add interval tests with fake timers
   - Target: 75-80% coverage

**Expected Outcome**: 2 out of 4 components at 75%+

### Phase 2: Backend Focus (2 weeks)
**Priority**: Improve backend coverage

1. **Python Backend** (Week 1)
   - Comprehensive `devices.py` endpoint tests
   - Service layer tests
   - Target: 85-90% coverage

2. **Go Backend** (Week 2)
   - Database error scenarios
   - Controller edge cases
   - Target: 90%+ coverage

**Expected Outcome**: All backends at 85%+

### Phase 3: Final Push (1 week)
**Priority**: Polish to 90%+

1. **Devices Frontend Polish**
   - Add remaining uncovered paths
   - Async operation tests
   - Target: 90%+ coverage

2. **Mentor Frontend Polish**
   - Complete uncovered lines
   - Target: 90%+ coverage

**Expected Outcome**: All 4 components at 90%+

## Total Effort Estimate

| Phase | Duration | Effort |
|-------|----------|--------|
| Phase 1: Quick Wins | 1 week | 8-12 hours |
| Phase 2: Backend Focus | 2 weeks | 16-21 hours |
| Phase 3: Final Push | 1 week | 6-10 hours |
| **Total** | **4 weeks** | **30-43 hours** |

## Success Criteria

- [ ] All 4 components achieve ≥90% coverage
- [ ] Zero failing tests across all components
- [ ] CI reports coverage on every PR
- [ ] CI fails if coverage drops >5%
- [ ] Coverage badges show current state
- [ ] Test quality is maintained (no low-value tests)

## Testing Best Practices

### Frontend (React + Vitest)
```javascript
// Use fake timers for interval testing
vi.useFakeTimers()
// ... test code ...
vi.advanceTimersByTime(5000)
vi.useRealTimers()

// Mock Math.random for deterministic tests
const mockRandom = vi.spyOn(Math, 'random')
mockRandom.mockReturnValue(0.5)
```

### Backend (Python + pytest)
```python
# Use async fixtures for database tests
@pytest.fixture
async def db_session():
    async with AsyncSession() as session:
        yield session
        await session.rollback()

# Mock external API calls
@patch('app.services.device_service.httpx.AsyncClient.post')
async def test_alert_forwarding(mock_post, db_session):
    mock_post.return_value = Response(200, json={"success": True})
    # ... test code ...
```

### Backend (Go + testify)
```go
// Use transaction-based isolation
func setupTestDB(t *testing.T) *gorm.DB {
    db := database.GetTestDB(t)
    tx := db.Begin()
    t.Cleanup(func() {
        tx.Rollback()
    })
    return tx
}

// Test error paths
t.Run("handles database error", func(t *testing.T) {
    // Use mock or intentionally trigger error
    // Assert proper error handling
})
```

## References

- [codecov.yml](./codecov.yml) - Codecov configuration
- [.github/workflows/ci.yml](./.github/workflows/ci.yml) - CI pipeline configuration
- [docs/TESTING.md](./docs/TESTING.md) - Testing guide
- [docs/TEST_COVERAGE_REPORT.md](./docs/TEST_COVERAGE_REPORT.md) - Coverage report

## Next Steps

1. **Immediate**: Fix Mentor Frontend failing tests to unblock coverage measurement
2. **Short-term**: Add Devices Frontend interval/simulation tests (quick wins)
3. **Medium-term**: Systematic backend test expansion
4. **Long-term**: Maintain 90%+ coverage as codebase evolves

---

**Document Created**: 2025-11-18  
**Status**: Infrastructure complete, test writing in progress  
**Next Review**: After Phase 1 completion
