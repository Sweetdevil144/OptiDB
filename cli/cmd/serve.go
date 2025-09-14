package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cli/internal/http"
	"cli/internal/logger"

	"github.com/spf13/cobra"
)

var (
	port string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the OptiDB web server with dashboard",
	Long: `Start the OptiDB web server with a real-time dashboard for database performance analysis.

Features:
- Real-time bottleneck monitoring
- Interactive HTMX dashboard
- REST API endpoints for integration
- Query detail analysis
- AI-powered recommendations

Examples:
  optidb serve --port 8090
  optidb serve --port 3000`,
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize database extensions, roles, and meta-store.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("NOTE: Assuming extensions (pg_stat_statements, etc.) and roles are created.")
		fmt.Println("Initialization logic will go here.")
	},
}

func runServe() {
	logger.LogInfo("Starting OptiDB web server")

	// Create server
	server := http.NewServer()
	if server == nil {
		logger.LogError("Failed to create web server")
		os.Exit(1)
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.LogInfo("Shutting down server...")
		server.Stop()
		os.Exit(0)
	}()

	// Start server
	fmt.Printf("🚀 OptiDB Web Server starting on port %s\n", port)
	fmt.Printf("📊 Dashboard: http://localhost:%s\n", port)
	fmt.Printf("🔗 API: http://localhost:%s/api/v1\n", port)
	fmt.Printf("💡 Press Ctrl+C to stop\n\n")

	if err := server.Start(port); err != nil {
		logger.LogErrorf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(initCmd)

	serveCmd.Flags().StringVar(&port, "port", "8090", "Port to run the web server on")
}
