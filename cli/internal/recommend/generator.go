package recommend

import (
	"fmt"
	"strings"
	"time"

	"cli/internal/store"
)

type RecommendationGenerator struct {
	templates map[string]RecommendationTemplate
}

type RecommendationTemplate struct {
	DDLTemplate       string
	RationaleTemplate string
	ImpactTemplate    string
	DefaultConfidence float64
	DefaultRisk       string
}

func NewRecommendationGenerator() *RecommendationGenerator {
	templates := map[string]RecommendationTemplate{
		"missing_index": {
			DDLTemplate:       "CREATE INDEX idx_%s_%s ON %s (%s);",
			RationaleTemplate: "Table '%s' with %d rows performs sequential scans on column '%s'. Adding an index will significantly improve query performance.",
			ImpactTemplate:    "Expected 50-90%% performance improvement for queries filtering by %s",
			DefaultConfidence: 0.85,
			DefaultRisk:       "low",
		},
		"composite_index": {
			DDLTemplate:       "CREATE INDEX idx_%s_%s ON %s (%s);",
			RationaleTemplate: "Multiple column filters on table '%s' would benefit from a composite index covering columns (%s).",
			ImpactTemplate:    "Expected 40-80%% improvement for multi-column WHERE clauses",
			DefaultConfidence: 0.75,
			DefaultRisk:       "low",
		},
		"correlated_subquery": {
			RationaleTemplate: "Correlated subquery executes once per outer row (%d calls). Consider rewriting as JOIN or EXISTS for better performance.",
			ImpactTemplate:    "Expected 30-70%% performance improvement by eliminating N+1 query pattern",
			DefaultConfidence: 0.70,
			DefaultRisk:       "medium",
		},
		"join_index": {
			DDLTemplate:       "CREATE INDEX idx_%s_%s ON %s (%s);",
			RationaleTemplate: "JOIN operation on table '%s' lacks index on column '%s', causing nested loop joins instead of more efficient hash/merge joins.",
			ImpactTemplate:    "Expected 40-80%% improvement in join performance",
			DefaultConfidence: 0.80,
			DefaultRisk:       "low",
		},
		"redundant_index": {
			DDLTemplate:       "DROP INDEX %s;",
			RationaleTemplate: "Index '%s' on table '%s' is redundant with existing index '%s' and consumes %s of storage.",
			ImpactTemplate:    "Reclaim %s storage and reduce maintenance overhead",
			DefaultConfidence: 0.90,
			DefaultRisk:       "low",
		},
	}

	return &RecommendationGenerator{templates: templates}
}

func (rg *RecommendationGenerator) GenerateIndexRecommendation(tableName, columnName string, rowCount int64, queryCount int64) store.Recommendation {
	template := rg.templates["missing_index"]

	return store.Recommendation{
		Type:           "missing_index",
		DDL:            fmt.Sprintf(template.DDLTemplate, tableName, columnName, tableName, columnName),
		Rationale:      fmt.Sprintf(template.RationaleTemplate, tableName, rowCount, columnName),
		Confidence:     rg.calculateConfidence("missing_index", rowCount, queryCount),
		ImpactEstimate: fmt.Sprintf(template.ImpactTemplate, columnName),
		RiskLevel:      template.DefaultRisk,
		CreatedAt:      time.Now(),
	}
}

func (rg *RecommendationGenerator) GenerateCompositeIndexRecommendation(tableName string, columns []string, rowCount int64) store.Recommendation {
	template := rg.templates["composite_index"]
	columnList := strings.Join(columns, ", ")
	indexName := fmt.Sprintf("%s_%s", tableName, strings.Join(columns, "_"))

	return store.Recommendation{
		Type:           "composite_index",
		DDL:            fmt.Sprintf(template.DDLTemplate, tableName, indexName, tableName, columnList),
		Rationale:      fmt.Sprintf(template.RationaleTemplate, tableName, columnList),
		Confidence:     template.DefaultConfidence,
		ImpactEstimate: template.ImpactTemplate,
		RiskLevel:      template.DefaultRisk,
		CreatedAt:      time.Now(),
	}
}

func (rg *RecommendationGenerator) GenerateCorrelatedSubqueryRecommendation(queryStats store.QueryStats) store.Recommendation {
	template := rg.templates["correlated_subquery"]

	// Generate a simple rewrite suggestion
	rewriteSQL := rg.generateSubqueryRewrite(queryStats.Query)

	return store.Recommendation{
		Type:           "correlated_subquery",
		RewriteSQL:     rewriteSQL,
		Rationale:      fmt.Sprintf(template.RationaleTemplate, queryStats.Calls),
		Confidence:     template.DefaultConfidence,
		ImpactEstimate: template.ImpactTemplate,
		RiskLevel:      template.DefaultRisk,
		CreatedAt:      time.Now(),
	}
}

func (rg *RecommendationGenerator) GenerateJoinIndexRecommendation(tableName, columnName string, avgJoinTime float64) store.Recommendation {
	template := rg.templates["join_index"]

	return store.Recommendation{
		Type:           "join_index",
		DDL:            fmt.Sprintf(template.DDLTemplate, tableName, columnName, tableName, columnName),
		Rationale:      fmt.Sprintf(template.RationaleTemplate, tableName, columnName),
		Confidence:     template.DefaultConfidence,
		ImpactEstimate: template.ImpactTemplate,
		RiskLevel:      template.DefaultRisk,
		CreatedAt:      time.Now(),
	}
}

func (rg *RecommendationGenerator) GenerateRedundantIndexRecommendation(redundantIndex, existingIndex, tableName string, sizeBytes int64) store.Recommendation {
	template := rg.templates["redundant_index"]
	sizeStr := formatBytes(sizeBytes)

	return store.Recommendation{
		Type:           "redundant_index",
		DDL:            fmt.Sprintf(template.DDLTemplate, redundantIndex),
		Rationale:      fmt.Sprintf(template.RationaleTemplate, redundantIndex, tableName, existingIndex, sizeStr),
		Confidence:     template.DefaultConfidence,
		ImpactEstimate: fmt.Sprintf(template.ImpactTemplate, sizeStr),
		RiskLevel:      template.DefaultRisk,
		CreatedAt:      time.Now(),
	}
}

func (rg *RecommendationGenerator) calculateConfidence(recType string, rowCount, queryCount int64) float64 {
	baseConfidence := rg.templates[recType].DefaultConfidence

	// Adjust confidence based on table size and query frequency
	if rowCount > 10000 && queryCount > 10 {
		return baseConfidence + 0.1
	} else if rowCount < 1000 || queryCount < 5 {
		return baseConfidence - 0.2
	}

	return baseConfidence
}

func (rg *RecommendationGenerator) generateSubqueryRewrite(originalQuery string) string {
	// Simple heuristic rewrite suggestions
	if strings.Contains(strings.ToUpper(originalQuery), "EXISTS") {
		return "-- Consider using JOIN instead of EXISTS subquery\n-- Example: SELECT ... FROM table1 t1 JOIN table2 t2 ON t1.id = t2.foreign_id"
	}

	if strings.Contains(strings.ToUpper(originalQuery), "SELECT") && strings.Contains(originalQuery, "(") {
		return "-- Consider rewriting correlated subquery as JOIN\n-- Example: Replace (SELECT ... WHERE outer.id = inner.id) with proper JOIN"
	}

	return "-- Consider rewriting subquery as JOIN or window function for better performance"
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
