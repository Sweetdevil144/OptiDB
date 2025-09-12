package store

import (
	"time"
)

type Query struct {
	ID          int64     `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	RawSQL      string    `json:"raw_sql"`
	NormSQL     string    `json:"norm_sql,omitempty"`
	QueryHash   int64     `json:"query_hash,omitempty"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

type QueryMetrics struct {
	ID              int64     `json:"id"`
	QueryID         int64     `json:"query_id"`
	MeanMS          float64   `json:"mean_ms"`
	Calls           int64     `json:"calls"`
	RowsReturned    int64     `json:"rows_returned,omitempty"`
	TotalMS         float64   `json:"total_ms"`
	SharedBlksHit   int64     `json:"shared_blks_hit,omitempty"`
	SharedBlksRead  int64     `json:"shared_blks_read,omitempty"`
	TempBlksRead    int64     `json:"temp_blks_read,omitempty"`
	TempBlksWritten int64     `json:"temp_blks_written,omitempty"`
	CapturedAt      time.Time `json:"captured_at"`
}

type QueryPlan struct {
	ID          int64     `json:"id"`
	QueryID     int64     `json:"query_id"`
	PlanJSON    string    `json:"plan_json"`
	HadSeqScan  bool      `json:"had_seq_scan"`
	EstRows     int64     `json:"est_rows,omitempty"`
	ActRows     int64     `json:"act_rows,omitempty"`
	BuffersHit  int64     `json:"buffers_hit,omitempty"`
	BuffersRead int64     `json:"buffers_read,omitempty"`
	CapturedAt  time.Time `json:"captured_at"`
}

type SchemaTable struct {
	ID           int64     `json:"id"`
	SchemaName   string    `json:"schema_name"`
	TableName    string    `json:"table_name"`
	RowsEst      int64     `json:"rows_est,omitempty"`
	Bytes        int64     `json:"bytes,omitempty"`
	LastAnalyzed time.Time `json:"last_analyzed,omitempty"`
	CapturedAt   time.Time `json:"captured_at"`
}

type SchemaIndex struct {
	ID            int64     `json:"id"`
	SchemaName    string    `json:"schema_name"`
	IndexName     string    `json:"index_name"`
	TableName     string    `json:"table_name"`
	Cols          []string  `json:"cols"`
	IsUnique      bool      `json:"is_unique"`
	IsPrimary     bool      `json:"is_primary"`
	SizeBytes     int64     `json:"size_bytes,omitempty"`
	Scans         int64     `json:"scans,omitempty"`
	TuplesRead    int64     `json:"tuples_read,omitempty"`
	TuplesFetched int64     `json:"tuples_fetched,omitempty"`
	CapturedAt    time.Time `json:"captured_at"`
}

type Recommendation struct {
	ID             int64     `json:"id"`
	QueryID        int64     `json:"query_id,omitempty"`
	Type           string    `json:"type"`
	DDL            string    `json:"ddl,omitempty"`
	RewriteSQL     string    `json:"rewrite_sql,omitempty"`
	Rationale      string    `json:"rationale"`
	Confidence     float64   `json:"confidence"`
	ImpactEstimate string    `json:"impact_estimate,omitempty"`
	RiskLevel      string    `json:"risk_level"`
	CreatedAt      time.Time `json:"created_at"`
}

type QueryStats struct {
	Query          string  `json:"query"`
	Calls          int64   `json:"calls"`
	MeanExecTime   float64 `json:"mean_exec_time"`
	TotalTime      float64 `json:"total_time"`
	Rows           int64   `json:"rows"`
	SharedBlksHit  int64   `json:"shared_blks_hit"`
	SharedBlksRead int64   `json:"shared_blks_read"`
}

type TableInfo struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`
	RowCount   int64  `json:"row_count"`
	SizeBytes  int64  `json:"size_bytes"`
}

type IndexInfo struct {
	SchemaName  string   `json:"schema_name"`
	TableName   string   `json:"table_name"`
	IndexName   string   `json:"index_name"`
	Columns     []string `json:"columns"`
	IsUnique    bool     `json:"is_unique"`
	IsPrimary   bool     `json:"is_primary"`
	SizeBytes   int64    `json:"size_bytes"`
	IndexScans  int64    `json:"index_scans"`
	TuplesRead  int64    `json:"tuples_read"`
	TuplesFetch int64    `json:"tuples_fetch"`
}
