package rules

import (
	"fmt"
	"regexp"
	"strings"

	"cli/internal/ai"
	"cli/internal/logger"
	"cli/internal/store"
)

type RuleEngine struct {
	minTableSize     int64
	minSeqScanTime   float64
	minCalls         int64
	correlationRegex *regexp.Regexp
	aiClient         *ai.OpenAIClient
	useAI            bool
}

func NewRuleEngine() *RuleEngine {
	// initialize AI client
	aiClient, err := ai.NewOpenAIClient()
	useAI := err == nil

	if useAI {
		logger.LogInfo("AI-powered recommendations enabled")
	} else {
		logger.LogInfof("AI client initialization failed, using heuristic rules: %v", err)
	}

	return &RuleEngine{
		minTableSize:     1000, // Minimum table size to suggest indexes
		minSeqScanTime:   0.1,  // Minimum time (ms) to consider slow
		minCalls:         5,    // Minimum calls to consider for optimization
		correlationRegex: regexp.MustCompile(`(?i)SELECT.*\(.*SELECT.*WHERE.*=.*\w+\.`),
		aiClient:         aiClient,
		useAI:            useAI,
	}
}

func (re *RuleEngine) AnalyzeQuery(query store.QueryStats, tables []store.TableInfo, indexes []store.IndexInfo) []store.Recommendation {
	logger.LogDebugf("Analyzing query with %d calls, %.2fms avg time", query.Calls, query.MeanExecTime)

	var recommendations []store.Recommendation

	// Skip if not enough calls to be significant
	if query.Calls < re.minCalls {
		logger.LogDebugf("Skipping query with insufficient calls (%d < %d)", query.Calls, re.minCalls)
		return recommendations
	}

	// AI-powered recommendations if available
	if re.useAI && re.aiClient != nil {
		logger.LogInfo("Using AI-powered recommendation generation")
		aiRecs, err := re.aiClient.GenerateRecommendations(query, tables, indexes)
		if err != nil {
			logger.LogErrorf("AI recommendation failed, falling back to heuristics: %v", err)
		} else {
			logger.LogInfof("Generated %d AI-powered recommendations", len(aiRecs))
			return aiRecs
		}
	}

	// HARDCODE : heuristic rules

	logger.LogInfo("Using heuristic rule-based recommendations")

	// Extract table names from query
	tableNames := re.extractTableNames(query.Query)
	logger.LogDebugf("Extracted table names from query: %v", tableNames)

	// Check for missing indexes
	if rec := re.detectMissingIndex(query, tableNames, tables, indexes); rec != nil {
		logger.LogInfof("Detected missing index recommendation for query")
		recommendations = append(recommendations, *rec)
	}

	// Check for correlated subqueries
	if rec := re.detectCorrelatedSubquery(query); rec != nil {
		logger.LogInfof("Detected correlated subquery recommendation for query")
		recommendations = append(recommendations, *rec)
	}

	// Check for inefficient joins
	if rec := re.detectIneffientJoin(query, tableNames, indexes); rec != nil {
		logger.LogInfof("Detected inefficient join recommendation for query")
		recommendations = append(recommendations, *rec)
	}

	// Check for redundant indexes
	if rec := re.detectRedundantIndex(indexes, tableNames); rec != nil {
		logger.LogInfof("Detected redundant index recommendation")
		recommendations = append(recommendations, *rec)
	}

	// Check for cardinality issues
	if rec := re.detectCardinalityIssues(query, tables); rec != nil {
		logger.LogInfof("Detected cardinality issue recommendation")
		recommendations = append(recommendations, *rec)
	}

	logger.LogDebugf("Generated %d heuristic recommendations for query", len(recommendations))
	return recommendations
}

