# Database Schema

This directory contains the database schema for the Raqeem monitoring system.

## Files

- `modern_schema.sql` - The authoritative database schema for PostgreSQL
  - Creates all tables with proper UUID types and indexes
  - Matches current SQLAlchemy models exactly
  - Production-ready schema for clean deployments

## Usage

### For Fresh Database Setup
```bash
# Initialize database with complete schema
kubectl exec -i deployment/postgres -- psql -U monitor -d monitoring_db < db/modern_schema.sql
```

### Schema Design

The schema uses a **device-centric architecture**:

- **Primary Key Types**: UUID for most tables, TEXT for device IDs to allow arbitrary identifiers
- **Device Identification**: `device_id` as TEXT strings (not foreign keys)
- **Reserved Word Handling**: PostgreSQL reserved words are avoided or properly quoted
- **Indexing**: Performance indexes on commonly queried columns

### Key Tables

- `devices` - Core device registry
- `device_metrics` - Time-series performance data
- `device_processes` - Running process information
- `device_activity` - User activity logs
- `device_alerts` - System alerts and notifications
- `screenshots` - User screenshot storage
- `remote_commands` - Device management commands

All tables include proper timestamps and UUID primary keys for scalability and uniqueness.