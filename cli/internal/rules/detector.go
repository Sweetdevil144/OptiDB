package rules

import (
	"fmt"
	"regexp"
	"strings"

	"cli/internal/store"
)

type RuleEngine struct {
	minTableSize     int64
	minSeqScanTime   float64
	minCalls         int64
	correlationRegex *regexp.Regexp
}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		minTableSize:     1000, // Minimum table size to suggest indexes
		minSeqScanTime:   0.1,  // Minimum time (ms) to consider slow
		minCalls:         5,    // Minimum calls to consider for optimization
		correlationRegex: regexp.MustCompile(`(?i)SELECT.*\(.*SELECT.*WHERE.*=.*\w+\.`),
	}
}

func (re *RuleEngine) AnalyzeQuery(query store.QueryStats, tables []store.TableInfo, indexes []store.IndexInfo) []store.Recommendation {
	var recommendations []store.Recommendation

	// Skip if not enough calls to be significant
	if query.Calls < re.minCalls {
		return recommendations
	}

	// Extract table names from query
	tableNames := re.extractTableNames(query.Query)

	// Check for missing indexes
	if rec := re.detectMissingIndex(query, tableNames, tables, indexes); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check for correlated subqueries
	if rec := re.detectCorrelatedSubquery(query); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check for inefficient joins
	if rec := re.detectIneffientJoin(query, tableNames, indexes); rec != nil {
		recommendations = append(recommendations, *rec)
	}

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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