func (re *RuleEngine) detectMissingIndex(query store.QueryStats, tableNames []string, tables []store.TableInfo, indexes []store.IndexInfo) *store.Recommendation {
	// Only suggest indexes for slow queries on large tables
	if query.MeanExecTime < re.minSeqScanTime {
		return nil
	}

	queryUpper := strings.ToUpper(query.Query)

	// Look for WHERE clauses that might benefit from indexes
	wherePatterns := []struct {
		pattern string
		column  string
	}{
		{`WHERE\s+(\w+)\s*=`, ""},
		{`WHERE\s+(\w+)\s*IN`, ""},
		{`WHERE\s+(\w+)\s*>`, ""},
		{`WHERE\s+(\w+)\s*<`, ""},
		{`WHERE\s+(\w+)\s*LIKE`, ""},
	}

	for _, pattern := range wherePatterns {
		regex := regexp.MustCompile(pattern.pattern)
		matches := regex.FindStringSubmatch(queryUpper)
		if len(matches) > 1 {
			column := strings.ToLower(matches[1])

			// Check if this column already has an index
			hasIndex := false
			for _, idx := range indexes {
				for _, tableName := range tableNames {
					if idx.TableName == tableName && len(idx.Columns) > 0 {
						if strings.ToLower(idx.Columns[0]) == column {
							hasIndex = true
							break
						}
					}
				}
				if hasIndex {
					break
				}
			}

			if !hasIndex {
				// Find the table to suggest index for
				for _, tableName := range tableNames {
					for _, table := range tables {
						if table.TableName == tableName && table.RowCount > re.minTableSize {
							return &store.Recommendation{
								Type:           "missing_index",
								DDL:            fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s);", tableName, column, tableName, column),
								Rationale:      fmt.Sprintf("Query performs sequential scan on table '%s' filtering by column '%s'. An index would improve performance.", tableName, column),
								Confidence:     0.8,
								ImpactEstimate: fmt.Sprintf("Expected 50-90%% performance improvement for queries filtering by %s", column),
								RiskLevel:      "low",
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (re *RuleEngine) detectCorrelatedSubquery(query store.QueryStats) *store.Recommendation {
	if query.MeanExecTime < re.minSeqScanTime*2 { // Higher threshold for correlated subqueries
		return nil
	}

	if re.correlationRegex.MatchString(query.Query) {
		return &store.Recommendation{
			Type:           "correlated_subquery",
			RewriteSQL:     "-- Consider rewriting correlated subquery as JOIN or EXISTS clause",
			Rationale:      "Query contains a correlated subquery that executes once per outer row. Consider rewriting as a JOIN or EXISTS for better performance.",
			Confidence:     0.7,
			ImpactEstimate: "Expected 30-70% performance improvement by eliminating correlated subquery",
			RiskLevel:      "medium",
		}
	}

	return nil
}

func (re *RuleEngine) detectIneffientJoin(query store.QueryStats, tableNames []string, indexes []store.IndexInfo) *store.Recommendation {
	if query.MeanExecTime < re.minSeqScanTime {
		return nil
	}

	queryUpper := strings.ToUpper(query.Query)

	// Look for JOIN conditions
	joinPattern := regexp.MustCompile(`JOIN\s+(\w+)\s+\w*\s*ON\s+\w+\.(\w+)\s*=\s*\w+\.(\w+)`)
	matches := joinPattern.FindAllStringSubmatch(queryUpper, -1)

	for _, match := range matches {
		if len(match) > 3 {
			joinTable := strings.ToLower(match[1])
			leftCol := strings.ToLower(match[2])
			rightCol := strings.ToLower(match[3])

			// Check if join columns have indexes
			hasLeftIndex := re.hasIndexOnColumn(leftCol, tableNames, indexes)
			hasRightIndex := re.hasIndexOnColumn(rightCol, []string{joinTable}, indexes)

			if !hasLeftIndex || !hasRightIndex {
				missingCol := leftCol
				missingTable := tableNames[0] // Simplified
				if !hasRightIndex {
					missingCol = rightCol
					missingTable = joinTable
				}

				return &store.Recommendation{
					Type:           "join_index",
					DDL:            fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s);", missingTable, missingCol, missingTable, missingCol),
					Rationale:      fmt.Sprintf("JOIN operation lacks index on column '%s' in table '%s', causing slow nested loop joins.", missingCol, missingTable),
					Confidence:     0.75,
					ImpactEstimate: "Expected 40-80% improvement in join performance",
					RiskLevel:      "low",
				}
			}
		}
	}

	return nil
}

func (re *RuleEngine) hasIndexOnColumn(column string, tableNames []string, indexes []store.IndexInfo) bool {
	for _, idx := range indexes {
		for _, tableName := range tableNames {
			if idx.TableName == tableName && len(idx.Columns) > 0 {
				if strings.ToLower(idx.Columns[0]) == column {
					return true
				}
			}
		}
	}
	return false
}

func (re *RuleEngine) extractTableNames(query string) []string {
	var tables []string

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		regexp.MustCompile(`(?i)\bJOIN\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		regexp.MustCompile(`(?i)\bUPDATE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(query, -1)
		for _, match := range matches {
			if len(match) > 1 {
				table := strings.ToLower(match[1])
				if !contains(tables, table) {
					tables = append(tables, table)
				}
			}
		}
	}

	return tables
}

func (re *RuleEngine) detectRedundantIndex(indexes []store.IndexInfo, tableNames []string) *store.Recommendation {
	// Look for indexes that are prefixes of other indexes on the same table
	for i, idx1 := range indexes {
		if !contains(tableNames, idx1.TableName) {
			continue
		}

		for j, idx2 := range indexes {
			if i >= j || idx1.TableName != idx2.TableName {
				continue
			}

			// Check if idx1 is a prefix of idx2 and unused
			if len(idx1.Columns) < len(idx2.Columns) && idx1.IndexScans < 10 {
				isPrefix := true
				for k, col := range idx1.Columns {
					if k >= len(idx2.Columns) || strings.EqualFold(strings.ToLower(col),strings.ToLower(idx2.Columns[k])) {
						isPrefix = false
						break
					}
				}

				if isPrefix {
					return &store.Recommendation{
						Type:           "redundant_index",
						DDL:            fmt.Sprintf("DROP INDEX %s;", idx1.IndexName),
						Rationale:      fmt.Sprintf("Index '%s' on table '%s' is redundant with '%s' and has low usage (%d scans). The larger index covers the same queries.", idx1.IndexName, idx1.TableName, idx2.IndexName, idx1.IndexScans),
						Confidence:     0.85,
						ImpactEstimate: fmt.Sprintf("Reclaim %s storage and reduce maintenance overhead", formatBytes(idx1.SizeBytes)),
						RiskLevel:      "low",
					}
				}
			}
		}
	}

	return nil
}

func (re *RuleEngine) detectCardinalityIssues(query store.QueryStats, tables []store.TableInfo) *store.Recommendation {
	// Simple heuristic: if a query returns way more or fewer rows than expected based on table size
	if len(tables) == 0 || query.Rows == 0 {
		return nil
	}

	// Find the largest table involved (rough estimate)
	var maxTableSize int64
	var tableName string
	for _, table := range tables {
		if table.RowCount > maxTableSize {
			maxTableSize = table.RowCount
			tableName = table.TableName
		}
	}

	if maxTableSize == 0 {
		return nil
	}

	// If query returns a very small fraction of a large table, might need better statistics
	selectivity := float64(query.Rows) / float64(maxTableSize)
	if maxTableSize > 100000 && selectivity < 0.001 && query.MeanExecTime > 1.0 {
		return &store.Recommendation{
			Type:           "cardinality_issue",
			DDL:            fmt.Sprintf("ANALYZE %s; -- or ALTER TABLE %s ALTER COLUMN <selective_column> SET STATISTICS 1000;", tableName, tableName),
			Rationale:      fmt.Sprintf("Query has very low selectivity (%.4f%%) on large table '%s' but still slow. Consider updating table statistics or creating expression indexes.", selectivity*100, tableName),
			Confidence:     0.60,
			ImpactEstimate: "Expected 20-50% improvement with better statistics",
			RiskLevel:      "low",
		}
	}

	return nil
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
