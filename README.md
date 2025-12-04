# Salesforce to Splunk Migration Tool

A Golang CLI application that automates Salesforce-to-Splunk data migration via REST APIs.

## Features

- ✅ Splunk authentication (token and basic auth)
- ✅ Splunk index creation
- ✅ Salesforce account configuration in Splunk
- ✅ Salesforce object data input setup
- ✅ Dry-run mode for testing
- ✅ Retry logic with exponential backoff
- ✅ SSL certificate bypass for self-signed certs

## Prerequisites

### Manual Setup Required
1. **Splunk Add-on for Salesforce** must be pre-installed on your Splunk instance
2. Splunk instance must be running and accessible
3. Valid Salesforce credentials with API access

## Installation

1. Clone the repository:
```powershell
git clone <repository-url>
cd salesforce_splunk_migration
```

2. Install dependencies:
```powershell
go mod download
```

3. Copy and configure credentials:
```powershell
cp .env.example .env
# Edit .env with your actual credentials
```

Or use `credentials.json`:
```powershell
# Edit credentials.json with your configuration
```

## Configuration

The application reads configuration from `credentials.json`. See the file for all available options.

### Required Fields

**Splunk:**
- `url`: Splunk REST API URL (e.g., `https://localhost:8089`)
- `username` + `password` OR `auth_token`

**Salesforce:**
- `endpoint`: Salesforce instance URL
- `username`: Salesforce username
- `password`: Salesforce password
- `api_version`: API version (default: 64.0)
- `auth_type`: `oauth_client_credentials` or `basic`
- `client_id` + `client_secret`: For OAuth authentication

## Usage

### Build the Application

```powershell
make build
```

### Run Full Migration

```powershell
./salesforce-splunk-migration.exe migrate --config credentials.json
```

### Validate Configuration

```powershell
./salesforce-splunk-migration.exe validate --config credentials.json
```

### Dry Run (No Changes Made)

```powershell
./salesforce-splunk-migration.exe migrate --config credentials.json --dry-run
```

### Individual Operations

**Create Indexes Only:**
```powershell
./salesforce-splunk-migration.exe create-index --config credentials.json
```

**Create Salesforce Account:**
```powershell
./salesforce-splunk-migration.exe create-account --config credentials.json
```

**Create Data Inputs:**
```powershell
./salesforce-splunk-migration.exe create-input --config credentials.json
```

## Migration Workflow

The migration follows these steps:

1. **Authentication** - Authenticate with Splunk REST API
2. **Index Creation** - Create required Splunk indexes
3. **Account Setup** - Configure Salesforce account in Splunk Add-on
4. **Data Input Setup** - Create Salesforce object data inputs
5. **Verification** - Verify all operations completed successfully

## Development

### Run Tests

```powershell
make test
```

### Run with Coverage

```powershell
make coverage
```

### Format Code

```powershell
make fmt
```

## Architecture

```
├── cmd/              # CLI commands (Cobra)
├── services/         # Business logic (Splunk API client)
├── models/           # Data structures (requests/responses)
├── utils/            # Configuration loading
├── helpers/          # Utility functions (logging, retry)
├── resources/        # Dashboard XMLs and other resources
└── main.go           # Application entry point
```

## API Reference

All Splunk API calls follow the patterns documented in `curlCalls.txt`:

- **Authentication:** `POST /services/auth/login`
- **Create Index:** `POST /services/data/indexes`
- **Create Account:** `POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account`
- **Create Data Input:** `POST /servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object`

## Troubleshooting

### SSL Certificate Errors

Set `skip_ssl_verify: true` in configuration for self-signed certificates.

### Authentication Failures

- Verify Splunk credentials are correct
- Check if Splunk instance is accessible
- Ensure Splunk REST API port (8089) is open

### Resource Already Exists (409)

The tool will skip existing resources automatically. Use `--dry-run` to preview operations.

## License

MIT

## Support

For issues and questions, please refer to the project documentation or create an issue in the repository.
