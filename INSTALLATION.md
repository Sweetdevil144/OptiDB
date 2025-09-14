## **COMPLETE SETUP GUIDE** (Easy Replication Steps)

### **Prerequisites**

- Docker & Docker Compose installed
- Go 1.23+ installed
- PostgreSQL client tools (optional, for manual testing)

### **Step 1: Database Setup (2 minutes)**

```bash
cd deploy
make down-clean  # Clean start
make up          # Start PostgreSQL 16 with extensions
make status      # Verify: database + extensions + roles
make seed        # Load demo data with slow queries
```

### **Step 2: CLI Setup (1 minute)**

```bash
cd ../cli
go build -o optidb  # Build CLI

# Optional: Create .env file for AI features
# cp .env.example .env  # (blocked by gitignore)
# Edit .env with your Azure OpenAI credentials
```

### **Step 3: Test AI-Powered Analysis (30 seconds)**

```bash
# Test with AI (if .env configured)
./optidb scan --min-duration 0.01 --top 5

# Test detailed recommendations
./optidb bottlenecks --limit 3

# Check logs for AI API calls and token usage
```

### **Step 4: Verify Everything Works**

Expected output should show:

- ✅ Database connection established
- ✅ AI-powered recommendations enabled (if configured)
- ✅ 2-4 recommendations generated with confidence scores
- ✅ DDL statements and plain English explanations
- ✅ Real OpenAI API calls with token tracking

### **Troubleshooting**

- **Connection refused**: Run `brew services stop postgresql@14` to stop local PostgreSQL
- **No slow queries**: Lower threshold with `--min-duration 0.001`
- **AI disabled**: Check .env file or use without AI (falls back to heuristics)