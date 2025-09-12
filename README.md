# OptiDB

- **DB**: PostgreSQL only (first-class).
- **Core**: ingest stats + plans, detect bottlenecks (rules + light ML), generate **actionable DDL/rewrite**, plain-English “why”.
- **Stretch**: **Impact Simulator** (hypopg + EXPLAIN ANALYZE), **grounded Q\&A** (templated, non-hallucinated).
- **UI**: server-rendered (Fiber + HTMX) for speed; no SPA yak-shaving.
- **CLI**: Cobra for headless eval and demo scripts.

---

# 🏗️ Repo Layout (lean)

```
/db-profiler
  /cmd
    /api            # Fiber main()
    /cli            # Cobra: init/scan/bottlenecks/simulate
  /internal
    /config         # env, flags
    /db             # pgx, migrations, roles, ext enable
    /ingest         # pg_stat_statements, EXPLAIN(ANALYZE, BUFFERS)
    /parse          # normalize + fingerprint (+ optional pg_query_go)
    /features       # TF-IDF, rollups
    /rules          # heuristics (index, joins, correlated, redundant, cardinality)
    /ml             # kmeans, MAD/IQR anomalies (gonum)
    /recommend      # DDL and rewrite synthesis + rationale strings
    /simulate       # hypopg compare: pre/post explain, %Δ + plan diff
    /store          # SQLite or pg meta store (DAOs)
    /http           # handlers, DTOs, templates(HTMX)
  /deploy           # docker-compose, seed.sql, Makefile
  /docs             # README, ARCHITECTURE, DEMO
```

---

# 🗄️ Meta Schema (SQLite or Postgres)

```sql
queries(id, fingerprint, raw_sql, norm_sql, first_seen, last_seen)
metrics(query_id, mean_ms, calls, rows, total_ms, captured_at)
plans(query_id, plan_json, had_seq_scan, est_rows, act_rows, buffers, captured_at)
schema_tables(table_name, rows_est, bytes)
schema_indexes(index_name, table_name, cols, unique, used, covers)
recommendations(id, query_id, type, ddl, rationale, confidence, created_at)
simulations(id, query_id, rec_id, before_ms, after_ms, improvement_pct, before_plan, after_plan, ran_at)
```

---

# 🧠 Detection & Recs (explainable, fast)

- **Missing Index**: selective WHERE + seq scan on big table → `CREATE INDEX …` (column order by selectivity/usage).
- **JOIN w/o Composite Index**: equi-join on (a,b) lacking covering index → suggest composite index.
- **Correlated Subquery**: detect via AST/patterns → advise `JOIN`/`EXISTS` rewrite (show skeleton).
- **Redundant/Covered Index**: (a,b) exists & (a) unused → drop hint (flag “validate in staging”).
- **Cardinality Mismatch**: |act−est|/max(est,1) > K → `ANALYZE`, raise stats target, or expression index.

**ML-light**: TF-IDF on normalized SQL + **K-Means** for “query families”; per-family **MAD/IQR** anomaly tags.

---

# 🌐 API (Fiber) + CLI (Cobra)

**Endpoints**

- `GET  /bottlenecks?limit=10`
- `GET  /queries/:id` (sql, metrics, plan facts, family, anomalies)
- `GET  /recommendations?query_id=…`
- `POST /simulate` `{query_id, rec_id, mode:"hypopg|real"}` → %Δ + plan diff
- `POST /chat` `{question, query_id?}` → grounded (templates from your data; LLM optional)

**CLI**

- `profiler init` (enable extensions, create roles/meta, seed demo)
- `profiler scan --top 100 --min-mean-ms 5`
- `profiler bottlenecks --top 10`
- `profiler simulate --query <id> --rec <id> --mode hypopg`

---

# 📆 72-Hour Plan (IST) — **Two-Person Tag Team**

### Day 1 — Ingest → Rules → API/CLI → Minimal UI

**Goal**: end-to-end scan to surfaced recs (raw but working).

| Time   | Abhi (Data/Rules/DB)                                                                                                                                             | Dev (API/UI/CLI)                                                                         |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| 0–2h   | `docker-compose` Postgres 16; enable `pg_stat_statements`, `auto_explain`, `hypopg`; create roles (`profiler_ro`, `profiler_sb`).                                    | Scaffold Fiber + Cobra; env/config; pgx pool; basic health endpoint.                          |
| 2–5h   | Seed schema (`users/orders/order_items/events`) + intentional slow queries (seq scans, bad joins, correlated subqueries).                                            | CLI: `init`, `scan`, `bottlenecks`. Wire `scan` to call API.                                  |
| 5–9h   | `/ingest`: pull `pg_stat_statements`; join with `pg_class`, `pg_index`; persist to meta store.                                                                       | `/http`: `GET /bottlenecks`, `GET /queries/:id`; server-rendered dashboard (HTMX) with top N. |
| 9–14h  | `/parse`: normalize & fingerprint; optional AST via `pg_query_go` (skip if short on time).                                                                           | DTOs for bottlenecks, query detail; simple plan facts chips (Seq/Index, est vs act).          |
| 14–20h | `/rules v1`: missing index, composite join index, correlated subquery (regex or AST), redundant index, cardinality skew; `/recommend`: DDL + rationale + confidence. | Wire rules to UI + CLI output; table of recs with “Why / DDL / Risk” columns.                 |
| 20–24h | Smoke pass on seeded data; adjust thresholds.                                                                                                                        | CLI demo script `scan→bottlenecks`. Minimal README.                                           |

