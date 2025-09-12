/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "optidb",
	Short: "AI-Powered Database Performance Profiler",
	Long: `OptiDB analyzes PostgreSQL query performance and provides actionable optimization recommendations.

Features:
- Scan pg_stat_statements for slow queries
- Detect missing indexes and inefficient queries  
- Generate DDL recommendations with confidence scores
- Analyze correlated subqueries and join patterns
- Plain English explanations for all recommendations

Examples:
  optidb scan --min-duration 1.0 --top 20
  optidb bottlenecks --limit 5
  optidb serve --port 8090`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
