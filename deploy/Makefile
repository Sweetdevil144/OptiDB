# OptiDB Makefile - Docker and Database Operations

# Configuration
DOCKER_COMPOSE = docker-compose
PROJECT_NAME = optidb
POSTGRES_CONTAINER = optidb_postgres

# Default target
.PHONY: help
help: ## Show this help message
	@echo "OptiDB Development Commands:"
	@echo "==========================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker Operations
.PHONY: up
up: ## Start all services (Postgres + optional PgAdmin)
	$(DOCKER_COMPOSE) up -d
	@echo "✅ OptiDB services are starting..."
	@echo "   Postgres: localhost:5432"
	@echo "   PgAdmin (dev): localhost:8080"
	@echo ""
	@echo "Waiting for Postgres to be ready..."
	@sleep 5
	@$(MAKE) status

.PHONY: up-dev
up-dev: ## Start all services including PgAdmin
	$(DOCKER_COMPOSE) --profile dev up -d
	@echo "✅ OptiDB services (including PgAdmin) are starting..."

.PHONY: down
down: ## Stop and remove all containers
	$(DOCKER_COMPOSE) down
	@echo "✅ OptiDB services stopped"

.PHONY: down-clean
down-clean: ## Stop containers and remove volumes (⚠️  destroys data)
	$(DOCKER_COMPOSE) down -v
	@echo "⚠️  OptiDB services stopped and data volumes removed"

.PHONY: restart
restart: ## Restart all services
	$(MAKE) down
	$(MAKE) up

.PHONY: logs
logs: ## Show logs from all services
	$(DOCKER_COMPOSE) logs -f

.PHONY: logs-postgres
logs-postgres: ## Show Postgres logs only
	$(DOCKER_COMPOSE) logs -f postgres

# Database Operations
.PHONY: init
init: ## Initialize database (create extensions, roles, meta schema)
	@echo "🔧 Initializing OptiDB database..."
	@echo "Extensions, roles, and meta schema are created automatically on first startup"
	@$(MAKE) up
	@sleep 10
	@$(MAKE) status
	@echo "✅ OptiDB initialization complete!"

.PHONY: status
status: ## Check database connection and extensions
	@echo "📊 Checking OptiDB status..."
	@docker exec -e PGPASSWORD=postgres $(POSTGRES_CONTAINER) psql -U postgres -d optidb -c "SELECT current_database(), current_user;" 2>/dev/null || echo "❌ Database not ready"
	@docker exec -e PGPASSWORD=postgres $(POSTGRES_CONTAINER) psql -U postgres -d optidb -c "SELECT extname FROM pg_extension WHERE extname = 'pg_stat_statements';" 2>/dev/null || echo "Extensions not loaded"
	@docker exec -e PGPASSWORD=postgres $(POSTGRES_CONTAINER) psql -U postgres -d optidb -c "SELECT rolname FROM pg_roles WHERE rolname LIKE 'profiler_%' ORDER BY rolname;" 2>/dev/null || echo "Roles not created"

.PHONY: connect
connect: ## Connect to database as postgres user
	docker exec -it -e PGPASSWORD=postgres $(POSTGRES_CONTAINER) psql -U postgres -d optidb

.PHONY: connect-ro
connect-ro: ## Connect as read-only profiler user
	docker exec -it -e PGPASSWORD=profiler_ro_pass $(POSTGRES_CONTAINER) psql -U profiler_ro -d optidb

.PHONY: connect-sb
connect-sb: ## Connect as sandbox profiler user
	docker exec -it -e PGPASSWORD=profiler_sb_pass $(POSTGRES_CONTAINER) psql -U profiler_sb -d optidb

# Development and Testing
.PHONY: seed
seed: ## Load demo data with performance problems
	@echo "🌱 Seeding demo data with intentional performance issues..."
	docker exec -i -e PGPASSWORD=postgres $(POSTGRES_CONTAINER) psql -U postgres -d optidb < seed.sql
	@echo "✅ Demo data loaded with slow queries executed"

.PHONY: scan
scan: ## Run query statistics scan (requires OptiDB CLI)
	@echo "📈 Scanning query statistics..."
	@echo "TODO: Implement via OptiDB CLI"
	# ../cli/optidb scan --top 100 --min-mean-ms 5

.PHONY: demo
demo: ## Full demo: seed -> scan -> open UI
	@echo "🎬 Running OptiDB demo..."
	$(MAKE) up
	$(MAKE) seed
	$(MAKE) scan
	@echo "🌐 Opening UI at http://localhost:8090"
	@echo "TODO: Start OptiDB web server"

# Maintenance
.PHONY: reset-stats
reset-stats: ## Reset pg_stat_statements
	docker exec $(POSTGRES_CONTAINER) psql -U postgres -d optidb -c "SELECT pg_stat_statements_reset();"
	@echo "✅ Query statistics reset"

.PHONY: vacuum-analyze
vacuum-analyze: ## Run VACUUM ANALYZE on all tables
	docker exec $(POSTGRES_CONTAINER) psql -U postgres -d optidb -c "VACUUM ANALYZE;"
	@echo "✅ VACUUM ANALYZE completed"

# Testing
.PHONY: test
test: ## Run tests (to be implemented)
	@echo "🧪 Running OptiDB tests..."
	@echo "TODO: Implement tests"

# Utilities
.PHONY: ps
ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

.PHONY: shell
shell: ## Get shell access to Postgres container
	docker exec -it $(POSTGRES_CONTAINER) bash

.PHONY: backup
backup: ## Backup database to file
	@echo "💾 Creating database backup..."
	docker exec $(POSTGRES_CONTAINER) pg_dump -U postgres optidb > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup complete"

# Clean up
.PHONY: clean
clean: ## Remove all containers, volumes, and networks
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f
	@echo "✅ Cleanup complete"

# Environment Information
.PHONY: info
info: ## Show environment information
	@echo "OptiDB Environment Information:"
	@echo "==============================="
	@echo "Docker Compose: $(shell $(DOCKER_COMPOSE) version --short 2>/dev/null || echo 'Not installed')"
	@echo "Project: $(PROJECT_NAME)"
	@echo "Postgres Container: $(POSTGRES_CONTAINER)"
	@echo ""
	@echo "Database Connections:"
	@echo "  Admin: postgres://postgres:postgres@localhost:5432/optidb"
	@echo "  Read-Only: postgres://profiler_ro:profiler_ro_pass@localhost:5432/optidb"
	@echo "  Sandbox: postgres://profiler_sb:profiler_sb_pass@localhost:5432/optidb"
	@echo "  Service: postgres://profiler_service:profiler_service_pass@localhost:5432/optidb"
	@echo ""
	@echo "Web Interfaces:"
	@echo "  PgAdmin (dev): http://localhost:8080"
	@echo "  OptiDB UI: http://localhost:8090 (when running)"
