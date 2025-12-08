# Salesforce to Splunk Migration Tool

A production-ready Golang application that automates Salesforce-to-Splunk data migration using FlowGraph orchestration and REST APIs. Built with enterprise-grade reliability, concurrency, and observability.

## Features

- ✅ **FlowGraph Orchestration** - Graph-based workflow execution with state management and checkpointing
- ✅ **Splunk Authentication** - Token-based authentication with automatic session management
- ✅ **Splunk Index Creation** - Automated index provisioning with conflict handling
- ✅ **Add-on Verification** - Validates Splunk Add-on for Salesforce installation
- ✅ **Account Configuration** - Salesforce account setup in Splunk with OAuth support
- ✅ **Parallel Data Inputs** - Concurrent creation of Salesforce object data inputs
- ✅ **Input Verification** - Post-creation validation of configured data inputs
- ✅ **Retry Logic** - Exponential backoff with configurable retry parameters
- ✅ **Connection Pooling** - HTTP client with connection reuse and keepalive
- ✅ **SSL Certificate Bypass** - Support for self-signed certificates
- ✅ **Structured Logging** - Zap-based logging with context and correlation
- ✅ **Comprehensive Testing** - Unit and integration tests with mocks
- ✅ **Graceful Error Handling** - Detailed error messages and recovery strategies

## Prerequisites

### Required Components
1. **Splunk Add-on for Salesforce** must be pre-installed on your Splunk instance
2. Splunk instance must be running and accessible (default port: 8089)
3. Valid Salesforce credentials with API access
4. Go 1.23.0 or later (for development)

## Installation

1. Clone the repository:
```powershell
git clone https://github.com/Relanto-LKM-POC/salesforce_splunk_migration.git
cd salesforce_splunk_migration
```

2. Install dependencies (includes embedded FlowGraph):
```powershell
go mod download
```

3. Configure credentials:
```powershell
cp credentials.json.example credentials.json
# Edit credentials.json with your actual Splunk and Salesforce credentials
```

## Configuration

The application uses `credentials.json` for configuration. The file supports both environment variable expansion and direct values.

### Configuration Structure

```json
{
  "SPLUNK_URL": "https://your-splunk-instance:8089",
  "SPLUNK_USERNAME": "admin",
  "SPLUNK_PASSWORD": "your-password",
  "SPLUNK_SKIP_SSL_VERIFY": true,
  "SPLUNK_INDEX_NAME": "salesforce",
  "SPLUNK_DEFAULT_INDEX": "salesforce",
  "SPLUNK_REQUEST_TIMEOUT": 30,
  "SPLUNK_MAX_RETRIES": 3,
  "SPLUNK_RETRY_DELAY": 2,
  
  "SALESFORCE_ENDPOINT": "https://login.salesforce.com",
  "SALESFORCE_API_VERSION": "64.0",
  "SALESFORCE_AUTH_TYPE": "oauth_client_credentials",
  "SALESFORCE_CLIENT_ID": "your-client-id",
  "SALESFORCE_CLIENT_SECRET": "your-client-secret",
  "SALESFORCE_ACCOUNT_NAME": "my_salesforce_account",
  
  "MIGRATION_CONCURRENT_REQUESTS": 5,
  "MIGRATION_DASHBOARD_DIRECTORY": "./resources/dashboards",
  "MIGRATION_LOG_LEVEL": "info",
  
  "DATA_INPUTS": [
    {
      "name": "account_input",
      "object": "Account",
      "object_fields": "*",
      "order_by": "LastModifiedDate",
      "start_date": "2020-01-01T00:00:00.000Z",
      "interval": 300,
      "delay": 0,
      "index": "salesforce"
    }
  ]
}
```

### Key Configuration Options

**Splunk Settings:**
- `SPLUNK_URL`: Splunk REST API endpoint (must include port 8089)
- `SPLUNK_SKIP_SSL_VERIFY`: Set to `true` for self-signed certificates
- `SPLUNK_MAX_RETRIES`: Number of retry attempts for failed requests (default: 3)
- `SPLUNK_RETRY_DELAY`: Initial delay between retries in seconds (default: 2)

**Salesforce Settings:**
- `SALESFORCE_AUTH_TYPE`: Either `oauth_client_credentials` or `basic`
- `SALESFORCE_API_VERSION`: Salesforce API version (default: 64.0)
- `SALESFORCE_ACCOUNT_NAME`: Unique identifier for the account in Splunk

