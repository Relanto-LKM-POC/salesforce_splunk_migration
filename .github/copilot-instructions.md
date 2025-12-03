# Salesforce to Splunk Migration Automation - AI Agent Instructions

## Project Overview
This is a **Golang CLI application** that automates Salesforce-to-Splunk data migration via REST APIs. The architecture follows a pipeline model: Splunk configuration (indexes, accounts, data inputs) → Dashboard XML validation → Pipeline trigger for deployment.

**Manual prerequisites (out of scope):**
1. Splunk Add-on for Salesforce must be pre-installed
2. Dashboard XML files must be manually created and stored in designated directory

## Architecture & Data Flow

### Core Components
1. **Configuration Layer** (`internal/config/`) - Viper-based config loading from YAML + env vars
2. **Splunk API Client** (`internal/splunk/`) - REST API wrapper with auth, retry, rate limiting
3. **Dashboard Pipeline** (`internal/dashboard/`) - XML scanner → validator → pipeline trigger → verifier
4. **Orchestrator** (`internal/orchestrator/`) - End-to-end migration workflow coordination
5. **Utilities** (`internal/utils/`) - Logging (Zap), retry with exponential backoff, validation

### Migration Workflow (8 Phases)
```
Init → Splunk Auth → Index Creation → Account Config → Data Input Setup → 
Dashboard XML Scan/Validate → Pipeline Trigger → Verification/Reporting
```

## Critical API Patterns

### Splunk REST API Conventions
All Splunk API calls follow this pattern from `working_curl.md`:

**CRITICAL: Always use HTTPS (not HTTP) for all Splunk API calls**

**Authentication:** Two methods supported
```bash
# Method 1: Basic Auth
-u username:password

# Method 2: Token Auth (preferred)
-H "Authorization: Splunk <token>"
```

**Creating Splunk Index:**
```bash
POST /services/data/indexes
URL: https://localhost:8089/services/data/indexes?output_mode=json
Parameters: name, datatype=event, maxTotalDataSizeMB
```

**Creating Salesforce Account in Splunk:**
```bash
POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account
URL: https://localhost:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json
Parameters: name, endpoint, sfdc_api_version, auth_type=basic, username, password, token
```

**Creating Data Input:**
```bash
POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object
URL: https://localhost:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json
Parameters: name, account, object, object_fields, order_by, start_date, interval, delay, index
```

### Project-Specific Conventions

1. **SSL Certificate Handling**: Always use `-k` flag equivalent (skip verification) - production Splunk instances use self-signed certs
2. **API Versioning**: Default to Salesforce API version `64.0` (configurable down to `42.0`)
3. **Endpoint Configuration**: 
   - Production: `login.salesforce.com`
   - Sandbox: `test.salesforce.com`
   - Custom: `yourcompany.my.salesforce.com`
4. **Index Naming**: Use lowercase with underscores (e.g., `salesforce_production`, `salesforce_uat`)

## Configuration Structure

### Primary Config Files
- `configs/config.yaml` - Main configuration (Splunk URL, credentials, API versions)
- `configs/objects.yaml` - Salesforce object definitions with field mappings
- `configs/mappings.yaml` - Dashboard to index mappings
- `.env` - Secrets (never commit - use `.env.example` as template)

### Configuration Loading Priority
1. Environment variables (highest)
2. `.env` file
3. YAML config files
4. Default values (lowest)

**Key environment variables:**
```
SPLUNK_URL, SPLUNK_USERNAME, SPLUNK_PASSWORD, SPLUNK_TOKEN
SALESFORCE_USERNAME, SALESFORCE_PASSWORD, SALESFORCE_TOKEN
```

## Development Workflows

### Building & Testing
```powershell
# Build
make build

# Run unit tests
make test

# Run with coverage
make coverage

# Integration tests (requires test Splunk instance)
make test-integration
```

### Running the Application
```powershell
# Full migration with dry-run
./salesforce-splunk-migration migrate --config configs/config.yaml --dry-run

# Specific phases
./salesforce-splunk-migration validate      # Pre-flight validation only
./salesforce-splunk-migration configure     # Splunk config only
./salesforce-splunk-migration deploy        # Dashboard deployment only
```

### Debugging
- Use `--log-level debug` flag for verbose output
- Logs sanitize sensitive data (passwords, tokens) automatically
- Check Splunk's internal logs: `index=_internal source=*salesforce*`

## Code Patterns & Standards

### Error Handling
```go
// Wrap errors with context - categorize by type
return pkg.errors.WrapConfig(err, "failed to load config file")
return pkg.errors.WrapAPI(err, "Splunk index creation failed")
```

### Retry Logic
All API calls use exponential backoff retry (configured in `internal/utils/retry.go`):
- Max retries: 3 (configurable)
- Initial delay: 1s
- Max delay: 30s
- Retry on: 429, 500, 502, 503, 504

### Logging Pattern
```go
logger.Info("creating Splunk index",
    zap.String("index", indexName),
    zap.String("datatype", "event"),
    zap.Int("maxSizeMB", maxSize))
```

### Testing Approach
- Use `testify/assert` for assertions
- Table-driven tests for multiple scenarios
- Mock Splunk API responses using `httptest`
- Integration tests optional (skip if no credentials)

## Dashboard Deployment Pipeline

**Critical distinction:** This application does NOT deploy dashboards directly. It:
1. Scans designated directory for dashboard XML files
2. Validates XML structure (SimpleXML format)
3. **Triggers external `splunk-dashboards` pipeline** with prepared XMLs
4. Monitors pipeline execution
5. Verifies deployment post-pipeline completion

### Dashboard XML Requirements
- Must be valid Splunk SimpleXML format
- Required elements: `<dashboard>`, `<label>`, `<row>`, `<panel>`
- Supported panel types: chart, table, single, event, map, viz
- Searches must reference correct index names (from mappings.yaml)

## Common Gotchas

1. **Splunk Add-on Verification**: Always verify `Splunk_TA_salesforce` add-on is installed before proceeding (see `internal/splunk/addon_verifier.go`)
2. **Field-Level Security**: Salesforce FLS permissions affect data ingestion - validate in pre-flight checks
3. **API Rate Limits**: Splunk Cloud has stricter limits than Enterprise - use rate limiter (`golang.org/x/time/rate`)
4. **Dashboard Index References**: Dashboard XMLs must reference indexes that exist - validate before deployment
5. **Security Token**: Salesforce requires security token appended to password unless IP whitelisted

## Key Files & Their Purpose

- `openapi.json` - Complete Splunk REST API schema (1336 lines) - reference for endpoint signatures
- `Salesforce_to_Splunk_Migration_Guide.md` - Manual process documentation (what we're automating)
- `working_curl.md` - Validated curl examples for all API calls - USE AS TRUTH
- `.taskmaster/tasks/tasks.json` - Detailed implementation roadmap (27 tasks, 5 phases)
- `.taskmaster/docs/prd.txt` - Full product requirements (scope, success criteria)

## Performance Targets
- <5 minutes for 10 dashboard migration (including validation)
- <500MB memory usage
- Support 100+ dashboards in single run
- Parallel API calls where safe (indexes, inputs)

## When Implementing New Features
1. Check `tasks.json` for task details and acceptance criteria
2. Reference `working_curl.md` for exact API call syntax
3. Follow error handling patterns from existing code
4. Add unit tests with >70% coverage target
5. Update `CHANGELOG.md` with changes
6. Sanitize sensitive data in logs

## Quick Reference: Task Status
See `.taskmaster/tasks/tasks.json` for complete task breakdown. Priority: TASK-001 through TASK-006 (Foundation Phase) must be completed before Phase 2.