**EOD D1 Deliverable**: Scan → detect → recommend visible in UI/CLI ✅

---

### Day 2 — Features/ML → **Simulator** → UI Polish

**Goal**: cluster/anomaly context + **impact simulator** WOW.

| Time   | Person A (Simulator & Tests)                                                                                                                            | Person B (Features/ML & UI)                                                                                        |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| 24–30h | `/simulate`: baseline `EXPLAIN (ANALYZE, BUFFERS)` capture.                                                                                             | `/features`: TF-IDF; `/ml`: **K-Means** families; label via table/verb bigrams.                                    |
| 30–38h | Add **hypopg**: `hypopg_create_index('CREATE INDEX …')` → re-EXPLAIN → compute `improvement_pct`; capture node diffs (Seq→Index). Cleanup hypopg state. | Per-family **MAD/IQR** anomalies; expose tags in `/queries/:id` + `/bottlenecks`.                                  |
| 38–44h | Guards: timeouts, concurrency caps, rollback on errors; unit tests for rules + simulate.                                                                | UI polish: Before/After cards with %Δ badge; plan snippet diff (node type change badges); confidence & risk notes. |
| 44–48h | Add optional `mode:"real"` on **sandbox schema** (not default).                                                                                         | Update CLI: `simulate` command; improve table formatting.                                                          |

**EOD D2 Deliverable**: **Impact Simulator** live + families/anomalies in UI ✅

---

### Day 3 — Grounded Q\&A → Hardening → Judge Demo

**Goal**: explain like a human, be robust, ship docs + scripted demo.

| Time   | Person A (Ops Hardening)                                                                                                    | Person B (Grounded Q\&A & Docs)                                                                                                                   |
| ------ | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 48–54h | Cache schema/stat calls (2–5 min); add `pg_locks` summary and “contention suspected” label; EXPLAIN timeouts + rate limits. | `/chat`: template-grounded answers: pulls query metrics, plan facts, and recommended DDL; returns cited explanation (no hallucinations required). |
| 54–66h | Finalize Make targets: `up/init/seed/scan/demo/test`; stabilize thresholds & defaults.                                      | Docs: README (90-sec Quick Start), ARCHITECTURE (diagram + flow), DEMO (script + screenshots/GIFs).                                               |
| 66–72h | Run full dry-run; capture screenshots; trim logs.                                                                           | Demo rehearsal: seed → scan → bottleneck → simulate → chat “Why is Query X slow?”.                                                                |

**EOD D3 Deliverable**: Grounded Q\&A + hardened ops + slick demo ✅

---

# 🎯 Acceptance Targets

- **Scan 100 queries** ≤ **2s** (warm cache).
- **Top recs precision**: ≥ **80%** show **>30%** simulated speedup.
- **Simulator** (hypopg): **≤1.5s** round-trip per query on demo data.
- **UI**: First content paint ≤ **1s**; plan diff visible ≤ **2s**.
- **Q\&A**: 100% grounded from stored facts (no external guessing).

---

# 🧪 Test Matrix (minimum)

- Missing index (single + composite) → ≥70% speedup on seeded cases.
- JOIN covering index suggestion appears only when absent.
- Correlated subquery flagged and rewrite sketch rendered.
- Redundant index flagged only when covered + unused.
- Anomaly triggers when mean_ms doubles vs baseline.
- Simulator cleans up hypopg state reliably; respects timeouts.

---

# 🛡️ If You Slip (pre-approved trims)

- Skip AST day-1; use plan + regex for correlated subquery; add AST later.
- Keep anomalies simple (MAD/IQR); defer change-point/seasonality.
- Server-rendered UI only; no React/D3; plain HTMX + badges.

---

# 🔌 Makefile (speed)

```
make up         # docker compose up -d
make init       # create roles, enable extensions, meta store
make seed       # demo schema + slow workloads
make scan       # ingest stats + plans
make demo       # seed -> scan -> open UI
make test       # rules + simulate
```

---

# 🧾 Demo Script (judge-proof)

1. `make demo` → Dashboard lists **Top Bottlenecks**.
2. Click one → **Why** (plain English) + **DDL**.
3. Hit **Run Simulation** → show **−XX%** latency; badge “Seq Scan → Index Scan”.
4. Ask **“Why is Query 12 slow?”** in Q\&A → grounded answer citing metrics & plan facts.

This is the **battle-ready, fuck-around-free** plan that merges your two drafts into a 72-hour execution path with clean parallelization for two people and a guaranteed “wow” moment.
