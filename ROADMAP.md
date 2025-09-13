# OptiDB — Build-Execution Roadmap (Go + Fiber/Cobra, PostgreSQL)

## Key Principles

- **Postgres-first**: single DB focus for depth and reliability.
- **Explainability**: every finding has evidence → plain-English “why” → actionable DDL/rewrite.
- **Fast wins first**: heuristics before ML; **hypopg** simulation before “real” indexes.
- **Server-rendered UI**: Fiber + HTMX for rapid iteration and low cognitive overhead.
- **Two-lane team flow**: Person A (Data/Rules/DB), Person B (API/UI/Orchestration & ML-light).

---

## Global Interfaces (no code; define contracts)

### 1) Internal Data Contracts (stored in meta store)

- **Query**: `id`, `fingerprint`, `raw_sql`, `norm_sql`, `first_seen`, `last_seen`.
- **Metric**: `query_id`, `mean_ms`, `calls`, `rows`, `total_ms`, `captured_at`.
- **Plan**: `query_id`, `plan_json` (opaque), `had_seq_scan` (bool), `est_rows`, `act_rows`, `buffers_present` (bool), `captured_at`.
- **SchemaTable**: `table_name`, `rows_est`, `bytes`.
- **SchemaIndex**: `index_name`, `table_name`, `columns[]`, `unique` (bool), `scan_count`, `covers_columns[]`.
- **Recommendation**: `id`, `query_id`, `type` (missing_index|composite_index|rewrite|redundant_index|stats), `ddl`, `rationale`, `confidence`, `risk`.
- **Simulation**: `id`, `query_id`, `rec_id`, `before_ms`, `after_ms`, `improvement_pct`, `before_plan`, `after_plan`, `ran_at`.

### 2) External API Contracts (HTTP)

- `GET /bottlenecks?limit=N` → list of bottlenecks with: `query_id`, `fingerprint`, `mean_ms`, `calls`, `reason`, `ddl`, `confidence`.
- `GET /queries/:id` → full query details: metrics, plan facts (booleans/labels), family label, anomaly tags, attached recommendations.
- `POST /simulate` `{ query_id, rec_id, mode: "hypopg" | "real" }` → `before_ms`, `after_ms`, `improvement_pct`, plan-change labels.
- `POST /chat` `{ question, query_id? }` → grounded, templated answer that cites your stored facts (no LLM required to pass).

---

## What “Done” Looks Like Per Module

### Ingest (Postgres → Meta)

**Inputs:** `pg_stat_statements`, `pg_class`, `pg_index`, optional `pg_stats`, EXPLAIN JSON.
**Outputs:** Populated Query/Metric/Plan/Schema\* rows.
**Checks:**

- Top-N by total time captured.
- For each captured query: fingerprint assigned, normalized SQL stored, latest plan facts derived.
- Statement timeout respected globally.

### Parse & Features

**Inputs:** `raw_sql`, `plan_json`.
**Outputs:** `norm_sql`, `fingerprint`, token stream for TF-IDF, flags derived from plan (seq vs index scan, join types, est vs act skew).
**Checks:**

- Normalization stable across literal variations.
- Fingerprint collision rate negligible on demo set.
- Feature vectors exist for >90% of captured queries.

### Rules (Heuristics v1)

**Inputs:** latest Plan facts, Schema, Metrics.
**Outputs:** `Recommendation` rows with rationale and confidence.
**Checks:**

- Each surfaced bottleneck includes **specific** target (table.column order for index).
- Confidence scoring consistent: Missing index > Composite > Stats > Rewrite > Redundant (with usage evidence).

### ML-Light (Families & Anomalies)

**Inputs:** features from normalized SQL and metrics per family.
**Outputs:** family label, anomaly tags.
**Checks:**

- Family labels are human-meaningful (e.g., `orders_join_2col`, `users_like_email`).
- Anomaly rule catches known seeded spike (>2× median).

### Simulator

**Inputs:** query + chosen recommendation (index DDL).
**Process:** baseline EXPLAIN → hypo index → re-EXPLAIN → diff & clean.
**Outputs:** improvement %, plan-node class change flags.
**Checks:**

- Hypo state always cleaned.
- Safety rails: timeouts, concurrency cap, error surfacing without crashing.

