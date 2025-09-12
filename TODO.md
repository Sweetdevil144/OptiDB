# OptiDB Development TODO - Team Coordination

## Day 1 Tasks (24 hours total)

### ✅ COMPLETED (0-2h): Docker Setup

#### Abhi (Data/Rules/DB)

- [x] Docker-compose Postgres 16
- [x] Enable pg_stat_statements, auto_explain extensions
- [x] Create roles (profiler_ro, profiler_sb)
- [x] Basic health checks and connection testing

#### Dev (API/UI/CLI)

- [x] Scaffold Fiber + Cobra projects
- [x] Environment/config setup
- [x] pgx pool initialization
- [x] Basic health endpoint

### ✅ COMPLETED (2-5h): Demo Schema + Slow Queries

#### Abhi (Data/Rules/DB)

- [x] Create demo tables (users, orders, order_items, events)
- [x] Insert realistic dummy data (30 users, 30 orders, 51 items, 34 events)
- [x] Add intentional performance problems:
  - [x] Missing index on users.email → seq scan
  - [x] Missing index on orders.user_id → seq scan
  - [x] Missing composite index on (user_id, status) → bad joins
  - [x] Correlated subqueries → inefficient queries
  - [x] Text search without index → slow LIKE queries
  - [x] JSON queries without GIN index → slow JSON ops
- [x] Execute slow queries 10x each to build pg_stat_statements data
- [x] Verify pg_stat_statements capturing slow queries (8.7ms avg for worst query)

#### Dev (API/UI/CLI)

- [x] CLI: `init`, `scan`, `bottlenecks` commands
- [x] Wire `scan` to call API endpoints
- [x] Basic command structure and help

### 🔄 NEXT (5-9h): Ingest & API Foundation

#### Abhi (Data/Rules/DB)

- [ ] Create `/ingest` module to pull pg_stat_statements
- [ ] Join with pg_class, pg_index for table/index metadata
- [ ] Persist to meta store (create simplified schema)
- [ ] Build query fingerprinting logic
- [ ] Create data access layer for query stats

#### Dev (API/UI/CLI)

- [ ] `/http`: `GET /bottlenecks`, `GET /queries/:id` endpoints
- [ ] Server-rendered dashboard (HTMX) with top N bottlenecks
- [ ] DTOs for bottlenecks, query detail
- [ ] Simple plan facts chips (Seq/Index, est vs act)

### 📋 PENDING (9-14h): Parse & UI Integration

#### Abhi (Data/Rules/DB)

- [ ] Create `/parse` module for query normalization
- [ ] Implement query fingerprinting (hash-based)
- [ ] Optional: AST parsing via pg_query_go (skip if time-pressed)
- [ ] Build query similarity detection

#### Dev (API/UI/CLI)

- [ ] Wire rules to UI + CLI output
- [ ] Table of recommendations with "Why / DDL / Risk" columns
- [ ] HTMX integration for dynamic updates
- [ ] Basic styling and layout

### 📋 PENDING (14-20h): Rules Engine v1

#### Abhi (Data/Rules/DB)

- [ ] Create `/rules` module with heuristics:
  - [ ] Missing index detection (seq scan on big tables)
  - [ ] Composite join index suggestions
  - [ ] Correlated subquery detection (regex/AST)
  - [ ] Redundant index detection
  - [ ] Cardinality skew detection (est vs actual rows)
- [ ] Create `/recommend` module:
  - [ ] DDL generation for index recommendations
  - [ ] Rationale generation (plain English explanations)
  - [ ] Confidence scoring (0.0-1.0)

#### Dev (API/UI/CLI)

- [ ] CLI demo script `scan→bottlenecks`
- [ ] Minimal README documentation
- [ ] Error handling and logging
- [ ] Basic testing framework

### 📋 PENDING (20-24h): Testing & Tuning

#### Abhi (Data/Rules/DB)

- [ ] Smoke test rules engine on seeded data
- [ ] Adjust detection thresholds based on results
- [ ] Validate recommendations make sense
- [ ] Performance tune the analysis pipeline

#### Dev (API/UI/CLI)

- [ ] End-to-end testing
- [ ] Performance validation
- [ ] UI responsiveness testing
- [ ] CLI output formatting

## Day 2 Tasks (24-48h)

### 📋 PENDING (24-30h): Impact Simulator Setup