**Migration Settings:**
- `MIGRATION_CONCURRENT_REQUESTS`: Number of parallel data input creations (default: 5)
- `MIGRATION_DASHBOARD_DIRECTORY`: Path to directory containing dashboard XML files (default: `./resources/dashboards`)
- `MIGRATION_LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`)

**Data Inputs:**
- Array of Salesforce objects to monitor
- Each input specifies the object type, fields, polling interval, and target index

### Dashboard Creation (Optional)

To enable automatic dashboard creation during migration:

1. **Create Dashboard Directory** (if not exists):
   ```powershell
   mkdir -p resources/dashboards
   ```

2. **Add Dashboard XML Files**: Place Splunk dashboard XML files in the directory:
   ```
   resources/
     dashboards/
       analytics_dashboard.xml
       home_dashboard.xml
   ```

3. **Configure Path** (already set in example):
   ```json
   {
     "MIGRATION_DASHBOARD_DIRECTORY": "./resources/dashboards"
   }
   ```

**Behavior**: 
- If configured and directory exists, all XML files are processed and dashboards created in Splunk
- If not configured or directory missing, this step is gracefully skipped
- Dashboard names are derived from filenames (without .xml extension)

## Usage

### Quick Start

Run the full migration workflow:

```powershell
# Build and run
go build -o salesforce-splunk-migration.exe .
.\salesforce-splunk-migration.exe

# Or run directly with Go
go run .
```

The application automatically:
1. Loads configuration from `credentials.json` (or path specified by `VAULT_PATH` env var)
2. Validates all configuration settings
3. Executes the complete migration workflow using FlowGraph orchestration

### Workflow Execution

The application runs as a single automated workflow powered by FlowGraph orchestration. There are no separate commands - the migration executes all steps in sequence:

1. **Authentication** - Authenticate with Splunk REST API
2. **Add-on Verification** - Check Splunk Add-on for Salesforce is installed
3. **Index Creation** - Create specified Splunk index
4. **Account Setup** - Configure Salesforce account credentials
5. **Load Inputs** - Parse data input configurations
6. **Create Inputs** - Create data inputs in parallel with concurrency control
7. **Verify Inputs** - Validate all inputs were created successfully
8. **Create Dashboards** - Create Splunk dashboards from XML templates (optional, skipped if not configured)

### Build the Application

```powershell
go build -o salesforce-splunk-migration.exe .
```

This creates `salesforce-splunk-migration.exe` in the current directory.

## Development

### Run Tests

```powershell
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./services
go test -v ./internal/workflows
```

### Run with Coverage

```powershell
# Generate coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

This generates `coverage.html` for viewing detailed coverage reports.

### Environment Variables

The application can read configuration from environment variables or `credentials.json`:

```powershell
# Set via environment (Windows PowerShell)
$env:SPLUNK_URL = "https://localhost:8089"
$env:SPLUNK_USERNAME = "admin"
$env:SPLUNK_PASSWORD = "changeme"

# Run with VAULT_PATH to specify custom config location
$env:VAULT_PATH = "C:\path\to\credentials.json"
go run .
```

## Architecture

The application is built with a clean, layered architecture for maintainability and testability:

```
salesforce_splunk_migration/
│
├── main.go                      # Application entry point, logger initialization
├── cmd/
│   └── app.go                   # Main execution flow, workflow orchestration
│
├── internal/
│   └── workflows/               # FlowGraph-based workflow implementation
│       ├── migration_graph.go   # Graph structure definition and execution
│       └── migration_processor.go # Custom node processor for migration steps
│
├── services/
│   └── splunk_service.go        # Splunk REST API client with retry logic
│
├── models/
│   ├── request.go               # Request data structures
│   └── response.go              # Response data structures (auth, Splunk API)
│
├── utils/
│   ├── config.go                # Configuration loading and validation
│   ├── http_client.go           # HTTP client with connection pooling
│   └── logger.go                # Structured logging with Zap
│
├── mocks/
│   ├── http_client_mock.go      # Mock HTTP client for testing
│   └── splunk_service_mock.go   # Mock Splunk service for testing
│
├── resources/
│   └── dashboards/              # Splunk dashboard XML templates
│
└── flowgraph/                   # Embedded FlowGraph framework
    ├── pkg/flowgraph/           # Core graph runtime and execution engine
    ├── internal/                # Internal graph components
    └── examples/                # FlowGraph usage examples
