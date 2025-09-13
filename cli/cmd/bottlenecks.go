package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"cli/internal/db"
	"cli/internal/ingest"
	"cli/internal/logger"
	"cli/internal/parse"
	"cli/internal/rules"
)

var (
	showDDL bool
	limit   int
)

var bottlenecksCmd = &cobra.Command{
	Use:   "bottlenecks",
	Short: "Show top database performance bottlenecks with recommendations",
	Long: `Display the most problematic queries with detailed optimization recommendations.
	
This command shows:
- Slowest queries with highest impact
- Specific DDL recommendations 
- Plain English explanations
- Confidence scores and risk levels`,
	Run: func(cmd *cobra.Command, args []string) {
		runBottlenecks()
	},
}

func init() {
	rootCmd.AddCommand(bottlenecksCmd)

	bottlenecksCmd.Flags().BoolVar(&showDDL, "ddl", true, "Show DDL recommendations")
	bottlenecksCmd.Flags().IntVar(&limit, "limit", 10, "Number of bottlenecks to show")
}

func runBottlenecks() {
	logger.LogInfo("Starting bottlenecks analysis")
	fmt.Println("🚨 Top Database Performance Bottlenecks")
	fmt.Println("=====================================")

	// Connect to database
	database, err := db.ConnectAsProfiler()
	if err != nil {
		logger.LogErrorf("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize components
	collector := ingest.NewStatsCollector(database)
	parser := parse.NewQueryParser()
	ruleEngine := rules.NewRuleEngine()
	logger.LogInfo("Initialized components for bottlenecks analysis")

	// Get slow queries
	queryStats, err := collector.GetSlowQueries(0.1) // 0.1ms threshold
	if err != nil {
		logger.LogErrorf("Failed to collect query stats: %v", err)
		log.Fatalf("Failed to collect query stats: %v", err)
	}

	if len(queryStats) == 0 {
		logger.LogInfo("No performance bottlenecks detected")
		fmt.Println("✅ No performance bottlenecks detected!")
		return
	}

	// Get metadata
	tables, err := collector.GetTableInfo()
	if err != nil {
		logger.LogErrorf("Failed to collect table info: %v", err)
		log.Fatalf("Failed to collect table info: %v", err)
	}

	indexes, err := collector.GetIndexInfo()
	if err != nil {
		logger.LogErrorf("Failed to collect index info: %v", err)
		log.Fatalf("Failed to collect index info: %v", err)
	}

	// Analyze and display bottlenecks
	logger.LogInfof("Analyzing %d queries for bottlenecks (limit: %d)", len(queryStats), limit)
	count := 0
	for _, query := range queryStats {
		if count >= limit {
			break
		}

		recommendations := ruleEngine.AnalyzeQuery(query, tables, indexes)
		if len(recommendations) == 0 {
			logger.LogDebugf("No recommendations for query: %s", query.Query[:min(50, len(query.Query))])
			continue // Skip queries with no recommendations
		}

		count++
		logger.LogInfof("Found bottleneck #%d with %d recommendations", count, len(recommendations))

		fmt.Printf("\n🔴 Bottleneck #%d\n", count)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━\n")

		// Query info
		fmt.Printf("📊 Query Stats:\n")
		fmt.Printf("   • Calls: %d\n", query.Calls)
		fmt.Printf("   • Avg Time: %.2f ms\n", query.MeanExecTime)
		fmt.Printf("   • Total Time: %.2f ms\n", query.TotalTime)
		fmt.Printf("   • Rows: %d\n", query.Rows)

		// Query fingerprint
		fingerprint := parser.GenerateFingerprint(query.Query)
		fmt.Printf("   • Fingerprint: %s\n", fingerprint[:12]+"...")

		// Show query (truncated)
		displayQuery := query.Query
		if len(displayQuery) > 200 {
			displayQuery = displayQuery[:197] + "..."
		}
		fmt.Printf("   • SQL: %s\n", displayQuery)

		// Recommendations
		fmt.Printf("\n💡 Recommendations (%d):\n", len(recommendations))

		for i, rec := range recommendations {
			fmt.Printf("\n   %d. %s\n", i+1, formatRecommendationType(rec.Type))
			fmt.Printf("      🎯 Confidence: %.0f%%\n", rec.Confidence*100)
			fmt.Printf("      ⚠️  Risk Level: %s\n", rec.RiskLevel)

			if rec.DDL != "" && showDDL {
				fmt.Printf("      🔧 DDL:\n")
				fmt.Printf("         %s\n", rec.DDL)
			}

			if rec.RewriteSQL != "" {
				fmt.Printf("      ✏️  Rewrite Suggestion:\n")
				fmt.Printf("         %s\n", rec.RewriteSQL)
			}

			fmt.Printf("      📝 Why: %s\n", rec.Rationale)

			if rec.ImpactEstimate != "" {
				fmt.Printf("      📈 Expected Impact: %s\n", rec.ImpactEstimate)
			}
		}

		fmt.Printf("\n" + strings.Repeat("─", 50))
	}

	if count == 0 {
		logger.LogInfo("No actionable bottlenecks found in top queries")
		fmt.Println("✅ No actionable bottlenecks found in top queries!")
	} else {
		logger.LogInfof("Bottlenecks analysis complete: found %d bottlenecks", count)
		fmt.Printf("\n\n📋 Summary: Found %d bottlenecks with optimization opportunities\n", count)
		fmt.Printf("💡 Use --ddl=false to hide DDL statements\n")
		fmt.Printf("🔧 Use --limit=N to show more/fewer results\n")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatRecommendationType(recType string) string {
	switch recType {
	case "missing_index":
		return "Missing Index"
	case "composite_index":
		return "Composite Index Opportunity"
	case "correlated_subquery":
		return "Correlated Subquery Optimization"
	case "join_index":
		return "JOIN Index Missing"
	case "redundant_index":
		return "Redundant Index"
	default:
		return recType
	}
}
