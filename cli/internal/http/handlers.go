package http

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cli/internal/db"
	"cli/internal/ingest"
	"cli/internal/logger"
	"cli/internal/rules"
	"cli/internal/store"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	collector  *ingest.StatsCollector
	ruleEngine *rules.RuleEngine
}

func NewHandlers(database *db.Config) *Handlers {
	conn, err := db.ConnectAsProfiler()
	if err != nil {
		logger.LogErrorf("Failed to create database connection for HTTP handlers: %v", err)
		return nil
	}

	collector := ingest.NewStatsCollector(conn)
	ruleEngine := rules.NewRuleEngine()

	return &Handlers{
		collector:  collector,
		ruleEngine: ruleEngine,
	}
}

// BottleneckDTO represents a bottleneck with recommendations
type BottleneckDTO struct {
	QueryID         string              `json:"query_id"`
	Query           string              `json:"query"`
	Calls           int64               `json:"calls"`
	MeanExecTime    float64             `json:"mean_exec_time"`
	TotalTime       float64             `json:"total_time"`
	Rows            int64               `json:"rows"`
	Fingerprint     string              `json:"fingerprint"`
	Recommendations []RecommendationDTO `json:"recommendations"`
	PlanFacts       PlanFactsDTO        `json:"plan_facts"`
}

// RecommendationDTO represents a single recommendation
type RecommendationDTO struct {
	Type           string  `json:"type"`
	DDL            string  `json:"ddl,omitempty"`
	RewriteSQL     string  `json:"rewrite_sql,omitempty"`
	Rationale      string  `json:"rationale"`
	Confidence     float64 `json:"confidence"`
	ImpactEstimate string  `json:"impact_estimate,omitempty"`
	RiskLevel      string  `json:"risk_level"`
}

// PlanFactsDTO represents query execution plan facts
type PlanFactsDTO struct {
	HasSeqScan    bool    `json:"has_seq_scan"`
	HasIndexScan  bool    `json:"has_index_scan"`
	EstimatedRows int64   `json:"estimated_rows"`
	ActualRows    int64   `json:"actual_rows"`
	BuffersHit    int64   `json:"buffers_hit"`
	BuffersRead   int64   `json:"buffers_read"`
	Selectivity   float64 `json:"selectivity"`
}

// QueryDetailDTO represents detailed query information
type QueryDetailDTO struct {
	QueryID         string              `json:"query_id"`
	Query           string              `json:"query"`
	Fingerprint     string              `json:"fingerprint"`
	Stats           QueryStatsDTO       `json:"stats"`
	Recommendations []RecommendationDTO `json:"recommendations"`
	PlanFacts       PlanFactsDTO        `json:"plan_facts"`
	Tables          []string            `json:"tables"`
}

// QueryStatsDTO represents query performance statistics
type QueryStatsDTO struct {
	Calls          int64   `json:"calls"`
	MeanExecTime   float64 `json:"mean_exec_time"`
	TotalTime      float64 `json:"total_time"`
	Rows           int64   `json:"rows"`
	SharedBlksHit  int64   `json:"shared_blks_hit"`
	SharedBlksRead int64   `json:"shared_blks_read"`
}

// ScanResultDTO represents scan command results
type ScanResultDTO struct {
	QueryID         string  `json:"query_id"`
	Query           string  `json:"query"`
	Calls           int64   `json:"calls"`
	MeanExecTime    float64 `json:"mean_exec_time"`
	TotalTime       float64 `json:"total_time"`
	Recommendations int     `json:"recommendations"`
}

// SystemStatusDTO represents overall system status
type SystemStatusDTO struct {
	Database DatabaseStatusDTO `json:"database"`
	Tables   TableStatusDTO    `json:"tables"`
	Indexes  IndexStatusDTO    `json:"indexes"`
	AI       AIStatusDTO       `json:"ai"`
}