#### Abhi (Data/Rules/DB)

- [ ] Create `/simulate` module
- [ ] Implement baseline EXPLAIN (ANALYZE, BUFFERS) capture
- [ ] Add hypopg extension for hypothetical indexes

#### Dev (API/UI/CLI)

- [ ] `/features`: TF-IDF implementation
- [ ] `/ml`: K-Means families
- [ ] Label via table/verb bigrams

### 📋 PENDING (30-38h): Hypopg Integration

#### Abhi (Data/Rules/DB)

- [ ] Implement hypopg_create_index() workflows
- [ ] Re-run EXPLAIN with hypothetical indexes
- [ ] Compute improvement percentages
- [ ] Capture before/after plan diffs
- [ ] Add cleanup logic for hypopg state

#### Dev (API/UI/CLI)

- [ ] Per-family MAD/IQR anomalies
- [ ] Expose tags in `/queries/:id` + `/bottlenecks`
- [ ] UI polish: Before/After cards with %Δ badge
- [ ] Plan snippet diff (node type change badges)

### 📋 PENDING (38-44h): Simulator Hardening

#### Abhi (Data/Rules/DB)

- [ ] Add guards: timeouts, concurrency caps
- [ ] Implement rollback on errors
- [ ] Create unit tests for rules + simulate
- [ ] Add error handling and logging

#### Dev (API/UI/CLI)

- [ ] Confidence & risk notes in UI
- [ ] Update CLI: `simulate` command
- [ ] Improve table formatting
- [ ] Error handling and user feedback

### 📋 PENDING (44-48h): Real Mode Support

#### Abhi (Data/Rules/DB)

- [ ] Add optional mode:"real" on sandbox schema
- [ ] Implement safe real-index testing
- [ ] Add safety checks and rollback

#### Dev (API/UI/CLI)

- [ ] Real mode UI controls
- [ ] Safety warnings and confirmations
- [ ] Progress indicators for long operations

## Day 3 Tasks (48-72h)

### 📋 PENDING (48-54h): Operational Hardening

#### Abhi (Data/Rules/DB)

- [ ] Implement caching for schema/stat calls (2-5 min TTL)
- [ ] Add pg_locks summary for contention detection
- [ ] Add "contention suspected" labels
- [ ] Implement EXPLAIN timeouts and rate limits

#### Dev (API/UI/CLI)

- [ ] `/chat`: template-grounded answers
- [ ] Pulls query metrics, plan facts, and recommended DDL
- [ ] Returns cited explanation (no hallucinations)
- [ ] Q&A interface integration

### 📋 PENDING (54-66h): Documentation & Demo

#### Abhi (Data/Rules/DB)

- [ ] Finalize Make targets: up/init/seed/scan/demo/test
- [ ] Stabilize detection thresholds and defaults
- [ ] Performance optimization

#### Dev (API/UI/CLI)

- [ ] Docs: README (90-sec Quick Start)
- [ ] ARCHITECTURE (diagram + flow)
- [ ] DEMO (script + screenshots/GIFs)
- [ ] User interface polish

### 📋 PENDING (66-72h): Final Testing

#### Abhi (Data/Rules/DB)

- [ ] Run full dry-run
- [ ] Capture screenshots
- [ ] Trim logs and optimize

#### Dev (API/UI/CLI)

- [ ] Demo rehearsal: seed → scan → bottleneck → simulate → chat
- [ ] Performance validation on acceptance targets
- [ ] Final UI/UX polish
- [ ] Documentation review

## Current Status: ✅ 2/5h Day 1 Complete

**Next Priority**:

- **Abhi**: Start `/ingest` module to pull pg_stat_statements data
- **Dev**: Build API endpoints and basic UI dashboard

## Performance Targets

- [ ] Scan 100 queries ≤ 2s (warm cache)
- [ ] Top recs precision ≥ 80% show >30% simulated speedup
- [ ] Simulator (hypopg) ≤ 1.5s round-trip per query
- [ ] UI first content paint ≤ 1s

## Test Data Available

- ✅ 30 users with realistic names/emails
- ✅ 30 orders across multiple statuses
- ✅ 51 order items with product names
- ✅ 34 events with JSON data
- ✅ Multiple slow query patterns in pg_stat_statements
- ✅ Worst query: 8.7ms avg (correlated subquery + missing indexes)
