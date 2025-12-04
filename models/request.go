// Package models contains request/response data structures
package models

// AuthRequest represents a Splunk authentication request
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IndexRequest represents a Splunk index creation request
type IndexRequest struct {
	Name     string `json:"name"`
	DataType string `json:"datatype"`
}

// SalesforceAccountRequest represents a Salesforce account creation request
type SalesforceAccountRequest struct {
	Name                         string `json:"name"`
	Endpoint                     string `json:"endpoint"`
	SFDCAPIVersion               string `json:"sfdc_api_version"`
	AuthType                     string `json:"auth_type"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	Token                        string `json:"token,omitempty"`
	ClientIDOAuthCredentials     string `json:"client_id_oauth_credentials,omitempty"`
	ClientSecretOAuthCredentials string `json:"client_secret_oauth_credentials,omitempty"`
}

// DataInputRequest represents a Salesforce object data input request
type DataInputRequest struct {
	Name         string `json:"name"`
	Account      string `json:"account"`
	Object       string `json:"object"`
	ObjectFields string `json:"object_fields"`
	OrderBy      string `json:"order_by"`
	StartDate    string `json:"start_date"`
	Interval     int    `json:"interval"`
	Delay        int    `json:"delay"`
	Index        string `json:"index"`
}