// DatabaseStatusDTO represents database connection status
type DatabaseStatusDTO struct {
	Connected       bool    `json:"connected"`
	TotalQueries    int     `json:"total_queries"`
	SlowQueries     int     `json:"slow_queries"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// TableStatusDTO represents table statistics
type TableStatusDTO struct {
	Count     int   `json:"count"`
	TotalRows int64 `json:"total_rows"`
}

// IndexStatusDTO represents index statistics
type IndexStatusDTO struct {
	Count     int   `json:"count"`
	TotalSize int64 `json:"total_size"`
}

// AIStatusDTO represents AI system status
type AIStatusDTO struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
}

// GetBottlenecks returns top N bottlenecks
func (h *Handlers) GetBottlenecks(c *fiber.Ctx) error {
	logger.LogInfo("HTTP: Getting bottlenecks")

	// Parse query parameters
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	minDurationStr := c.Query("min_duration", "0.1")
	minDuration, err := strconv.ParseFloat(minDurationStr, 64)
	if err != nil {
		minDuration = 0.1
	}

	// Get slow queries
	queryStats, err := h.collector.GetSlowQueries(minDuration)
	if err != nil {
		logger.LogErrorf("Failed to get slow queries: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve query statistics",
		})
	}

	// Get metadata
	tables, err := h.collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to get table info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve table information",
		})
	}

	indexes, err := h.collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to get index info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve index information",
		})
	}

	// Convert to DTOs
	var bottlenecks []BottleneckDTO
	for i, query := range queryStats {
		if i >= limit {
			break
		}

		// Generate recommendations
		recommendations := h.ruleEngine.AnalyzeQuery(query, tables, indexes)

		// Convert recommendations to DTOs
		var recDTOs []RecommendationDTO
		for _, rec := range recommendations {
			recDTOs = append(recDTOs, RecommendationDTO{
				Type:           rec.Type,
				DDL:            rec.DDL,
				RewriteSQL:     rec.RewriteSQL,
				Rationale:      rec.Rationale,
				Confidence:     rec.Confidence,
				ImpactEstimate: rec.ImpactEstimate,
				RiskLevel:      rec.RiskLevel,
			})
		}

		// Generate plan facts
		planFacts := h.generatePlanFacts(query, tables)

		// Generate fingerprint
		fingerprint := h.generateFingerprint(query.Query)

		bottleneck := BottleneckDTO{
			QueryID:         fingerprint[:12],
			Query:           query.Query,
			Calls:           query.Calls,
			MeanExecTime:    query.MeanExecTime,
			TotalTime:       query.TotalTime,
			Rows:            query.Rows,
			Fingerprint:     fingerprint,
			Recommendations: recDTOs,
			PlanFacts:       planFacts,
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	logger.LogInfof("HTTP: Returning %d bottlenecks", len(bottlenecks))
	// Check if this is an HTMX request (expecting HTML)
	if c.Get("HX-Request") == "true" {
		return c.SendString(h.renderBottlenecksHTML(bottlenecks, limit))
	}

	return c.JSON(fiber.Map{
		"bottlenecks": bottlenecks,
		"total":       len(bottlenecks),
		"limit":       limit,
	})
}

// GetQueryDetail returns detailed information about a specific query
func (h *Handlers) GetQueryDetail(c *fiber.Ctx) error {
	queryID := c.Params("id")
	logger.LogInfof("HTTP: Getting query detail for ID: %s", queryID)

	// Get all queries and find the one with matching fingerprint
	queryStats, err := h.collector.GetSlowQueries(0.0) // Get all queries
	if err != nil {
		logger.LogErrorf("Failed to get queries: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve queries",
		})
	}

	// Find query by fingerprint prefix
	var targetQuery *store.QueryStats
	for _, query := range queryStats {
		fingerprint := h.generateFingerprint(query.Query)
		if len(fingerprint) >= 12 && fingerprint[:12] == queryID {
			targetQuery = &query
			break
		}
	}

	if targetQuery == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Query not found",
		})
	}

	// Get metadata
	tables, err := h.collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to get table info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve table information",
		})
	}

	indexes, err := h.collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to get index info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve index information",
		})
	}

	// Generate recommendations
	recommendations := h.ruleEngine.AnalyzeQuery(*targetQuery, tables, indexes)

	// Convert recommendations to DTOs
	var recDTOs []RecommendationDTO
	for _, rec := range recommendations {
		recDTOs = append(recDTOs, RecommendationDTO{
			Type:           rec.Type,
			DDL:            rec.DDL,
			RewriteSQL:     rec.RewriteSQL,
			Rationale:      rec.Rationale,
			Confidence:     rec.Confidence,
			ImpactEstimate: rec.ImpactEstimate,
			RiskLevel:      rec.RiskLevel,
		})
	}

	// Generate plan facts
	planFacts := h.generatePlanFacts(*targetQuery, tables)

	// Extract table names
	tableNames := h.extractTableNames(targetQuery.Query)

	// Generate fingerprint
	fingerprint := h.generateFingerprint(targetQuery.Query)

	queryDetail := QueryDetailDTO{
		QueryID:     queryID,
		Query:       targetQuery.Query,
		Fingerprint: fingerprint,
		Stats: QueryStatsDTO{
			Calls:          targetQuery.Calls,
			MeanExecTime:   targetQuery.MeanExecTime,
			TotalTime:      targetQuery.TotalTime,
			Rows:           targetQuery.Rows,
			SharedBlksHit:  targetQuery.SharedBlksHit,
			SharedBlksRead: targetQuery.SharedBlksRead,
		},
		Recommendations: recDTOs,
		PlanFacts:       planFacts,
		Tables:          tableNames,
	}

	logger.LogInfo("HTTP: Returning query detail")
	return c.JSON(queryDetail)
}

// Helper functions
func (h *Handlers) generateFingerprint(query string) string {
	// Simple MD5-based fingerprint
	hash := md5.Sum([]byte(query))
	return fmt.Sprintf("%x", hash)
}

// GetScanResults returns scan results similar to CLI scan command
func (h *Handlers) GetScanResults(c *fiber.Ctx) error {
	logger.LogInfo("HTTP: Getting scan results")

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	minDurationStr := c.Query("min_duration", "0.1")
	minDuration, err := strconv.ParseFloat(minDurationStr, 64)
	if err != nil {
		minDuration = 0.1
	}

	// Get slow queries
	queryStats, err := h.collector.GetSlowQueries(minDuration)
	if err != nil {
		logger.LogErrorf("Failed to get slow queries: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve query statistics",
		})
	}

	// Get metadata
	tables, err := h.collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to get table info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve table information",
		})
	}

	indexes, err := h.collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to get index info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve index information",
		})
	}

	// Convert to scan results
	var scanResults []ScanResultDTO
	for i, query := range queryStats {
		if i >= limit {
			break
		}

		// Generate recommendations
		recommendations := h.ruleEngine.AnalyzeQuery(query, tables, indexes)

		// Generate fingerprint
		fingerprint := h.generateFingerprint(query.Query)

		scanResult := ScanResultDTO{
			QueryID:         fingerprint[:12],
			Query:           query.Query,
			Calls:           query.Calls,
			MeanExecTime:    query.MeanExecTime,
			TotalTime:       query.TotalTime,
			Recommendations: len(recommendations),
		}

		scanResults = append(scanResults, scanResult)
	}

	logger.LogInfof("HTTP: Returning %d scan results", len(scanResults))
	return c.JSON(fiber.Map{
		"results": scanResults,
		"total":   len(scanResults),
		"limit":   limit,
	})
}

// GetSystemStatus returns database and system status
func (h *Handlers) GetSystemStatus(c *fiber.Ctx) error {
	logger.LogInfo("HTTP: Getting system status")

	// Get basic stats
	queryStats, err := h.collector.GetSlowQueries(0.0)
	if err != nil {
		logger.LogErrorf("Failed to get query stats: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve system status",
		})
	}

	tables, err := h.collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to get table info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve table information",
		})
	}

	indexes, err := h.collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to get index info: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve index information",
		})
	}

	// Calculate system metrics
	totalQueries := len(queryStats)
	slowQueries := 0
	totalTime := 0.0
	avgTime := 0.0

	for _, query := range queryStats {
		if query.MeanExecTime > 1.0 {
			slowQueries++
		}
		totalTime += query.TotalTime
	}

	if totalQueries > 0 {
		avgTime = totalTime / float64(totalQueries)
	}

	status := SystemStatusDTO{
		Database: DatabaseStatusDTO{
			Connected:       true,
			TotalQueries:    totalQueries,
			SlowQueries:     slowQueries,
			AvgResponseTime: avgTime,
		},
		Tables: TableStatusDTO{
			Count:     len(tables),
			TotalRows: h.calculateTotalRows(tables),
		},
		Indexes: IndexStatusDTO{
			Count:     len(indexes),
			TotalSize: h.calculateTotalIndexSize(indexes),
		},
		AI: AIStatusDTO{
			Enabled: h.ruleEngine != nil,
			Status:  h.getAIStatus(),
		},
	}

	return c.JSON(status)
}

// Helper functions for system status
func (h *Handlers) calculateTotalRows(tables []store.TableInfo) int64 {
	total := int64(0)
	for _, table := range tables {
		total += table.RowCount
	}
	return total
}

func (h *Handlers) calculateTotalIndexSize(indexes []store.IndexInfo) int64 {
	total := int64(0)
	for _, index := range indexes {
		total += index.SizeBytes
	}
	return total
}

func (h *Handlers) getAIStatus() string {
	if h.ruleEngine != nil {
		return "active"
	}
	return "disabled"
}

func (h *Handlers) generatePlanFacts(query store.QueryStats, tables []store.TableInfo) PlanFactsDTO {
	// Simple heuristics for plan facts
	hasSeqScan := query.MeanExecTime > 1.0 && query.SharedBlksRead > 0
	hasIndexScan := query.SharedBlksHit > query.SharedBlksRead

	// Estimate vs actual rows (simplified)
	estimatedRows := int64(float64(query.Rows) * 1.2) // Rough estimate
	actualRows := query.Rows

	selectivity := 0.0
	if len(tables) > 0 {
		// Find largest table for selectivity calculation
		maxRows := int64(0)
		for _, table := range tables {
			if table.RowCount > maxRows {
				maxRows = table.RowCount
			}
		}
		if maxRows > 0 {
			selectivity = float64(actualRows) / float64(maxRows)
		}
	}

	return PlanFactsDTO{
		HasSeqScan:    hasSeqScan,
		HasIndexScan:  hasIndexScan,
		EstimatedRows: estimatedRows,
		ActualRows:    actualRows,
		BuffersHit:    query.SharedBlksHit,
		BuffersRead:   query.SharedBlksRead,
		Selectivity:   selectivity,
	}
}

func (h *Handlers) extractTableNames(query string) []string {
	// Simple regex-based table extraction
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

// renderBottlenecksHTML renders bottlenecks as HTML for HTMX
func (h *Handlers) renderBottlenecksHTML(bottlenecks []BottleneckDTO, limit int) string {
	if len(bottlenecks) == 0 {
		return `
		<div class="text-center py-8 text-gray-400">
			<div class="text-6xl mb-4">🔍</div>
			<h3 class="text-xl font-semibold mb-2">No Bottlenecks Found</h3>
			<p class="text-sm">All queries are performing well! Try adjusting the filters above.</p>
		</div>`
	}

	html := ""
	for _, bottleneck := range bottlenecks {
		// Calculate performance score
		score := h.calculatePerformanceScoreFromBottleneck(bottleneck)
		scoreColor := h.getScoreColor(score)

		// Format execution time
		execTime := fmt.Sprintf("%.2fms", bottleneck.MeanExecTime)
		if bottleneck.MeanExecTime > 1000 {
			execTime = fmt.Sprintf("%.2fs", bottleneck.MeanExecTime/1000)
		}

		// Plan facts chips
		planChips := h.renderPlanFactsChips(bottleneck.PlanFacts)

		// Recommendations
		recommendationsHTML := h.renderRecommendations(bottleneck.Recommendations)

		// Query preview (truncated)
		queryPreview := bottleneck.Query
		if len(queryPreview) > 100 {
			queryPreview = queryPreview[:100] + "..."
		}

		html += fmt.Sprintf(`
		<div class="bg-white/5 backdrop-blur-sm rounded-lg p-6 mb-4 border border-white/10 hover:bg-white/10 transition-all duration-300">
			<div class="flex items-start justify-between mb-4">
				<div class="flex-1">
					<div class="flex items-center space-x-3 mb-2">
						<span class="text-sm font-mono text-blue-300">%s</span>
						<span class="px-2 py-1 text-xs rounded-full %s">%d%% Performance</span>
						<span class="px-2 py-1 text-xs rounded-full bg-red-500/20 text-red-300">%s</span>
					</div>
					<div class="text-sm text-gray-300 mb-3 font-mono bg-gray-800/50 p-3 rounded border-l-2 border-blue-500">
						%s
					</div>
					<div class="flex flex-wrap gap-2 mb-3">
						%s
					</div>
				</div>
				<div class="text-right text-sm text-gray-400">
					<div>Calls: <span class="text-white font-semibold">%d</span></div>
					<div>Rows: <span class="text-white font-semibold">%d</span></div>
					<div>Total: <span class="text-white font-semibold">%.2fms</span></div>
				</div>
			</div>
			%s
		</div>`,
			bottleneck.QueryID,
			scoreColor,
			score,
			execTime,
			queryPreview,
			planChips,
			bottleneck.Calls,
			bottleneck.Rows,
			bottleneck.TotalTime,
			recommendationsHTML,
		)
	}

	return html
}

// calculatePerformanceScoreFromBottleneck calculates a performance score (0-100)
func (h *Handlers) calculatePerformanceScoreFromBottleneck(bottleneck BottleneckDTO) int {
	// Base score on execution time and plan efficiency
	score := 100

	// Penalize slow execution times
	if bottleneck.MeanExecTime > 1000 {
		score -= 40
	} else if bottleneck.MeanExecTime > 100 {
		score -= 20
	} else if bottleneck.MeanExecTime > 10 {
		score -= 10
	}

	// Penalize sequential scans
	if bottleneck.PlanFacts.HasSeqScan && !bottleneck.PlanFacts.HasIndexScan {
		score -= 30
	}

	// Penalize low selectivity
	if bottleneck.PlanFacts.Selectivity > 0.5 {
		score -= 20
	}

	// Bonus for good buffer hit ratio
	if bottleneck.PlanFacts.BuffersHit > bottleneck.PlanFacts.BuffersRead {
		score += 10
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// getScoreColor returns CSS classes for performance score
func (h *Handlers) getScoreColor(score int) string {
	if score >= 80 {
		return "bg-green-500/20 text-green-300"
	} else if score >= 60 {
		return "bg-yellow-500/20 text-yellow-300"
	} else {
		return "bg-red-500/20 text-red-300"
	}
}

// renderPlanFactsChips renders plan facts as chips
func (h *Handlers) renderPlanFactsChips(facts PlanFactsDTO) string {
	chips := ""

	if facts.HasSeqScan {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-red-500/20 text-red-300">Seq Scan</span>`
	}
	if facts.HasIndexScan {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-green-500/20 text-green-300">Index Scan</span>`
	}

	// Row estimation accuracy
	accuracy := float64(facts.ActualRows) / float64(facts.EstimatedRows)
	if accuracy > 0.8 && accuracy < 1.2 {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-blue-500/20 text-blue-300">Good Est</span>`
	} else {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-orange-500/20 text-orange-300">Poor Est</span>`
	}

	// Buffer efficiency
	if facts.BuffersHit > facts.BuffersRead {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-green-500/20 text-green-300">Cache Hit</span>`
	} else {
		chips += `<span class="px-2 py-1 text-xs rounded-full bg-red-500/20 text-red-300">Disk Read</span>`
	}

	return chips
}

