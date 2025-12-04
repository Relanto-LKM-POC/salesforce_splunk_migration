// Package models contains request/response data structures
package models

// SplunkResponse represents a generic Splunk API response
type SplunkResponse struct {
	Links    map[string]string `json:"links"`
	Origin   string            `json:"origin"`
	Updated  string            `json:"updated"`
	Entry    []Entry           `json:"entry"`
	Paging   Paging            `json:"paging"`
	Messages []Message         `json:"messages"`
}

// Entry represents an entry in the Splunk response
type Entry struct {
	Name    string                 `json:"name"`
	ID      string                 `json:"id"`
	Updated string                 `json:"updated"`
	Links   map[string]string      `json:"links"`
	Author  string                 `json:"author"`
	Content map[string]interface{} `json:"content"`
}

// Paging represents pagination information
type Paging struct {
	Total   int `json:"total"`
	PerPage int `json:"perPage"`
	Offset  int `json:"offset"`
}

// Message represents a message in the Splunk response
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AuthResponse struct {
	SessionKey string `json:"sessionKey"`
}
