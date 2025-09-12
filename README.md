# OptiDB

**Ship**: ingest → analyze (rules + light ML) → recommend (DDL + rewrite hints) → **simulate** (before/after) → UI & CLI.
**Wow moment**: “Run Simulation” shows **−70–95%** latency with plan diff (Seq → Index).

---

## 🏗️ Architecture (lean)

- **Fiber API** (`:8080`) ←→ **Service Layer** (rules/ML) ←→ **Postgres** (target DB)
- **SQLite** (fast meta-store) or same Postgres for profiler data
- **Cobra CLI** for headless ops (`profiler scan/simulate`)
- **Optional**: MCP sidecar day-2 if ahead of schedule

```bash
/db-profiler
  /cmd
    /api        # Fiber main()
    /cli        # Cobra commands
  /internal
    /config     # env, flags
    /db         # pgx pool
    /ingest     # pg_stat_statements, plan fetch
    /parse      # norm/fingerprint, AST (pg_query_go)
    /features   # TF-IDF, metrics prep (gonum)
    /rules      # heuristics (indexes, rewrites, dup idx)
    /ml         # kmeans + MAD/IQR anomaly
    /recommend  # DDL generator + rationale
    /simulate   # EXPLAIN ANALYZE pre/post (+ hypopg)
    /store      # sqlite/meta DAO
    /http       # handlers, DTOs
    /ui         # server-rendered templates (HTMX)  ← faster than React
  /deploy
    docker-compose.yml
  /docs
    README.md  ARCHITECTURE.md  DEMO.md
```

---

## 🗄️ DB Setup (target Postgres)

- Enable: `pg_stat_statements`, `auto_explain`, `hypopg` (for hypothetical index sim).
- Seed demo schema: `users, orders, order_items, events` (biggish), plus 8–12 intentional anti-patterns.
- Roles: `profiler_ro` (read), `profiler_sb` (sandbox schema operations).

---

## 🧠 Heuristics (fast + explainable)

1. **Missing Index**: selective predicates + seq scan on large table → `CREATE INDEX …(columns by selectivity)`
2. **JOIN w/o Composite Index**: equi-join on (a,b) w/o covering index → composite index recommendation
3. **Correlated Subquery** → `JOIN`/`EXISTS` rewrite suggestion
4. **Redundant/Covered Index**: `(a,b)` exists; unused `(a)` → drop hint
5. **Cardinality Skew**: |act−est|/est > K → `ANALYZE`, `ALTER TABLE … SET STATISTICS`, or expression index

**ML light**: TF-IDF + **K-Means** for query families; **MAD/IQR** anomalies per family.

---

## 📑 Profiler Meta Schema (SQLite or Postgres)

```sql
queries(id, fingerprint, raw_sql, norm_sql, first_seen, last_seen)
metrics(query_id, mean_ms, calls, rows, total_ms, captured_at)
plans(query_id, plan_json, had_seq_scan, est_rows, act_rows, buffers, captured_at)
schema_tables(table_name, rows_est, bytes)
schema_indexes(index_name, table_name, cols, unique, used, covers)
recommendations(id, query_id, type, ddl, rationale, confidence, created_at)
simulations(id, query_id, rec_id, before_ms, after_ms, before_plan, after_plan, improvement_pct, ran_at)
```

---

## 🌐 API (Fiber)

- `GET  /bottlenecks?limit=10` → list (reason, evidence, DDL)
- `GET  /queries/:id` → sql, metrics, plan facts, cluster, anomaly
- `GET  /recommendations?query_id=…`
- `POST /simulate` `{query_id, rec_id, mode:"hypopg|real"}` → % improvement + plan diff
- `POST /chat` `{question, query_id?}` → grounded answer (no hallucinations; template on your own facts)

**UI**: server-rendered pages (Fiber + HTMX): Dashboard, Query Detail (with **Run Simulation** button).

---

## 🧰 CLI (Cobra)

- `profiler init` (enable extensions, seed demo)
- `profiler scan --top 100 --min-mean-ms 5`
- `profiler bottlenecks --top 10`
- `profiler simulate --query <id> --rec <id> --mode hypopg`