### UI/CLI

**Inputs:** API responses.
**Outputs:** fast dashboard & detail; CLI parity for judges.
**Checks:**

- Dashboard paints under target; simulate result displays in <2s with clear %Δ and node change.

---

## Step-By-Step Execution (No code, just actions & acceptance)

### Day 1 (0–24h) — From data to visible recs

#### 5–9h: Ingest & API foundation

### **Abhi**

- Define **scan query** parameters: `top`, `min_mean_ms`, exclude boilerplate statements.
- Decide EXPLAIN sampling policy: _baseline for top-K only_ (K configurable; default 20).
- Derive plan facts to booleans/labels: `had_seq_scan`, `join_types[]`, `plan_vs_actual_ratio`, `filters[]`.
- Persist first pass of Query/Metric/Plan and Schema\*.

**Acceptance:** After one scan, meta has ≥20 queries with metrics and plan facts.

### **Dev**

- Expose `GET /bottlenecks` and `GET /queries/:id` (payloads as above).
- Render a **top list**: query preview, mean_ms, calls, “Reason (pending)” placeholder, “View”.

**Acceptance:** Opening `/bottlenecks` lists top queries from the seeded dataset.

---

#### 9–14h: Parse, fingerprint, and UI integration

**Abhi**

- Lock normalization rules (comments removal, literals → placeholders, whitespace fold, lowercase).
- Lock fingerprint method (hash selected; document rationale).
- Propose tokenization scheme for TF-IDF (tables, verbs, operators).

**Acceptance:** Same logical query with different literals maps to same fingerprint; verified on ≥3 seeded variants.

**Dev**

- In detail view, show **Plan Fact chips** (Seq vs Index, Est vs Act, Join types).
- Wire “reasoning area” ready for rules output.

**Acceptance:** Detail page is informative even before rules — chips display correctly.

---

#### 14–20h: Heuristics v1 (the brain)

**Abhi**

- **Missing index** logic: thresholds (table size, selectivity proxy), column ordering policy, rationale template, confidence scheme.
- **Composite join index**: detect join predicate cols; declare covering-index rule and rationale.
- **Correlated subquery**: regex/plan heuristic acceptable if AST deferred; rationale template includes rewrite sketch.
- **Redundant index**: coverage check plus low usage proof; conservative risk note.
- **Cardinality mismatch**: skew threshold K (start at 3–5); suggest ANALYZE/statistics bump and when to use expression index.

**Acceptance:** For each seeded slow pattern, at least one **specific** recommendation is produced with DDL + rationale + confidence.

**Dev**

- Surface recommendations inline on both list and detail.
- Include “copy DDL” affordance and **risk** badge.

**Acceptance:** `/bottlenecks` now shows **Reason, DDL, Confidence**; clicking opens detail with evidence.

---

#### 20–24h: Validation & tuning

- Re-run scan; verify: top **3–5** queries have **clear, correct** recs.
- Measure scan latency on warm cache (<2s target).
- Adjust thresholds to avoid noisy/low-value recs.
- Confirm CLI path: `init → scan → bottlenecks` is stable.

**Day-1 Done Definition**

- End-to-end pipeline yields actionable recs viewable in UI/CLI, with evidence chips.

---

### Day 2 (24–48h) — Simulator & ML-light polish

#### 24–30h: Simulator baseline

**Abhi**

- Define baseline EXPLAIN capture contract (timeout, buffers, annotations).
- Define improvement computation policy: min measurable delta; floor values to avoid misreporting tiny gains.
- Define plan change labels: “Seq → Index”, “Nested Loop → Hash Join”, etc.

**Acceptance:** A single query can be baseline-measured, and metrics are persisted.

**Dev**

- Add “Run Simulation” control and placeholder result area in detail view.

**Acceptance:** UI reacts to the action and shows “Running… / Completed” states.

---

#### 30–38h: Hypothetical index path

**Abhi**

- Agree on **hypopg** usage rules, cleanup guarantees, and supported DDL patterns (single/compound columns).
- Define **failure handling**: if hypopg cannot emulate (rare), show message; never leave dirty state.
- Define **real-mode** policy (Day-2 end or Day-3): only on **sandbox** schema, never on live tables.

