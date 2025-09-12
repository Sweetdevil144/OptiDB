# OptiDB Docker Setup

Simple PostgreSQL setup for database profiling.

## Quick Start

```bash
# Start database
make up

# Check if working
make status

# Connect to database
make connect
```

## What's Running

- **PostgreSQL 16** on port 5432
- **pg_stat_statements** - tracks query performance
- **auto_explain** - logs slow queries automatically

## Database Users

- `postgres` / `postgres` - admin access
- `profiler_ro` / `profiler_ro_pass` - read-only for statistics
- `profiler_sb` / `profiler_sb_pass` - sandbox for testing

## Commands

```bash
make up         # Start database
make down       # Stop database
make status     # Check if working
make connect    # Connect as admin
make connect-ro # Connect as read-only user
make connect-sb # Connect as sandbox user
make logs       # Show database logs
```

## Test It Works

```bash
# 1. Start database
make up

# 2. Check status (should show database ready + extensions + roles)
make status

# 3. Connect and test
make connect
```

In PostgreSQL:

```sql
-- Check extensions
\dx

-- Check roles
SELECT rolname FROM pg_roles WHERE rolname LIKE 'profiler_%';

-- Test query stats
SELECT count(*) FROM pg_stat_statements;

-- Exit
\q
```

## Connection Info

- **Host**: localhost:5432
- **Database**: optidb
- **Admin**: postgres/postgres
- **Read-only**: profiler_ro/profiler_ro_pass

## If Something Breaks

```bash
# Restart everything
make down
make up

# Check logs
make logs

# Nuclear option (deletes all data)
make down-clean
make up
```

That's it! Database is ready for the OptiDB application.
