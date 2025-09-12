package ingest

import (
	"database/sql"
	"fmt"
	"strings"

	"cli/internal/store"
)

type StatsCollector struct {
	db *sql.DB
}

func NewStatsCollector(db *sql.DB) *StatsCollector {
	return &StatsCollector{db: db}
}

func (sc *StatsCollector) GetQueryStats() ([]store.QueryStats, error) {
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
			return nil, fmt.Errorf("failed to scan query stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (sc *StatsCollector) GetTableInfo() ([]store.TableInfo, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			n_tup_ins + n_tup_upd + n_tup_del as row_count,
			pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
		FROM pg_stat_user_tables
		ORDER BY size_bytes DESC
	`

	rows, err := sc.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table info: %w", err)
	}
	defer rows.Close()

	var tables []store.TableInfo
	for rows.Next() {
		var t store.TableInfo
		err := rows.Scan(&t.SchemaName, &t.TableName, &t.RowCount, &t.SizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}
		tables = append(tables, t)
	}

	return tables, nil
}

func (sc *StatsCollector) GetIndexInfo() ([]store.IndexInfo, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			string_to_array(regexp_replace(indexdef, '.*\((.*)\).*', '\1'), ', ') as columns,
			indisunique,
			indisprimary,
			pg_relation_size(schemaname||'.'||indexname) as size_bytes,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes
		JOIN pg_index ON pg_stat_user_indexes.indexrelid = pg_index.indexrelid
		JOIN pg_indexes ON pg_stat_user_indexes.indexname = pg_indexes.indexname
		ORDER BY size_bytes DESC
	`

	rows, err := sc.db.Query(query)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan index info: %w", err)
		}

		// Parse column array
		colsStr = strings.Trim(colsStr, "{}")
		if colsStr != "" {
			idx.Columns = strings.Split(colsStr, ",")
			for i, col := range idx.Columns {
				idx.Columns[i] = strings.TrimSpace(col)
			}
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func (sc *StatsCollector) GetSlowQueries(minDurationMS float64) ([]store.QueryStats, error) {
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
			return nil, fmt.Errorf("failed to scan slow query: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}