**Acceptance:** On known seeded case, simulation reports large % improvement and proper node change labels.

**Dev**

- Add **before/after** panel: baseline vs simulated ms, %Δ, node-change badges, notes on write overhead risk.

**Acceptance:** Result is obvious and demo-ready (numbers & badges).

---

#### 38–44h: Families & anomalies

**Abhi + Dev (split)**

- Families: choose k (8–12) and label policy (dominant tokens).
- Anomalies: MAD/IQR per family; consistent thresholds.
- Expose family label & anomaly tags in detail + small chips in list.

**Acceptance:** At least two distinct families visible in demo; one anomaly case shows tag.

---

#### 44–48h: Simulator hardening

- Concurrency cap (policy only; impl later): max in-flight explains.
- Rate-limit per minute to protect DB.
- Error presentation standardized: “Why this failed, what to try”.

**Acceptance:** Fuzz test with 3 parallel simulations doesn’t degrade the DB or UX.

---

### Day 3 (48–72h) — Hardening, Q\&A, demo packaging

#### 48–54h: Operational hardening

**Abhi**

- TTL cache policy for schema/stats (2–5 min).
- `pg_locks` summary inclusion decision & “contention suspected” tag rules.
- Global timeouts and default thresholds documented & set from env.

**Acceptance:** Under load, scan + simulate remain responsive; lock contention badge appears on synthetic block.

---

#### 48–54h (parallel): Grounded Q\&A (no LLM required)

**Dev**

- Define **answer templates** that **quote your own data** only:

  - “Why is query X slow?” → plan facts + evidence + top recommendation (+ expected %Δ if simulation exists).
  - “What index should I add?” → bubble the highest-confidence index DDL.
  - “Show impact” → trigger simulation and report.

**Acceptance:** Given a query id, `/chat` returns a cohesive, accurate, non-hallucinated answer with citations (“based on plan captured at T, mean_ms Y, had_seq_scan=true”).

---

#### 54–66h: Docs & demo flow

- **Make targets** finalized (`up/init/seed/scan/demo/test`).
- README “90-sec Quick Start” + animated screenshots.
- DEMO script with fixed order: **Scan → Top Bottleneck → Simulate → Q\&A**.
- ARCHITECTURE doc: one diagram (data flow), one list (rules), one block (simulator).

**Acceptance:** A fresh machine can run the demo without tribal knowledge.

---

#### 66–72h: Final test & polish

- Re-seed; run complete dry-run.
- Validate acceptance targets (scan latency, simulator responsiveness, clarity of recs).
- Tighten any thresholds, clean logs, ensure dignified failure messages.

**Acceptance:** Team can execute demo in ≤90 seconds, repeatably.

---

## Acceptance Targets (unchanged, but now tied to checks)

- **Scan 100 queries** ≤ **2s** (warm cache).
- **≥80%** of top recs show **>30%** simulated speedup on seeded data.
- **Simulator** (hypopg) ≤ **1.5s** per run on demo data; always cleans up.
- **UI** FCP ≤ **1s**; plan diff appears ≤ **2s**.
- **Q\&A** always grounded, with explicit references to captured facts (no external claims).

---

## Risk Controls & Pre-approved Trims

- **Skip AST** if slipping: use plan & regex for correlated subqueries.
- **Families only** if time-pressed: defer anomalies to Day-3.
- **UI server-rendered only**: no SPA/React.
- **“Real” mode** only on sandbox schema; optional for demo.

---

## Division of Responsibility (locked)

- **Person A (Abhi):** Ingest → Parse/Features (facts) → Rules → Simulator → Ops hardening.
- **Person B (Dev):** API/DTOs → UI/HTMX → CLI → Families/Anomalies surface → Q\&A templates → Docs/Demo.

---

## Immediate Next (from your status)

- **Abhi:** Finish ingest joins (pg_stat_statements + pg_class/pg_index), persist first plan facts, lock normalization/fingerprint.
- **Dev:** Finish `/bottlenecks` & `/queries/:id` payloads + table/Detail pages; integrate “Reason / DDL / Confidence” placeholders so rules can slot in.

This keeps everything **technical, sequenced, and testable** without code blobs—so you can sprint straight into implementation with zero ambiguity.
