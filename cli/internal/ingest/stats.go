package ingest

import (
	"database/sql"
	"fmt"
	"strings"

	"cli/internal/logger"
	"cli/internal/store"
)

type StatsCollector struct {
	db *sql.DB
}

func NewStatsCollector(db *sql.DB) *StatsCollector {
	return &StatsCollector{db: db}
}

func (sc *StatsCollector) GetQueryStats() ([]store.QueryStats, error) {
	logger.LogInfo("Collecting query statistics from pg_stat_statements")

	query := `
		SELECT 
			query,
			calls,
			mean_exec_time,
			total_exec_time,
			rows,
			shared_blks_hit,
			shared_blks_read
		FROM pg_stat_statements 
		WHERE query NOT LIKE '%pg_stat_statements%'
		  AND calls > 1
		ORDER BY mean_exec_time DESC
		LIMIT 100
	`

	rows, err := sc.db.Query(query)
	if err != nil {
		logger.LogErrorf("Failed to query pg_stat_statements: %v", err)
		return nil, fmt.Errorf("failed to query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	var stats []store.QueryStats
	for rows.Next() {
		var s store.QueryStats
		err := rows.Scan(
			&s.Query,
			&s.Calls,
			&s.MeanExecTime,
			&s.TotalTime,
			&s.Rows,
			&s.SharedBlksHit,
			&s.SharedBlksRead,
		)
		if err != nil {
			logger.LogErrorf("Failed to scan query stats row: %v", err)
			return nil, fmt.Errorf("failed to scan query stats: %w", err)
		}
		stats = append(stats, s)
	}

	logger.LogInfof("Collected %d query statistics records", len(stats))
	return stats, nil
}

func (sc *StatsCollector) GetTableInfo() ([]store.TableInfo, error) {
	logger.LogInfo("Collecting table information from pg_stat_user_tables")

	query := `
		SELECT 
			schemaname,
			relname as tablename,
			n_tup_ins + n_tup_upd + n_tup_del as row_count,
			pg_total_relation_size(schemaname||'.'||relname) as size_bytes
		FROM pg_stat_user_tables
		ORDER BY size_bytes DESC
	`

	rows, err := sc.db.Query(query)
	if err != nil {
		logger.LogErrorf("Failed to query table info: %v", err)
		return nil, fmt.Errorf("failed to query table info: %w", err)
	}
	defer rows.Close()

	var tables []store.TableInfo
	for rows.Next() {
		var t store.TableInfo
		err := rows.Scan(&t.SchemaName, &t.TableName, &t.RowCount, &t.SizeBytes)
		if err != nil {
			logger.LogErrorf("Failed to scan table info row: %v", err)
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}
		tables = append(tables, t)
	}

	logger.LogInfof("Collected %d table info records", len(tables))
	return tables, nil
}

func (sc *StatsCollector) GetIndexInfo() ([]store.IndexInfo, error) {
	logger.LogInfo("Collecting index information from pg_stat_user_indexes")

	query := `
		SELECT 
			psi.schemaname,
			psi.relname as tablename,
			psi.indexrelname as indexname,
			COALESCE(
				array_to_string(
					array(
						SELECT pg_get_indexdef(psi.indexrelid, k + 1, true)
						FROM generate_subscripts(pi.indkey, 1) as k
						ORDER BY k
					), 
					','
				), 
				''
			) as columns,
			pi.indisunique,
			pi.indisprimary,
			pg_relation_size(psi.indexrelid) as size_bytes,
			psi.idx_scan,
			psi.idx_tup_read,
			psi.idx_tup_fetch
		FROM pg_stat_user_indexes psi
		JOIN pg_index pi ON psi.indexrelid = pi.indexrelid
		ORDER BY size_bytes DESC
	`

	rows, err := sc.db.Query(query)
	if err != nil {
		logger.LogErrorf("Failed to query index info: %v", err)
		return nil, fmt.Errorf("failed to query index info: %w", err)
	}
	defer rows.Close()

	var indexes []store.IndexInfo
	for rows.Next() {
		var idx store.IndexInfo
		var colsStr string
		err := rows.Scan(
			&idx.SchemaName,
			&idx.TableName,
			&idx.IndexName,
			&colsStr,
			&idx.IsUnique,
			&idx.IsPrimary,
			&idx.SizeBytes,
			&idx.IndexScans,
			&idx.TuplesRead,
			&idx.TuplesFetch,
		)
		if err != nil {
			logger.LogErrorf("Failed to scan index info row: %v", err)
			return nil, fmt.Errorf("failed to scan index info: %w", err)
		}

		// Parse column names
		if colsStr != "" {
			idx.Columns = strings.Split(colsStr, ",")
			for i, col := range idx.Columns {
				idx.Columns[i] = strings.TrimSpace(col)
			}
		}

		logger.LogDebugf("Found index: %s on %s.%s with columns [%s]",
			idx.IndexName, idx.SchemaName, idx.TableName, strings.Join(idx.Columns, ", "))
		indexes = append(indexes, idx)
	}

	logger.LogInfof("Collected %d index info records", len(indexes))
	return indexes, nil
}

func (sc *StatsCollector) GetSlowQueries(minDurationMS float64) ([]store.QueryStats, error) {
	logger.LogInfof("Collecting slow queries with min duration: %.2fms", minDurationMS)

	query := `
		SELECT 
			query,
			calls,
			mean_exec_time,
			total_exec_time,
			rows,
			shared_blks_hit,
			shared_blks_read
		FROM pg_stat_statements 
		WHERE query NOT LIKE '%pg_stat_statements%'
		  AND mean_exec_time > $1
		  AND calls > 1
		ORDER BY mean_exec_time DESC
		LIMIT 50
	`

	rows, err := sc.db.Query(query, minDurationMS)
	if err != nil {
		logger.LogErrorf("Failed to query slow queries: %v", err)
		return nil, fmt.Errorf("failed to query slow queries: %w", err)
	}
	defer rows.Close()

	var stats []store.QueryStats
	for rows.Next() {
		var s store.QueryStats
		err := rows.Scan(
			&s.Query,
			&s.Calls,
			&s.MeanExecTime,
			&s.TotalTime,
			&s.Rows,
			&s.SharedBlksHit,
			&s.SharedBlksRead,
		)
		if err != nil {
			logger.LogErrorf("Failed to scan slow query row: %v", err)
			return nil, fmt.Errorf("failed to scan slow query: %w", err)
		}
		logger.LogDebugf("Found slow query: calls=%d, mean_time=%.2fms, query=%s",
			s.Calls, s.MeanExecTime, s.Query[:min(50, len(s.Query))])
		stats = append(stats, s)
	}

	logger.LogInfof("Collected %d slow queries", len(stats))
	return stats, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
