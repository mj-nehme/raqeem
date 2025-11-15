# Database Persistence Issue - Root Cause and Solution

## Problems Identified

### 1. Duplicate Main Functions (RESOLVED)
**Issue**: There were multiple migration entry points:
- `mentor/backend/src/cmd/migrate/main.go` - Separate migration tool
- `mentor/backend/src/migrate_db.go` - Another migration script with `//go:build ignore`
- Migrations already run automatically in `database.Connect()` called by `main.go`

**Resolution**: Removed `mentor/backend/src/cmd/migrate/` directory as it was redundant. The main application automatically runs migrations on startup via `database.Connect()`.

### 2. Persistent Database Volumes (ROOT CAUSE OF STALE TABLES)
**Issue**: PostgreSQL uses Kubernetes PersistentVolumeClaims (PVC) which preserve data across restarts.

When you run:
```bash
./stop.sh && ./start.sh
```

The database volume persists, keeping all old tables and data. This is why you see stale tables even after "migrations" - the migrations add new columns/tables but don't drop existing ones.

## Solution

### Enhanced Stop Script
The `stop.sh` script now accepts a `--clean` flag:

```bash
# Stop and PRESERVE database (old behavior)
./stop.sh

# Stop and DELETE database (fresh start)
./stop.sh --clean
```

### Usage Patterns

#### Development Workflow (Preserve Data)
When you want to keep your test data between restarts:
```bash
./stop.sh      # Preserves database
./start.sh     # Uses existing data
```

#### Schema Changes / Fresh Start
When you change database models or want a clean slate:
```bash
./stop.sh --clean   # Deletes database volumes
./start.sh          # Creates fresh database with new schema
```

#### Manual PVC Cleanup
If you need to manually clear the database:
```bash
kubectl delete pvc --all -n default
```

## How Migrations Work Now

1. **Automatic on Startup**: When the mentor backend starts, `database.Connect()` automatically runs `AutoMigrate()` for all models
2. **No Manual Migration Needed**: You don't need to run any separate migration command
3. **Additive Only**: GORM's AutoMigrate only adds new tables/columns, it never drops existing ones

## Verification

After fixing, verify the setup:

```bash
# 1. Clean stop
./stop.sh --clean

# 2. Verify PVCs are gone
kubectl get pvc -n default
# Should show: No resources found

# 3. Fresh start
./start.sh

# 4. Check tables are fresh
kubectl exec -it -n default $(kubectl get pod -n default -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U monitor -d monitoring_db -c "\dt"
```

## Why This Happened

The Docker/Kubernetes setup uses persistent volumes to ensure data isn't lost when containers restart. This is correct for production but can cause confusion in development when you're changing schemas frequently.

The key insight: **Stopping and starting containers does NOT reset the database** - you must explicitly delete the volumes.

## Best Practices Going Forward

1. **Schema changes**: Always use `./stop.sh --clean` before restart
2. **Testing**: Use `./stop.sh --clean` to ensure clean test environment
3. **Development**: Use regular `./stop.sh` to preserve test data
4. **Never run manual migrations**: The app handles it automatically