// renderRecommendations renders recommendations as HTML
func (h *Handlers) renderRecommendations(recommendations []RecommendationDTO) string {
	if len(recommendations) == 0 {
		return `
		<div class="text-center py-4 text-gray-400">
			<div class="text-2xl mb-2">✅</div>
			<p class="text-sm">No recommendations available</p>
		</div>`
	}

	html := `<div class="mt-4 space-y-3">`

	for _, rec := range recommendations {
		riskColor := h.getRiskColor(rec.RiskLevel)
		confidence := int(rec.Confidence * 100)

		html += fmt.Sprintf(`
		<div class="bg-gray-800/50 rounded-lg p-4 border-l-4 border-blue-500">
			<div class="flex items-start justify-between mb-2">
				<div class="flex items-center space-x-2">
					<span class="px-2 py-1 text-xs rounded-full bg-blue-500/20 text-blue-300">%s</span>
					<span class="px-2 py-1 text-xs rounded-full %s">%s Risk</span>
					<span class="px-2 py-1 text-xs rounded-full bg-purple-500/20 text-purple-300">%d%% Confidence</span>
				</div>
			</div>
			<p class="text-sm text-gray-300 mb-3">%s</p>
			<div class="bg-gray-900/50 p-3 rounded font-mono text-xs text-green-300 mb-2">
				%s
			</div>
			<p class="text-xs text-gray-400">%s</p>
		</div>`,
			strings.Title(rec.Type),
			riskColor,
			strings.Title(rec.RiskLevel),
			confidence,
			rec.Rationale,
			rec.DDL,
			rec.ImpactEstimate,
		)
	}

	html += `</div>`
	return html
}

// getRiskColor returns CSS classes for risk level
func (h *Handlers) getRiskColor(riskLevel string) string {
	switch strings.ToLower(riskLevel) {
	case "low":
		return "bg-green-500/20 text-green-300"
	case "medium":
		return "bg-yellow-500/20 text-yellow-300"
	case "high":
		return "bg-red-500/20 text-red-300"
	default:
		return "bg-gray-500/20 text-gray-300"
	}
}
