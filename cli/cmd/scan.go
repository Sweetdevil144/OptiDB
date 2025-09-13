package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"cli/internal/db"
	"cli/internal/ingest"
	"cli/internal/logger"
	"cli/internal/rules"
)

var (
	minDuration float64
	topN        int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan database for slow queries and performance issues",
	Long: `Analyze pg_stat_statements to identify slow queries and potential optimizations.
	
This command will:
- Pull query statistics from pg_stat_statements
- Analyze table and index metadata  
- Generate performance recommendations`,
	Run: func(cmd *cobra.Command, args []string) {
		runScan()
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().Float64Var(&minDuration, "min-duration", 0.1, "Minimum query duration in ms to analyze")
	scanCmd.Flags().IntVar(&topN, "top", 20, "Number of top queries to analyze")
}

func runScan() {
	logger.LogInfo("Starting database scan for performance issues")
	fmt.Println("🔍 Scanning database for performance issues...")

	// Connect to database as profiler_ro
	database, err := db.ConnectAsProfiler()
	if err != nil {
		logger.LogErrorf("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize components
	collector := ingest.NewStatsCollector(database)
	ruleEngine := rules.NewRuleEngine()
	logger.LogInfo("Initialized stats collector and rule engine")

	// Collect query statistics
	fmt.Println("📊 Collecting query statistics...")
	queryStats, err := collector.GetSlowQueries(minDuration)
	if err != nil {
		logger.LogErrorf("Failed to collect query stats: %v", err)
		log.Fatalf("Failed to collect query stats: %v", err)
	}

	if len(queryStats) == 0 {
		logger.LogInfof("No slow queries found with duration > %.1fms", minDuration)
		fmt.Printf("✅ No slow queries found (duration > %.1fms)\n", minDuration)
		return
	}

	// Collect table and index information
	fmt.Println("🗄️  Collecting table and index metadata...")
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

	// Analyze queries and generate recommendations
	fmt.Printf("🔬 Analyzing %d slow queries...\n", len(queryStats))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nQUERY\tCALLS\tAVG TIME (ms)\tTOTAL TIME (ms)\tRECOMMENDATIONS")
	fmt.Fprintln(w, "-----\t-----\t------------\t--------------\t---------------")

	totalRecommendations := 0

	for i, query := range queryStats {
		if i >= topN {
			break
		}

		// Parse and analyze query
		recommendations := ruleEngine.AnalyzeQuery(query, tables, indexes)

		// Display query summary
		shortQuery := query.Query
		if len(shortQuery) > 50 {
			shortQuery = shortQuery[:47] + "..."
		}

		recCount := len(recommendations)
		totalRecommendations += recCount

		fmt.Fprintf(w, "%s\t%d\t%.2f\t%.2f\t%d\n",
			shortQuery, query.Calls, query.MeanExecTime, query.TotalTime, recCount)

		// Show recommendations
		if recCount > 0 {
			fmt.Fprintf(w, "\t\t\t\t\n")
			for _, rec := range recommendations {
				fmt.Fprintf(w, "\t\t\t\t• %s (%.0f%% confidence)\n", rec.Type, rec.Confidence*100)
				if rec.DDL != "" {
					fmt.Fprintf(w, "\t\t\t\t  DDL: %s\n", rec.DDL)
				}
				fmt.Fprintf(w, "\t\t\t\t  %s\n", rec.Rationale)
				fmt.Fprintf(w, "\t\t\t\t\n")
			}
		}
	}

	w.Flush()

	// Summary
	fmt.Printf("\n📈 Scan Summary:\n")
	fmt.Printf("   • Analyzed %d slow queries\n", len(queryStats))
	fmt.Printf("   • Found %d tables with %d indexes\n", len(tables), len(indexes))
	fmt.Printf("   • Generated %d recommendations\n", totalRecommendations)

	if totalRecommendations > 0 {
		fmt.Printf("\n💡 Run 'optidb bottlenecks' to see detailed recommendations\n")
	}
}
