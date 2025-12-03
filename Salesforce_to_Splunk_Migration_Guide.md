# Complete Guide: Salesforce to Splunk Migration

## Table of Contents
1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Salesforce Integration User's Account Setup](#salesforce-account-setup)
4. [Installing Splunk Add-on for Salesforce](#installing-splunk-add-on-for-salesforce)
5. [Configuring Salesforce Account in Splunk](#configuring-salesforce-account-in-splunk)
6. [Creating Indexes for Salesforce Data](#creating-indexes-for-salesforce-data)
7. [Creating Data Inputs for Salesforce Objects](#creating-data-inputs-for-salesforce-objects)
8. [Dashboard Creation with Golang Pipeline](#dashboard-creation-with-golang-pipeline)
9. [Automation with REST API](#automation-with-rest-api)
10. [Summary Checklist](#summary-checklist)
11. [Additional Resources](#additional-resources)

---

## Overview

This guide provides comprehensive steps for integrating Salesforce data with Splunk, including:
- Installing and configuring the Splunk Add-on for Salesforce
- Setting up Salesforce accounts with basic authentication
- Creating data inputs for Salesforce objects
- Automating configurations using REST API

**Key Benefits:**
- Real-time data collection from Salesforce to Splunk
- Automated configuration management via REST API
- Support for custom and standard Salesforce objects
- Simple and reliable basic authentication

---

## Prerequisites

### Splunk Requirements
- **Splunk Enterprise** 8.0+ or **Splunk Cloud**
- Access to Splunk management port (8089)
- Admin or sufficient privileges to:
  - Install apps
  - Create data inputs
  - Configure authentication tokens

### Salesforce Requirements
- **Salesforce Account** with:
  - API access enabled
  - Read permissions on objects you want to collect
  - Security token (if not in trusted IP range)

### Network Requirements
- Network connectivity from Splunk to Salesforce API endpoints
- Firewall rules allowing HTTPS (443) to:
  - `login.salesforce.com` (Production)
  - `test.salesforce.com` (Sandbox)
  - Custom domain (if applicable)

---

## Salesforce Account Setup

### Step 1: Create Integration User

1. Log into Salesforce as Administrator
2. Go to **Setup → Users → Users**
3. Click **New User**
4. Fill in required details:
   - **Profile**: System Administrator (or custom profile with API access)
   - **Email**: integration-user@yourcompany.com
   - **Username**: Must be globally unique
5. Ensure the user has:
   - **API Enabled** permission
   - **View All Data** or specific object permissions

### Step 2: Generate Security Token

1. Log in as the integration user
2. Go to **Settings → My Personal Information → Reset My Security Token**
3. Click **Reset Security Token**
4. Check email for security token
5. **Save this token securely** - you'll need it for authentication

### Step 3: Verify Field-Level Security (FLS)

1. Navigate to **Setup → Users → Profiles**
2. Select the integration user's profile
3. Go to **Object Settings** for each object you want to collect
4. Verify **Field-Level Security** for required fields:
   - Ensure "Read" permission is granted
   - Missing FLS permissions cause empty fields in Splunk

### Step 4: Verify Permissions

1. Ensure the integration user has sufficient permissions:
   - **API Enabled** checkbox is checked
   - **Read** access to all required Salesforce objects
2. Verify the user can successfully log in to Salesforce
3. Test API access using Salesforce Developer Console if needed

**Note**: OAuth 2.0 authentication (Client Credentials and Authorization Code flows) is also available as an alternative authentication method. Refer to the official Splunk Add-on for Salesforce documentation for OAuth setup instructions if needed.

---

## Installing Splunk Add-on for Salesforce

### Installation Methods

#### Method 1: Via Splunkbase (Web UI)

1. Log into Splunk Web
2. Go to **Apps → Find More Apps**
3. Search for "Splunk Add-on for Salesforce"
4. Click **Install**
5. Enter your Splunkbase credentials
6. Click **Agree and Install**
7. Restart Splunk when prompted

#### Method 2: Manual Installation

1. Download the add-on from [Splunkbase](https://splunkbase.splunk.com/app/3549)
2. Extract the `.tgz` or `.tar.gz` file
3. Copy to `$SPLUNK_HOME/etc/apps/Splunk_TA_salesforce`
4. Set appropriate permissions:
   ```bash
   chown -R splunk:splunk $SPLUNK_HOME/etc/apps/Splunk_TA_salesforce
   ```
5. Restart Splunk:
   ```bash
   $SPLUNK_HOME/bin/splunk restart
   ```

### Post-Installation Verification

1. Go to **Apps** in Splunk Web
2. Verify "Splunk Add-on for Salesforce" is listed
3. Click **Launch App**
4. Verify you see **Configuration**, **Inputs**, and **Dashboard** tabs

---

## Configuring Salesforce Account in Splunk

### Method1: Using Splunk Web (Recommended)

#### Basic Authentication Setup

1. In Splunk Web, click **Splunk Add-on for Salesforce**
2. Go to **Configuration → Account**
3. Click **Add**
4. Fill in the form:
   - **Account Name**: `salesforce_production` (unique identifier)
   - **Endpoint**: 
     - Production: `login.salesforce.com`
     - Sandbox: `test.salesforce.com`
     - Custom: `yourcompany.my.salesforce.com`
   - **Salesforce API Version**: `64.0` (default, or choose from 42.0-64.0)
   - **Auth Type**: `Basic Authentication`
   - **Username**: Your Salesforce username
   - **Password**: Your Salesforce password
   - **Security Token**: Token from email (leave blank if in trusted IP range)
5. Click **Add**
6. Verify success message

**Note**: OAuth 2.0 authentication methods (Client Credentials and Authorization Code flows) are also supported. For OAuth setup, refer to the official documentation.

### Method 2: Using REST API

You can create a Salesforce account configuration using the REST API:

```bash
curl -k -X POST \
  "https://YOUR_SPLUNK_URL:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" \
  -H "Authorization: Splunk YOUR_SPLUNK_AUTH_TOKEN" \
  -d "name=salesforce_UAT_curl" \
  -d "endpoint=https://login.salesforce.com" \
  -d "api_version=64.0" \
  -d "auth_type=basic" \
  -d "username=YOUR_SALESFORCE_USERNAME" \
  -d "password=YOUR_SALESFORCE_PASSWORD" \
  -d "token=YOUR_SALESFORCE_SECURITY_TOKEN"
```

**Parameters:**
- `name`: Unique identifier for this account configuration
- `endpoint`: Salesforce login endpoint
  - Production: `https://login.salesforce.com`
  - Sandbox: `https://test.salesforce.com`
- `api_version`: Salesforce API version (42.0-64.0)
- `auth_type`: Authentication type (`basic` for basic authentication)
- `username`: Your Salesforce username
- `password`: Your Salesforce password
- `token`: Your Salesforce security token (if required)

---

## Creating Indexes for Salesforce Data

Before creating data inputs, it's recommended to create dedicated indexes for your Salesforce data. This provides better data organization, access control, and retention management.

### Method 1: Using Splunk Web (Recommended)

1. Log into Splunk Web
2. Go to **Settings → Indexes**
3. Click **New Index** (top right)
4. Fill in the form:

| Field | Description | Example |
|-------|-------------|----------|
| **Index Name** | Unique identifier for the index (lowercase, no spaces) | `salesforce` or `sf_production` |
| **Index Data Type** | Type of data to store | `Events` (default) |
| **Max Size of Entire Index** | Maximum total size for the index | `500000` MB (500 GB) or `auto` |
| **App** | Associate with an app (optional) | `Splunk_TA_salesforce` |
| **Home Path** | Location for hot/warm buckets | `$SPLUNK_DB/salesforce/db` (default) |
| **Cold Path** | Location for cold buckets | `$SPLUNK_DB/salesforce/colddb` (default) |
| **Thawed Path** | Location for thawed/restored buckets | `$SPLUNK_DB/salesforce/thaweddb` (default) |
| **Max Data Size** | Size before rolling to warm | `auto` (recommended) |
| **Frozen Path** | Archive location for aged data | Leave empty or specify archive path |

5. Configure retention settings (optional):
   - **Max Hot Buckets**: Number of hot buckets (default: `3`)
   - **Max Warm Buckets**: Number of warm buckets before rolling to cold
   - **Frozen Time Period In Secs**: Time before archiving data (default: `188697600` = 6 years)

6. Click **Save**
7. Verify the index appears in the list

**Recommended Index Names:**
- `salesforce` - General Salesforce data
- `salesforce_objects` - Salesforce object data
- `salesforce_production` - Production environment
- `salesforce_sandbox` - Sandbox/UAT environment

### Method 2: Using REST API

You can create indexes programmatically using the Splunk REST API:

```bash
curl -k -X POST \
  "https://YOUR_SPLUNK_URL:8089/services/data/indexes" \
  -H "Authorization: Splunk YOUR_SPLUNK_AUTH_TOKEN" \
  -d "name=salesforce" \
  -d "datatype=event" \
  -d "maxTotalDataSizeMB=500000"
```

**Parameters:**
- `name`: Index name (lowercase, no spaces)
- `datatype`: Type of index (`event` for events, `metric` for metrics)
- `maxTotalDataSizeMB`: Maximum size in megabytes (optional)
- `frozenTimePeriodInSecs`: Retention period in seconds (optional, default: 188697600)
- `homePath`: Custom home path (optional, uses default if not specified)
- `coldPath`: Custom cold path (optional, uses default if not specified)
- `thawedPath`: Custom thawed path (optional, uses default if not specified)

**Example: Create multiple indexes for different environments**

```bash
#!/bin/bash

SPLUNK_URL="https://YOUR_SPLUNK_URL:8089"
SPLUNK_TOKEN="YOUR_SPLUNK_AUTH_TOKEN"

# Create production index
echo "Creating salesforce_production index..."
curl -k -X POST \
  "${SPLUNK_URL}/services/data/indexes" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "name=salesforce_production" \
  -d "datatype=event" \
  -d "maxTotalDataSizeMB=1000000"

# Create sandbox index
echo "Creating salesforce_sandbox index..."
curl -k -X POST \
  "${SPLUNK_URL}/services/data/indexes" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "name=salesforce_sandbox" \
  -d "datatype=event" \
  -d "maxTotalDataSizeMB=500000"

echo "Indexes created successfully!"
```

### Verify Index Creation

**Via Splunk Web:**
1. Go to **Settings → Indexes**
2. Verify your new index is listed
3. Check the status shows as "Enabled"

**Via REST API:**
```bash
curl -k -X GET \
  "${SPLUNK_URL}/services/data/indexes/salesforce?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}"
```

### Best Practices for Index Configuration

#### 1. Naming Convention
- Use lowercase names with underscores
- Include environment identifier (e.g., `salesforce_prod`, `salesforce_uat`)
- Keep names descriptive but concise

#### 2. Size Planning
- Estimate daily data volume from Salesforce
- Plan for 30-90 days of hot/warm data
- Configure cold storage for long-term retention
- Monitor index size regularly

#### 3. Retention Policy
- Align with compliance requirements
- Default: 6 years (188697600 seconds)
- Configure frozen path for archival if needed
- Document retention policies

#### 4. Performance Optimization
- Use separate indexes for different data sources or environments
- Configure appropriate bucket sizes
- Monitor bucket count and distribution
- Use SSD for hot buckets when possible

#### 5. Access Control
- Configure role-based access to indexes
- Restrict write access to data collection tier
- Grant search permissions to appropriate roles
- Audit index access regularly

### Common Index Configuration Examples

**Development/Testing:**
```bash
name=salesforce_dev
maxTotalDataSizeMB=100000
frozenTimePeriodInSecs=2592000  # 30 days
```

**Production:**
```bash
name=salesforce_production
maxTotalDataSizeMB=2000000
frozenTimePeriodInSecs=188697600  # 6 years
```

**Compliance (Extended Retention):**
```bash
name=salesforce_compliance
maxTotalDataSizeMB=3000000
frozenTimePeriodInSecs=315360000  # 10 years
```

---

## Creating Data Inputs for Salesforce Objects

### Available Default Inputs

The add-on provides 7 pre-configured inputs for common objects:
- **account** - Account object
- **dashboard** - Dashboard metadata
- **loginhistory** - User login history
- **opportunity** - Opportunity object
- **report** - Report metadata
- **user** - User object
- **case** - Case object

**Note**: To use with [Splunk App for Salesforce](https://splunkbase.splunk.com/app/1931/), enable: account, dashboard, loginhistory, opportunity, report, and user.

### Method 1: Using Splunk Web (Recommended)

1. In Splunk Add-on for Salesforce, go to **Inputs**
2. Click **Create New Input**
3. Select **Salesforce Object**
4. Fill in the form:

| Field | Description | Example |
|-------|-------------|---------|
| **Name** | Unique identifier for this input | `sf_accounts_production` |
| **Salesforce Account** | Account name configured earlier | `salesforce_production` |
| **Object** | Salesforce object API name | `Account` |
| **Object Fields** | Comma-separated field list | `Id,Name,Type,Industry,BillingCountry,AnnualRevenue,CreatedDate,LastModifiedDate` |
| **Order By** | Field to order results (usually datetime) | `LastModifiedDate` |
| **Query Start Date** | Start date for data collection | `2024-01-01T00:00:00.000Z` |
| **Limit** | Max records per query | `1000` |
| **Interval** | Seconds between queries | `1200` (20 minutes) |
| **Delay** | Delay collection to account for Salesforce processing lag | `60` (1 minute) |
| **Index** | Target Splunk index | `salesforce` or `default` |
| **Sourcetype** | (Auto-generated) | `salesforce:object:account` |

5. Click **Add**

### Common Salesforce Objects

#### Standard Objects

**Account:**
```
Fields: Id,Name,Type,Industry,BillingCountry,BillingCity,AnnualRevenue,NumberOfEmployees,CreatedDate,LastModifiedDate
```

**Contact:**
```
Fields: Id,FirstName,LastName,Email,AccountId,Department,Title,MobilePhone,CreatedDate,LastModifiedDate
```

**Opportunity:**
```
Fields: Id,Name,AccountId,StageName,Amount,Probability,CloseDate,Type,LeadSource,CreatedDate,LastModifiedDate
```

**Case:**
```
Fields: Id,CaseNumber,Subject,Status,Priority,Origin,AccountId,ContactId,CreatedDate,LastModifiedDate,ClosedDate
```

**Lead:**
```
Fields: Id,FirstName,LastName,Company,Email,Status,LeadSource,Industry,Rating,CreatedDate,LastModifiedDate
```

**User:**
```
Fields: Id,Username,Email,FirstName,LastName,IsActive,UserRoleId,ProfileId,CreatedDate,LastModifiedDate
```

**Task:**
```
Fields: Id,Subject,Status,Priority,ActivityDate,WhoId,WhatId,OwnerId,CreatedDate,LastModifiedDate
```

**Event:**
```
Fields: Id,Subject,StartDateTime,EndDateTime,WhoId,WhatId,OwnerId,CreatedDate,LastModifiedDate
```

#### Custom Objects

Custom objects use `__c` suffix:
```
Object: MyCustomObject__c
Fields: Id,Name,Custom_Field_1__c,Custom_Field_2__c,CreatedDate,LastModifiedDate
```

### Best Practices for Object Inputs

#### 1. Field Selection
- Only select fields you need (reduces API calls and data volume)
- Always include: `Id`, `CreatedDate`, `LastModifiedDate`
- Check field-level security permissions in Salesforce

#### 2. Order By Field
- Use `LastModifiedDate` for incremental updates
- **Known Issue**: May skip records in some cases
- **Workaround**: Create two inputs with different order fields:
  - Input 1: Order by `LastModifiedDate`
  - Input 2: Order by `CreatedDate`
  - Use `dedup` in Splunk searches

#### 3. Interval Configuration
- Adjust based on data volume and update frequency
- Recommended intervals:
  - High-frequency updates: 300-600 seconds (5-10 min)
  - Normal updates: 1200-3600 seconds (20-60 min)
  - Low-frequency: 7200+ seconds (2+ hours)

#### 4. Delay Parameter
- Set delay to account for Salesforce processing lag
- Recommended: 30-60 seconds
- Prevents missing recently created/modified records

#### 5. Start Date
- For initial setup: Start 90 days back (default)
- For backfilling: Set to earliest required date
- Format: `YYYY-MM-DDThh:mm:ss.000Z` (UTC)

#### 6. Checkpoint Management
- From version 2.0.0+, add-on prompts to use existing checkpoints
- **Yes**: Continue from last checkpoint
- **No**: Reset and start from Query Start Date

### Method 2: Using REST API

You can create Salesforce object inputs using the REST API:

```bash
curl -k -X POST \
  "https://YOUR_SPLUNK_URL:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" \
  -H "Authorization: Splunk YOUR_SPLUNK_AUTH_TOKEN" \
  -d "name=sf_accounts" \
  -d "account=salesforce_UAT_curl" \
  -d "object=Account" \
  -d "object_fields=Id,Name,Type,Industry,BillingCountry,AnnualRevenue,CreatedDate,LastModifiedDate" \
  -d "order_by=LastModifiedDate" \
  -d "start_date=2024-01-01T00:00:00.000Z" \
  -d "limit=1000" \
  -d "interval=1200" \
  -d "delay=60" \
  -d "index=salesforce"
```

**Parameters:**
- `name`: Unique identifier for this object input
- `account`: Salesforce account name (from account configuration)
- `object`: Salesforce object API name (e.g., `Account`, `Contact`, `Opportunity`)
- `object_fields`: Comma-separated list of fields to retrieve
- `order_by`: Field to order results (typically `LastModifiedDate`)
- `start_date`: Start date for data collection (ISO 8601 format in UTC)
- `limit`: Maximum records per query (recommended: 1000)
- `interval`: Collection interval in seconds (e.g., `1200` for 20 minutes)
- `delay`: Delay in seconds to account for Salesforce processing lag
- `index`: Target Splunk index (e.g., `default`, `salesforce`)

---

## Dashboard Creation with Golang Pipeline

### Overview

After successfully configuring data inputs and collecting Salesforce data in Splunk, you can create dashboards using the standalone Golang pipeline. This pipeline automates the creation of Salesforce dashboards in Splunk.


---

## Automation with REST API

### Overview

You can automate the configuration of Salesforce Add-on for Salesforce using the Splunk REST API. This allows you to programmatically create accounts and data inputs without manual configuration through the Splunk Web UI.

**Note**: For infrastructure-as-code approaches, Terraform can also be considered for automation. Refer to the Terraform configuration files in the `terraform/` directory if available.

### Prerequisites

- **Splunk REST API access** (port 8089)
- **Splunk authentication token** or admin credentials
- **curl** or similar HTTP client
- **Splunk Add-on for Salesforce** already installed

### Generate Splunk Auth Token

Before using the REST API, generate an authentication token:

```bash
curl -k -u admin:your_password https://YOUR_SPLUNK_URL:8089/services/authorization/tokens?output_mode=json \
  -d name=salesforce_automation \
  -d audience=automation
```

Save the token from the response for use in subsequent API calls.

### Create Salesforce Account

Use curl to create a Salesforce account configuration:

```bash
curl -k -X POST \
  "https://YOUR_SPLUNK_URL:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" \
  -H "Authorization: Splunk YOUR_SPLUNK_AUTH_TOKEN" \
  -d "name=salesforce_UAT_curl" \
  -d "endpoint=login.salesforce.com" \
  -d "sfdc_api_version=64.0" \
  -d "auth_type=basic" \
  -d "username=your_salesforce_user@company.com" \
  -d "password=YourSalesforcePassword" \
  -d "token=YourSalesforceSecurityToken"
```

**Parameters:**
- `name`: Unique identifier for this account (e.g., `salesforce_UAT_curl`)
- `endpoint`: Salesforce endpoint (`login.salesforce.com` for production, `test.salesforce.com` for sandbox)
- `sfdc_api_version`: API version (e.g., `64.0`)
- `auth_type`: Authentication type (`basic` for username/password/token)
- `username`: Your Salesforce username
- `password`: Your Salesforce password
- `token`: Your Salesforce security token (if not in trusted IP range)

### Create Salesforce Object Input

After creating an account, create data inputs for Salesforce objects:

```bash
curl -k -X POST \
  "https://YOUR_SPLUNK_URL:8089/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" \
  -H "Authorization: Splunk YOUR_SPLUNK_AUTH_TOKEN" \
  -d "name=my_salesforce_input_testing_curl" \
  -d "account=salesforce_UAT_curl" \
  -d "object=Account" \
  -d "object_fields=Id,Name,CreatedDate,LastModifiedDate" \
  -d "order_by=LastModifiedDate" \
  -d "start_date=2024-01-01T00:00:00.000Z" \
  -d "interval=300" \
  -d "delay=60" \
  -d "index=default"
```

**Parameters:**
- `name`: Unique identifier for this input (e.g., `my_salesforce_input_testing_curl`)
- `account`: Salesforce account name created earlier (e.g., `salesforce_UAT_curl`)
- `object`: Salesforce object API name (e.g., `Account`, `Contact`, `Opportunity`)
- `object_fields`: Comma-separated list of fields to collect
- `order_by`: Field to order results (typically `LastModifiedDate`)
- `start_date`: Start date for data collection (format: `YYYY-MM-DDThh:mm:ss.000Z`)
- `interval`: Collection interval in seconds (e.g., `300` for 5 minutes)
- `delay`: Delay in seconds to account for Salesforce processing lag (e.g., `60`)
- `index`: Target Splunk index (e.g., `default`, `salesforce`)

### Example: Complete Setup Script

Here's a complete bash script to automate the entire setup:

```bash
#!/bin/bash

# Configuration
SPLUNK_URL="https://YOUR_SPLUNK_URL:8089"
SPLUNK_TOKEN="YOUR_SPLUNK_AUTH_TOKEN"
SF_ENDPOINT="login.salesforce.com"
SF_USERNAME="your_salesforce_user@company.com"
SF_PASSWORD="YourSalesforcePassword"
SF_TOKEN="YourSalesforceSecurityToken"
ACCOUNT_NAME="salesforce_production"

echo "Creating Salesforce account..."
curl -k -X POST \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "name=${ACCOUNT_NAME}" \
  -d "endpoint=${SF_ENDPOINT}" \
  -d "sfdc_api_version=64.0" \
  -d "auth_type=basic" \
  -d "username=${SF_USERNAME}" \
  -d "password=${SF_PASSWORD}" \
  -d "token=${SF_TOKEN}"

echo ""
echo "Creating Account object input..."
curl -k -X POST \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "name=sf_accounts" \
  -d "account=${ACCOUNT_NAME}" \
  -d "object=Account" \
  -d "object_fields=Id,Name,Type,Industry,BillingCountry,AnnualRevenue,CreatedDate,LastModifiedDate" \
  -d "order_by=LastModifiedDate" \
  -d "start_date=2024-01-01T00:00:00.000Z" \
  -d "interval=1200" \
  -d "delay=60" \
  -d "index=salesforce"

echo ""
echo "Creating Contact object input..."
curl -k -X POST \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "name=sf_contacts" \
  -d "account=${ACCOUNT_NAME}" \
  -d "object=Contact" \
  -d "object_fields=Id,FirstName,LastName,Email,AccountId,CreatedDate,LastModifiedDate" \
  -d "order_by=LastModifiedDate" \
  -d "start_date=2024-01-01T00:00:00.000Z" \
  -d "interval=1200" \
  -d "delay=60" \
  -d "index=salesforce"

echo ""
echo "Setup complete!"
```

### Verify Configuration

After running the automation, verify the configuration:

1. **List all accounts:**
   ```bash
   curl -k -X GET \
     "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account?output_mode=json" \
     -H "Authorization: Splunk ${SPLUNK_TOKEN}"
   ```

2. **List all object inputs:**
   ```bash
   curl -k -X GET \
     "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json" \
     -H "Authorization: Splunk ${SPLUNK_TOKEN}"
   ```

### Update Existing Configuration

To update an existing account or input, use POST with the resource name:

```bash
# Update an account
curl -k -X POST \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account/${ACCOUNT_NAME}?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "password=NewPassword" \
  -d "token=NewSecurityToken"

# Update an input
curl -k -X POST \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object/sf_accounts?output_mode=json" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
  -d "interval=3600"
```

### Delete Configuration

To delete an account or input:

```bash
# Delete an account
curl -k -X DELETE \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account/${ACCOUNT_NAME}" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}"

# Delete an input
curl -k -X DELETE \
  "${SPLUNK_URL}/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object/sf_accounts" \
  -H "Authorization: Splunk ${SPLUNK_TOKEN}"
```

### Best Practices

1. **Use Auth Tokens**: Prefer authentication tokens over basic auth for better security
2. **Secure Credentials**: Store credentials in environment variables or secure vaults
3. **Error Handling**: Always check API response codes and handle errors appropriately
4. **Idempotency**: Check if resources exist before creating to avoid duplicates
5. **Version Control**: Store automation scripts in version control
6. **Testing**: Test scripts in a non-production environment first
7. **Logging**: Log API responses for troubleshooting

### Alternative Automation Tools

While REST API provides direct control, consider these alternatives for larger deployments:

- **Terraform**: Infrastructure-as-code approach with state management (see `terraform/` directory)

---

## Summary Checklist

### Phase 1: Prerequisites
- [ ] Salesforce account with API access
- [ ] Security token generated
- [ ] Splunk Enterprise/Cloud installed
- [ ] Heavy Forwarder or IDM configured
- [ ] Network connectivity verified

### Phase 2: Salesforce Configuration
- [ ] Integration user created
- [ ] Field-level security configured
- [ ] OAuth app created (if using OAuth)
- [ ] Permissions verified

### Phase 3: Splunk Add-on Installation
- [ ] Add-on downloaded from Splunkbase
- [ ] Add-on installed on appropriate tiers
- [ ] Add-on configured and restarted
- [ ] Inputs disabled on search heads

### Phase 4: Account Configuration
- [ ] Salesforce account configured in Splunk
- [ ] Authentication method selected
- [ ] Credentials entered and tested

### Phase 5: Data Input Configuration
- [ ] Object inputs created for required objects
- [ ] Fields selected and verified
- [ ] Start dates configured
- [ ] Intervals optimized
- [ ] Indexes created and assigned

### Phase 6: Automation (Optional)
- [ ] REST API scripts created
- [ ] Authentication token generated
- [ ] Automation scripts tested
- [ ] Configuration verified via API

### Phase 7: Verification
- [ ] Data appearing in Splunk
- [ ] Fields populated correctly
- [ ] No errors in internal logs
- [ ] Dashboards displaying data
- [ ] Performance acceptable

### Phase 8: Monitoring
- [ ] Set up alerts for failures
- [ ] Monitor API usage
- [ ] Check data quality regularly
- [ ] Review internal logs
- [ ] Update configurations as needed

---

## Additional Resources

### Official Documentation
- [Splunk Add-on for Salesforce Documentation](https://splunk.github.io/splunk-add-on-for-salesforce/)
- [Salesforce API Documentation](https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/)
- [Splunk REST API Reference](https://docs.splunk.com/Documentation/Splunk/latest/RESTREF/RESTprolog)

### Community Resources
- [Splunk Community](https://community.splunk.com/)
- [Splunk Answers](https://community.splunk.com/t5/Splunk-Answers/ct-p/en-us-splunk-answers)
- [GitHub - Splunk Add-on for Salesforce](https://github.com/splunk/splunk-add-on-for-salesforce)

### Related Splunk Apps
- [Splunk App for Salesforce](https://splunkbase.splunk.com/app/1931/)
- [Splunk Add-on for Salesforce Analytics](https://splunkbase.splunk.com/app/3612/)

---

**Document Version**: 2.0  
**Last Updated**: December 2, 2025  
**Maintained By**: DevOps Team