```

### Key Components

**FlowGraph Orchestration (`internal/workflows/`)**
- **MigrationGraph**: Defines the directed acyclic graph (DAG) of migration steps
- **MigrationNodeProcessor**: Implements custom processing logic for each node
- **State Management**: Thread-safe counters and state tracking across workflow
- **Error Handling**: Node-level error propagation and recovery

**Services Layer (`services/`)**
- **SplunkService**: Encapsulates all Splunk REST API interactions
- Interface-based design for easy mocking and testing
- Connection pooling and keepalive for performance
- Automatic retry with exponential backoff

**HTTP Client (`utils/http_client.go`)**
- Configurable timeout, retry, and connection pooling
- SSL certificate verification bypass support
- Request/response logging for debugging
- Context-aware cancellation

**Logging (`utils/logger.go`)**
- Structured logging with Zap
- Contextual fields (correlation IDs, error details)
- Configurable log levels
- Production-ready JSON output

**Configuration (`utils/config.go`)**
- Environment variable and JSON file support
- Type-safe configuration with validation
- Dynamic data input loading
- Secure credential handling

## Splunk API Reference

The tool interacts with the following Splunk REST API endpoints:

### Authentication
```
POST /services/auth/login
Content-Type: application/x-www-form-urlencoded

username=<username>&password=<password>&output_mode=json

Response: { "sessionKey": "<token>" }
```

### Check Add-on Installation
```
GET /services/apps/local/Splunk_TA_salesforce?output_mode=json
Authorization: Splunk <token>

Response: { "entry": [{ "name": "Splunk_TA_salesforce", ... }] }
```

### Create Index
```
POST /services/data/indexes
Authorization: Splunk <token>
Content-Type: application/x-www-form-urlencoded

name=<index_name>&datatype=event&output_mode=json
```

### Create Salesforce Account
```
POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account
Authorization: Splunk <token>
Content-Type: application/x-www-form-urlencoded

name=<account_name>&endpoint=<sf_endpoint>&client_id=<id>&client_secret=<secret>...
```

### Create Data Input
```
POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object
Authorization: Splunk <token>
Content-Type: application/x-www-form-urlencoded

name=<input_name>&account=<account>&object=<sf_object>&interval=<seconds>...
```

### List Data Inputs
```
GET /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json
Authorization: Splunk <token>

Response: { "entry": [{ "name": "<input_name>", ... }] }
```

For detailed cURL examples, see `curlCalls.txt`.

## Troubleshooting

### SSL Certificate Errors

If you encounter SSL certificate verification errors with self-signed certificates:

```json
{
  "SPLUNK_SKIP_SSL_VERIFY": true
}
```

**Security Note**: Only use this in development/testing environments. In production, use proper CA-signed certificates.

### Authentication Failures

**Problem**: `authentication failed with status 401`

**Solutions**:
1. Verify Splunk credentials are correct
2. Check if Splunk instance is accessible: `curl -k https://your-splunk:8089`
3. Ensure Splunk management port (8089) is open and not blocked by firewall
4. Verify the user has `admin` or appropriate REST API access role

### Add-on Not Found

**Problem**: `Splunk Add-on for Salesforce is not installed`

**Solution**:
1. Install the Splunk Add-on for Salesforce from Splunkbase
2. Restart Splunk: `$SPLUNK_HOME/bin/splunk restart`
3. Verify installation: Check "Apps" > "Manage Apps" in Splunk Web UI
4. Ensure add-on is enabled

### Resource Already Exists (409 Conflict)

**Problem**: `resource already exists` errors

**Behavior**: The tool automatically handles 409 conflicts gracefully and continues execution. Existing resources are not modified or recreated.

### Connection Timeout

**Problem**: Requests timing out

**Solutions**:
1. Increase timeout in configuration:
   ```json
   {
     "SPLUNK_REQUEST_TIMEOUT": 60
   }
   ```
2. Check network connectivity to Splunk instance
3. Verify Splunk instance is not overloaded

### Data Input Creation Failures

**Problem**: Some data inputs fail to create

**Diagnostics**:
- Check logs for specific error messages
- Verify Salesforce account credentials are valid
- Ensure Salesforce objects exist and are accessible
- Check field names in `object_fields` are correct

**Retry Strategy**: The tool uses exponential backoff for transient failures. Configure retry behavior:
```json
{
  "SPLUNK_MAX_RETRIES": 5,
  "SPLUNK_RETRY_DELAY": 3
}
```

### Parallel Execution Issues

**Problem**: Too many concurrent requests causing failures