---

# 📆 3-Day Plan (Asia/Kolkata)

### Day 1 — Core Ingest → Rules → API/CLI (0–24h)

**0–2h**

- Repo scaffold; wire **Fiber** + **Cobra**; env/config; `pgx` pool.
- `docker-compose`: Postgres with extensions.

**2–6h**

- Seed demo schema + synthetic workload (anti-patterns).
- Implement **ingest**: pull from `pg_stat_statements`, enrich with table/index stats.

**6–10h**

- **parse**: normalize SQL (strip literals), fingerprint; optional AST via `pg_query_go`.
- Store to meta DB; first dashboards (server-render tables).

**10–16h**

- **rules v1**: missing index, composite join index, correlated subquery, redundant index.
- **recommend**: DDL + rationale (plain English) + confidence.

**16–20h**

- **API**: `/bottlenecks`, `/queries/:id`, `/recommendations`.
- **CLI**: `scan`, `bottlenecks`.

**20–24h**

- Smoke demo: list top 10 bottlenecks with DDL.
  **Deliver**: Ingest → Rules → REST + CLI ✅

---

### Day 2 — ML Light → Simulator → UI Polish (24–48h)

**24–30h**

- **features**: TF-IDF vectors; **K-Means**; label clusters; anomaly via **MAD/IQR**.
- Surface cluster & anomaly in API/UI.

**30–40h**

- **simulate** (killer feature):

  - Baseline `EXPLAIN (ANALYZE, BUFFERS)`
  - `hypopg_create_index('CREATE INDEX …')`
  - Re-EXPLAIN; compute `%Δ` & node diffs; cleanup.

**40–44h**

- **UI polish**: before/after cards, badges (Seq→Index), confidence, risk notes.

**44–48h**

- Docs v1 (README quick-start, DEMO script).
  **Deliver**: Clustering + anomalies + **Impact Simulator** + polished UI ✅

---

### Day 3 — Grounded Q\&A → Hardening → Demo (48–72h)

**48–56h**

- **/chat** grounded answers: pull facts (metrics, plan, rec) → template response; (LLM optional).
- Q’s: “Why is Query X slow?”, “What index to add?”, “Show impact”.

**56–66h**

- Hardening: rate-limit EXPLAIN, add timeouts; cache schema/stat calls 2–5 min;
- Add `activity/locks` panel (basic `pg_locks` join) with “contention suspected” tag.

**66–72h**

- **Demo script** (`make demo`): seed → scan → show bottleneck → simulate (−XX%) → chat why.
- Final screenshots for README.

**Deliver**: Grounded chat + ops sanity + smooth demo ✅

---

## 🎯 Acceptance & Perf Targets

- **Scan 100 queries** ≤ **2s** (excluding initial cold cache)
- **Rules precision**: ≥ 80% of top recs give **>30%** sim speedup
- **Simulator**: hypopg round-trip ≤ **1.5s** for single query
- **UI**: first content paint ≤ **1s**, plan diff within **2s**

---

## 🧪 Minimal Test Matrix

- Missing index (single & composite) → improvement ≥ 70% on seeded cases
- Correlated subquery → JOIN rewrite sample shown
- Redundant index flagged correctly
- Anomaly spike detected when mean_ms ×2 vs baseline
- Simulator works in `hypopg` and `real` (sandbox schema)

---

## 🛡️ Risk Trims (use if behind)

- Skip AST; rely on plan + regex for correlated subquery (Day 1).
- Defer anomalies; keep only K-Means OR just rules.
- Keep UI server-rendered (no SPA build chain).

---

## ⚙️ Makefile Targets (speed)

```
make up          # docker compose up -d
make seed        # create schema + demo data
make scan        # cobra scan
make demo        # seed->scan->open browser
make test        # unit tests for rules/simulate
```

---

## 🔌 Optional (Only if ahead): MCP Sidecar

- Slot MCP calls for `explain_analyze`, `pg_stat_*`, `simulate_index` to stream results & cache centrally.
- Not required to win; nice bonus if time permits.
