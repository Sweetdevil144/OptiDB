package http

import (
	"strconv"

	"cli/internal/logger"
	"cli/internal/store"

	"github.com/gofiber/fiber/v2"
)

// GetDashboard renders the main dashboard with HTMX
func (h *Handlers) GetDashboard(c *fiber.Ctx) error {
	logger.LogInfo("HTTP: Rendering dashboard")

	// Render the dashboard HTML
	html := h.renderDashboardHTML()

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

func (h *Handlers) renderDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OptiDB - AI-Powered Database Performance Dashboard</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        .gradient-bg {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .glass-effect {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        .plan-fact-chip {
            @apply inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold transition-all duration-200;
        }
        .seq-scan { 
            @apply bg-gradient-to-r from-red-500 to-red-600 text-white shadow-lg;
        }
        .index-scan { 
            @apply bg-gradient-to-r from-green-500 to-green-600 text-white shadow-lg;
        }
        .confidence-high { 
            @apply bg-gradient-to-r from-green-500 to-green-600 text-white;
        }
        .confidence-medium { 
            @apply bg-gradient-to-r from-yellow-500 to-orange-500 text-white;
        }
        .confidence-low { 
            @apply bg-gradient-to-r from-red-500 to-red-600 text-white;
        }
        .risk-low { 
            @apply bg-gradient-to-r from-green-500 to-green-600 text-white;
        }
        .risk-medium { 
            @apply bg-gradient-to-r from-yellow-500 to-orange-500 text-white;
        }
        .risk-high { 
            @apply bg-gradient-to-r from-red-500 to-red-600 text-white;
        }
        .query-card {
            @apply bg-white rounded-xl shadow-lg hover:shadow-xl transition-all duration-300 border border-gray-100;
        }
        .query-card:hover {
            @apply transform -translate-y-1;
        }
        .metric-card {
            @apply bg-gradient-to-br from-blue-50 to-indigo-100 rounded-xl p-6 border border-blue-200;
        }
        .loading-spinner {
            @apply animate-spin rounded-full h-8 w-8 border-4 border-blue-200 border-t-blue-600;
        }
        .pulse-animation {
            animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: .5; }
        }
        .fade-in {
            animation: fadeIn 0.5s ease-in-out;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body class="bg-gradient-to-br from-gray-50 to-blue-50 min-h-screen" x-data="dashboard()">
    <div class="min-h-screen">
        <header class="gradient-bg shadow-2xl">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="flex justify-between items-center py-6">
                    <div class="flex items-center space-x-4">
                        <div class="bg-white rounded-xl p-3 shadow-lg">
                            <i class="fas fa-database text-2xl text-blue-600"></i>
                        </div>
                        <div>
                            <h1 class="text-3xl font-bold text-white">OptiDB</h1>
                            <p class="text-blue-100 text-sm">AI-Powered Database Performance Profiler</p>
                        </div>
                    </div>
                    <div class="flex items-center space-x-4">
                        <div class="glass-effect rounded-lg px-4 py-2">
                            <span class="text-white text-sm" id="last-updated">Last updated: Just now</span>
                        </div>
                        <button @click="refreshBottlenecks()" 
                                class="bg-white text-blue-600 px-6 py-3 rounded-lg font-semibold hover:bg-blue-50 transition-all duration-200 shadow-lg hover:shadow-xl flex items-center space-x-2">
                            <i class="fas fa-sync-alt" :class="{ 'animate-spin': loading }"></i>
                            <span>Refresh</span>
                        </button>
                    </div>
                </div>
            </div>
        </header>

        <main class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                <div class="metric-card">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-blue-600 text-sm font-medium">Total Queries</p>
                            <p class="text-3xl font-bold text-gray-900" id="total-queries">-</p>
                        </div>
                        <i class="fas fa-chart-line text-3xl text-blue-500"></i>
                    </div>
                </div>
                <div class="metric-card">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-orange-600 text-sm font-medium">Slow Queries</p>
                            <p class="text-3xl font-bold text-gray-900" id="slow-queries">-</p>
                        </div>
                        <i class="fas fa-exclamation-triangle text-3xl text-orange-500"></i>
                    </div>
                </div>
                <div class="metric-card">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-purple-600 text-sm font-medium">Recommendations</p>
                            <p class="text-3xl font-bold text-gray-900" id="total-recommendations">-</p>
                        </div>
                        <i class="fas fa-lightbulb text-3xl text-purple-500"></i>
                    </div>
                </div>
                <div class="metric-card">
                    <div class="flex items-center justify-between">
                        <div>
                            <p class="text-orange-600 text-sm font-medium">Avg Response</p>
                            <p class="text-3xl font-bold text-gray-900" id="avg-response">-</p>
                        </div>
                        <i class="fas fa-clock text-3xl text-orange-500"></i>
                    </div>
                </div>
            </div>

            <div class="bg-white rounded-xl shadow-lg p-6 mb-8 border border-gray-100">
                <div class="flex flex-wrap items-center gap-6">
                    <div class="flex-1 min-w-64">
                        <label class="block text-sm font-semibold text-gray-700 mb-2">Query Limit</label>
                        <select id="limit-select" @change="updateFilters()" 
                                class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200">
                            <option value="5">Top 5</option>
                            <option value="10" selected>Top 10</option>
                            <option value="20">Top 20</option>
                            <option value="50">Top 50</option>
                        </select>
                    </div>
                    <div class="flex-1 min-w-64">
                        <label class="block text-sm font-semibold text-gray-700 mb-2">Min Duration (ms)</label>
                        <input type="number" id="min-duration" value="0.1" step="0.1" min="0"
                               @change="updateFilters()"
                               class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200">
                    </div>
                    <div class="flex-1 min-w-64">
                        <label class="block text-sm font-semibold text-gray-700 mb-2">Analysis Type</label>
                        <select id="analysis-type" @change="updateFilters()" 
                                class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200">
                            <option value="all">All Issues</option>
                            <option value="missing_index">Missing Indexes</option>
                            <option value="correlated_subquery">Correlated Subqueries</option>
                            <option value="inefficient_join">Inefficient Joins</option>
                            <option value="redundant_index">Redundant Indexes</option>
                            <option value="cardinality_issue">Cardinality Issues</option>
                        </select>
                    </div>
                    <div class="flex items-end">
                        <button @click="exportData()" 
                                class="bg-gradient-to-r from-green-500 to-green-600 text-white px-6 py-3 rounded-lg font-semibold hover:from-green-600 hover:to-green-700 transition-all duration-200 shadow-lg hover:shadow-xl flex items-center space-x-2">
                            <i class="fas fa-download"></i>
                            <span>Export</span>
                        </button>
                    </div>
                </div>
            </div>

            <div class="bg-white rounded-xl shadow-lg border border-gray-100 overflow-hidden">
                <div class="px-6 py-4 bg-gradient-to-r from-gray-50 to-blue-50 border-b border-gray-200">
                    <div class="flex items-center justify-between">
                        <div>
                            <h2 class="text-2xl font-bold text-gray-900 flex items-center space-x-3">
                                <i class="fas fa-tachometer-alt text-blue-600"></i>
                                <span>Performance Bottlenecks</span>
                            </h2>
                            <p class="text-gray-600 mt-1">AI-powered analysis with actionable recommendations</p>
                        </div>
                        <div class="flex items-center space-x-4">
                            <div class="text-sm text-gray-500">
                                <span id="bottleneck-count" class="font-semibold">Loading...</span>
                            </div>
                            <div class="flex space-x-2">
                                <button @click="toggleView('cards')" 
                                        :class="{ 'bg-blue-600 text-white': viewMode === 'cards', 'bg-gray-200 text-gray-700': viewMode !== 'cards' }"
                                        class="px-3 py-2 rounded-lg transition-all duration-200">
                                    <i class="fas fa-th-large"></i>
                                </button>
                                <button @click="toggleView('table')" 
                                        :class="{ 'bg-blue-600 text-white': viewMode === 'table', 'bg-gray-200 text-gray-700': viewMode !== 'table' }"
                                        class="px-3 py-2 rounded-lg transition-all duration-200">
                                    <i class="fas fa-table"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div x-show="loading" class="flex justify-center items-center py-16">
                    <div class="text-center">
                        <div class="loading-spinner mx-auto mb-4"></div>
                        <p class="text-gray-600 font-medium">Analyzing database performance...</p>
                        <p class="text-gray-500 text-sm">This may take a few moments</p>
                    </div>
                </div>

                <div id="bottlenecks-content" 
                     hx-get="/api/v1/bottlenecks?limit=10&min_duration=0.1" 
                     hx-trigger="load"
                     hx-target="this"
                     hx-swap="innerHTML"
                     class="fade-in">
                </div>
            </div>
        </main>
    </div>

    <script>
        function dashboard() {
            return {
                loading: false,
                viewMode: 'cards',
                
                async refreshBottlenecks() {
                    this.loading = true;
                    const limit = document.getElementById('limit-select').value;
                    const minDuration = document.getElementById('min-duration').value;
                    const analysisType = document.getElementById('analysis-type').value;
                    
                    try {
                        await htmx.ajax('GET', '/api/v1/bottlenecks?limit=' + limit + '&min_duration=' + minDuration + '&type=' + analysisType, {
                            target: '#bottlenecks-content',
                            swap: 'innerHTML'
                        });
                        
                        document.getElementById('last-updated').textContent = 'Last updated: ' + new Date().toLocaleTimeString();
                        this.updateStats();
                    } catch (error) {
                        console.error('Failed to refresh bottlenecks:', error);
                    } finally {
                        this.loading = false;
                    }
                },
                
                updateFilters() {
                    this.refreshBottlenecks();
                },
                
                toggleView(mode) {
                    this.viewMode = mode;
                    // Re-render content with new view mode
                    this.refreshBottlenecks();
                },
                
                updateStats() {
                    // Update stats from the loaded data
                    const content = document.getElementById('bottlenecks-content');
                    const queries = content.querySelectorAll('.query-card, tr[data-query]');
                    const recommendations = content.querySelectorAll('.recommendation-item');
                    
                    document.getElementById('total-queries').textContent = queries.length;
                    document.getElementById('slow-queries').textContent = queries.length;
                    document.getElementById('total-recommendations').textContent = recommendations.length;
                    
                    // Calculate average response time
                    let totalTime = 0;
                    queries.forEach(query => {
                        const timeEl = query.querySelector('[data-time]');
                        if (timeEl) {
                            totalTime += parseFloat(timeEl.textContent) || 0;
                        }
                    });
                    const avgTime = queries.length > 0 ? (totalTime / queries.length).toFixed(2) : '0.00';
                    document.getElementById('avg-response').textContent = avgTime + 'ms';
                },
                
                exportData() {
                    // Export functionality
                    const data = {
                        timestamp: new Date().toISOString(),
                        filters: {
                            limit: document.getElementById('limit-select').value,
                            minDuration: document.getElementById('min-duration').value,
                            analysisType: document.getElementById('analysis-type').value
                        },
                        data: document.getElementById('bottlenecks-content').innerHTML
                    };
                    
                    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = 'optidb-analysis-' + new Date().toISOString().split('T')[0] + '.json';
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                },
                
                viewQueryDetail(queryId) {
                    window.open('/api/v1/queries/' + queryId, '_blank');
                }
            }
        }
    </script>
</body>
</html>`
}

// GetBottlenecksTable renders the bottlenecks table for HTMX
func (h *Handlers) GetBottlenecksTable(c *fiber.Ctx) error {
	logger.LogInfo("HTTP: Rendering bottlenecks table")

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

	analysisType := c.Query("type", "all")

	// Get slow queries
	queryStats, err := h.collector.GetSlowQueries(minDuration)
	if err != nil {
		logger.LogErrorf("Failed to get slow queries: %v", err)
		return c.SendString(`<div class="text-center py-16 text-red-600">
            <i class="fas fa-exclamation-triangle text-4xl mb-4"></i>
            <p class="text-xl font-semibold">Error loading bottlenecks</p>
            <p class="text-gray-500">Please try refreshing the page</p>
        </div>`)
	}

	// Get metadata
	tables, err := h.collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to get table info: %v", err)
		return c.SendString(`<div class="text-center py-16 text-red-600">
            <i class="fas fa-exclamation-triangle text-4xl mb-4"></i>
            <p class="text-xl font-semibold">Error loading table information</p>
        </div>`)
	}

	indexes, err := h.collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to get index info: %v", err)
		return c.SendString(`<div class="text-center py-16 text-red-600">
            <i class="fas fa-exclamation-triangle text-4xl mb-4"></i>
            <p class="text-xl font-semibold">Error loading index information</p>
        </div>`)
	}

	// Generate HTML content
	html := `<div class="p-6">`

	if len(queryStats) == 0 {
		html += `<div class="text-center py-16">
            <i class="fas fa-check-circle text-6xl text-green-500 mb-4"></i>
            <h3 class="text-2xl font-bold text-gray-900 mb-2">No Bottlenecks Found</h3>
            <p class="text-gray-600">Your database is performing well! No slow queries detected.</p>
        </div>`
	} else {
		// Cards View
		html += `<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">`

		for i, query := range queryStats {
			if i >= limit {
				break
			}

			// Generate recommendations
			recommendations := h.ruleEngine.AnalyzeQuery(query, tables, indexes)

			// Filter by analysis type if specified
			if analysisType != "all" {
				filteredRecs := []store.Recommendation{}
				for _, rec := range recommendations {
					if rec.Type == analysisType {
						filteredRecs = append(filteredRecs, rec)
					}
				}
				recommendations = filteredRecs
			}

			// Skip if no recommendations after filtering
			if analysisType != "all" && len(recommendations) == 0 {
				continue
			}

			// Generate plan facts
			planFacts := h.generatePlanFacts(query, tables)

			// Generate fingerprint
			fingerprint := h.generateFingerprint(query.Query)
			queryID := fingerprint[:12]

			// Truncate query for display
			displayQuery := query.Query
			if len(displayQuery) > 120 {
				displayQuery = displayQuery[:117] + "..."
			}

			// Calculate performance score
			performanceScore := h.calculatePerformanceScore(query, planFacts)

			html += `<div class="query-card p-6" data-query="true">
                <div class="flex items-start justify-between mb-4">
                    <div class="flex-1">
                        <div class="flex items-center space-x-2 mb-2">
                            <span class="text-xs font-semibold text-gray-500 bg-gray-100 px-2 py-1 rounded">ID: ` + queryID + `</span>
                            <div class="flex items-center space-x-1">
                                <span class="text-sm font-medium text-gray-600">Performance:</span>
                                <span class="text-sm font-bold ` + h.getPerformanceColor(performanceScore) + `">` + strconv.Itoa(performanceScore) + `%</span>
                            </div>
                        </div>
                        <div class="bg-gray-50 rounded-lg p-3 mb-4">
                            <code class="text-sm text-gray-800 font-mono break-all">` + displayQuery + `</code>
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-2 gap-4 mb-4">
                    <div class="text-center">
                        <div class="text-2xl font-bold text-blue-600" data-time="` + strconv.FormatFloat(query.MeanExecTime, 'f', 2, 64) + `">` + strconv.FormatFloat(query.MeanExecTime, 'f', 2, 64) + `ms</div>
                        <div class="text-xs text-gray-500">Avg Time</div>
                    </div>
                    <div class="text-center">
                        <div class="text-2xl font-bold text-green-600">` + strconv.FormatInt(query.Calls, 10) + `</div>
                        <div class="text-xs text-gray-500">Calls</div>
                    </div>
                </div>

                <div class="mb-4">
                    <div class="flex flex-wrap gap-2">`

			// Plan facts chips
			if planFacts.HasSeqScan {
				html += `<span class="plan-fact-chip seq-scan">
                    <i class="fas fa-search mr-1"></i>Seq Scan
                </span>`
			}
			if planFacts.HasIndexScan {
				html += `<span class="plan-fact-chip index-scan">
                    <i class="fas fa-database mr-1"></i>Index Scan
                </span>`
			}
			html += `<span class="plan-fact-chip bg-blue-100 text-blue-800">
                <i class="fas fa-chart-line mr-1"></i>Est: ` + strconv.FormatInt(planFacts.EstimatedRows, 10) + `
            </span>`
			html += `<span class="plan-fact-chip bg-purple-100 text-purple-800">
                <i class="fas fa-chart-bar mr-1"></i>Act: ` + strconv.FormatInt(planFacts.ActualRows, 10) + `
            </span>`

			html += `</div>
                </div>

                <div class="mb-4">
                    <div class="flex items-center justify-between mb-2">
                        <span class="text-sm font-semibold text-gray-700">Recommendations</span>
                        <span class="text-xs text-gray-500">` + strconv.Itoa(len(recommendations)) + ` found</span>
                    </div>`

			// Show recommendations
			for j, rec := range recommendations {
				if j >= 3 { // Show only first 3
					html += `<div class="text-xs text-gray-500 text-center py-2">
                        +` + strconv.Itoa(len(recommendations)-3) + ` more recommendations...
                    </div>`
					break
				}

				confidenceClass := "confidence-low"
				if rec.Confidence > 0.7 {
					confidenceClass = "confidence-high"
				} else if rec.Confidence > 0.4 {
					confidenceClass = "confidence-medium"
				}

				html += `<div class="recommendation-item bg-gray-50 rounded-lg p-3 mb-2">
                    <div class="flex items-center justify-between mb-1">
                        <span class="text-sm font-medium text-gray-800">` + rec.Type + `</span>
                        <span class="plan-fact-chip ` + confidenceClass + ` text-xs">
                            ` + strconv.FormatFloat(rec.Confidence*100, 'f', 0, 64) + `%
                        </span>
                    </div>
                    <p class="text-xs text-gray-600">` + rec.Rationale + `</p>
                </div>`
			}

			html += `</div>

                <div class="flex space-x-2">
                    <button onclick="viewQueryDetail('` + queryID + `')" 
                            class="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors duration-200 flex items-center justify-center space-x-2">
                        <i class="fas fa-eye"></i>
                        <span>View Details</span>
                    </button>
                    <button onclick="copyToClipboard('` + queryID + `')" 
                            class="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-300 transition-colors duration-200">
                        <i class="fas fa-copy"></i>
                    </button>
                </div>
            </div>`
		}

		html += `</div>`
	}

	html += `</div>`

	// Update count
	html += `<script>
        document.getElementById('bottleneck-count').textContent = '` + strconv.Itoa(len(queryStats)) + ` bottlenecks found';
    </script>`

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// Helper functions
func (h *Handlers) calculatePerformanceScore(query store.QueryStats, planFacts PlanFactsDTO) int {
	// Simple performance scoring based on execution time and plan facts
	score := 100

	// Penalize for high execution time
	if query.MeanExecTime > 10 {
		score -= 40
	} else if query.MeanExecTime > 5 {
		score -= 25
	} else if query.MeanExecTime > 1 {
		score -= 10
	}

	// Penalize for sequential scans
	if planFacts.HasSeqScan {
		score -= 20
	}

	// Bonus for index scans
	if planFacts.HasIndexScan {
		score += 10
	}

	// Penalize for poor selectivity
	if planFacts.Selectivity > 0.5 {
		score -= 15
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func (h *Handlers) getPerformanceColor(score int) string {
	if score >= 80 {
		return "text-green-600"
	} else if score >= 60 {
		return "text-yellow-600"
	} else {
		return "text-red-600"
	}
}