**Solution**: Reduce concurrency:
```json
{
  "MIGRATION_CONCURRENT_REQUESTS": 2
}
```

### Logging and Debugging

Enable debug logging for detailed troubleshooting:

```json
{
  "MIGRATION_LOG_LEVEL": "debug"
}
```

Debug logs include:
- Full HTTP request/response details
- FlowGraph node execution traces
- Retry attempts and backoff calculations
- Connection pool statistics

## FlowGraph Integration

This tool uses FlowGraph for workflow orchestration, providing:

- **Graph-based Execution**: Migration steps as nodes in a directed acyclic graph
- **State Management**: Thread-safe state tracking across nodes
- **Checkpointing**: Ability to resume from specific steps (future enhancement)
- **Metrics Collection**: Built-in instrumentation for performance monitoring
- **Validation**: Graph structure validation before execution
- **Error Handling**: Node-level error propagation with graceful degradation

### Workflow Visualization

```
[authenticate]
     ↓
[check_salesforce_addon]
     ↓
[create_index]
     ↓
[create_account]
     ↓
[load_data_inputs]
     ↓
[create_data_inputs] ← Parallel execution with semaphore
     ↓
[verify_inputs]
     ↓
[create_dashboards] ← Optional, skipped if not configured
```

### Custom Node Processor

The `MigrationNodeProcessor` implements FlowGraph's `NodeProcessor` interface:

```go
type NodeProcessor interface {
    Process(ctx context.Context, node *Node, input map[string]interface{}) (map[string]interface{}, error)
    CanProcess(nodeType NodeType) bool
}
```

Each migration step is a self-contained node with:
- Input/output state passing
- Error handling and recovery
- Logging and observability
- Metrics and timing

## Performance Considerations

### Connection Pooling

The HTTP client maintains a connection pool with:
- `MaxIdleConns`: 100 (total idle connections)
- `MaxConnsPerHost`: 100 (per-host connections)
- Keepalive enabled for connection reuse

### Concurrent Data Input Creation

Controlled by `MIGRATION_CONCURRENT_REQUESTS` (default: 5):
- Too high: May overwhelm Splunk instance
- Too low: Increases total migration time
- Recommended: 3-10 depending on Splunk capacity

### Retry and Backoff

Default configuration:
- Max retries: 3
- Initial delay: 2 seconds
- Backoff multiplier: 2.0
- Max delay: 16 seconds (2 * 2^3)

## Testing

### Unit Tests

The codebase includes comprehensive unit tests with mocking:

```powershell
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific test
go test -v -run TestSplunkService_Authenticate ./services
```

### Test Coverage

```powershell
make coverage
```

Current coverage areas:
- ✅ HTTP client with retry logic
- ✅ Splunk service methods
- ✅ Configuration loading and validation
- ✅ Logger initialization
- ✅ FlowGraph workflow execution
- ✅ Migration node processor

### Integration Tests

Located in `internal/workflows/*_test.go`:
- Full workflow execution with mocks
- Error handling and recovery
- Parallel execution behavior
- State management

## Docker Support

Build and run in a container:

```powershell
# Build Docker image
docker build -t salesforce-splunk-migration .

# Run with mounted config
docker run -v ${PWD}/credentials.json:/app/credentials.json salesforce-splunk-migration
```

The included `Dockerfile` creates a minimal container with the compiled binary.

## License

MIT License - see LICENSE file for details.

## Support and Documentation

- **GitHub Issues**: Report bugs and request features
- **Salesforce to Splunk Migration Guide**: See `Salesforce_to_Splunk_Migration_Guide.md`
- **FlowGraph Documentation**: See `flowgraph/README.md`
- **API Examples**: See `curlCalls.txt`

## Changelog

### Version 1.0.0 (Current)
- ✅ FlowGraph-based workflow orchestration
- ✅ Parallel data input creation with concurrency control
- ✅ Comprehensive error handling and retry logic
- ✅ Connection pooling and keepalive
- ✅ Structured logging with Zap
- ✅ Full test coverage with mocks
- ✅ Add-on verification step
- ✅ Post-migration input verification

## Acknowledgments

- Built with [FlowGraph](https://github.com/flowgraph/flowgraph) for workflow orchestration
- Logging powered by [Zap](https://github.com/uber-go/zap)
- Configuration parsing with [mapstructure](https://github.com/mitchellh/mapstructure)
- Testing with [testify](https://github.com/stretchr/testify)
