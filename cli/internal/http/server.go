package http

import (
	"cli/internal/db"
	"cli/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app      *fiber.App
	handlers *Handlers
}

func NewServer() *Server {
	// Create database config
	dbConfig := db.NewConfig()

	// Create handlers
	handlers := NewHandlers(dbConfig)
	if handlers == nil {
		logger.LogError("Failed to create HTTP handlers")
		return nil
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.LogErrorf("HTTP Error %d: %v", code, err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${ip})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Create server
	server := &Server{
		app:      app,
		handlers: handlers,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.app.Group("/api/v1")

	// Core analysis endpoints (matching CLI commands)
	api.Get("/scan", s.handlers.GetScanResults)        // CLI: optidb scan
	api.Get("/bottlenecks", s.handlers.GetBottlenecks) // CLI: optidb bottlenecks
	api.Get("/queries/:id", s.handlers.GetQueryDetail) // Query detail view

	// System status and monitoring
	api.Get("/status", s.handlers.GetSystemStatus) // System overview
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "optidb-api",
			"version": "1.0.0",
		})
	})

	// Dashboard routes
	s.app.Get("/", s.handlers.GetDashboard)
	s.app.Get("/dashboard", s.handlers.GetDashboard)

	// API documentation
	s.app.Get("/docs", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"endpoints": map[string]interface{}{
				"GET /api/v1/scan":        "Scan database for slow queries (CLI: optidb scan)",
				"GET /api/v1/bottlenecks": "Get performance bottlenecks (CLI: optidb bottlenecks)",
				"GET /api/v1/queries/:id": "Get detailed query analysis",
				"GET /api/v1/status":      "Get system status and metrics",
				"GET /api/v1/health":      "Health check endpoint",
				"GET /":                   "Main dashboard",
				"GET /dashboard":          "Dashboard (alias)",
			},
			"parameters": map[string]interface{}{
				"limit":        "Number of results to return (default: 10-20)",
				"min_duration": "Minimum query duration in ms (default: 0.1)",
				"type":         "Filter by analysis type (all, missing_index, correlated_subquery, etc.)",
			},
		})
	})
}

func (s *Server) Start(port string) error {
	if port == "" {
		port = "8090"
	}

	logger.LogInfof("Starting HTTP server on port %s", port)
	return s.app.Listen(":" + port)
}

func (s *Server) Stop() error {
	logger.LogInfo("Stopping HTTP server")
	return s.app.Shutdown()
}
