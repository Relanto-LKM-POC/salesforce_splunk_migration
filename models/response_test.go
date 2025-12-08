package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplunkResponse(t *testing.T) {
	t.Run("Success_CreateSplunkResponse", func(t *testing.T) {
		response := SplunkResponse{
			Links: map[string]string{
				"create": "/services/data/indexes",
				"list":   "/services/data/indexes/_all",
			},
			Origin:  "https://splunk.example.com",
			Updated: "2024-01-01T00:00:00Z",
			Entry:   []Entry{},
			Paging: Paging{
				Total:   100,
				PerPage: 30,
				Offset:  0,
			},
			Messages: []Message{},
		}

		assert.NotNil(t, response.Links)
		assert.Equal(t, "https://splunk.example.com", response.Origin)
		assert.Equal(t, 100, response.Paging.Total)
		assert.Equal(t, 30, response.Paging.PerPage)
		assert.Equal(t, 0, response.Paging.Offset)
	})

	t.Run("Success_WithEntries", func(t *testing.T) {
		entry := Entry{
			Name:    "test_index",
			ID:      "idx-12345",
			Updated: "2024-01-01T00:00:00Z",
			Links: map[string]string{
				"self": "/services/data/indexes/test_index",
			},
			Author: "admin",
			Content: map[string]interface{}{
				"datatype": "event",
				"status":   "active",
			},
		}

		response := SplunkResponse{
			Entry: []Entry{entry},
		}

		require.Len(t, response.Entry, 1)
		assert.Equal(t, "test_index", response.Entry[0].Name)
		assert.Equal(t, "idx-12345", response.Entry[0].ID)
		assert.Equal(t, "admin", response.Entry[0].Author)
	})

	t.Run("Success_WithMessages", func(t *testing.T) {
		message := Message{
			Type: "INFO",
			Text: "Index created successfully",
		}

		response := SplunkResponse{
			Messages: []Message{message},
		}

		require.Len(t, response.Messages, 1)
		assert.Equal(t, "INFO", response.Messages[0].Type)
		assert.Equal(t, "Index created successfully", response.Messages[0].Text)
	})

	t.Run("Success_EmptyResponse", func(t *testing.T) {
		response := SplunkResponse{}

		assert.Nil(t, response.Links)
		assert.Empty(t, response.Origin)
		assert.Empty(t, response.Entry)
		assert.Empty(t, response.Messages)
	})
}

func TestEntry(t *testing.T) {
	t.Run("Success_CreateEntry", func(t *testing.T) {
		entry := Entry{
			Name:    "salesforce_data",
			ID:      "idx-67890",
			Updated: "2024-01-15T10:30:00Z",
			Links: map[string]string{
				"self":   "/services/data/indexes/salesforce_data",
				"edit":   "/services/data/indexes/salesforce_data/edit",
				"remove": "/services/data/indexes/salesforce_data/remove",
			},
			Author: "admin",
			Content: map[string]interface{}{
				"datatype":               "event",
				"maxTotalDataSizeMB":     500000,
				"frozenTimePeriodInSecs": 188697600,
			},
		}

		assert.Equal(t, "salesforce_data", entry.Name)
		assert.Equal(t, "idx-67890", entry.ID)
		assert.Equal(t, "admin", entry.Author)
		assert.Contains(t, entry.Links, "self")
		assert.Contains(t, entry.Content, "datatype")
		assert.Equal(t, "event", entry.Content["datatype"])
	})

	t.Run("Success_EntryWithComplexContent", func(t *testing.T) {
		entry := Entry{
			Name: "complex_index",
			Content: map[string]interface{}{
				"nested": map[string]interface{}{
					"level1": "value1",
					"level2": map[string]interface{}{
						"level3": "value3",
					},
				},
				"array": []string{"item1", "item2", "item3"},
			},
		}

		assert.NotNil(t, entry.Content["nested"])
		assert.NotNil(t, entry.Content["array"])
	})

	t.Run("Success_EmptyEntry", func(t *testing.T) {
		entry := Entry{}

		assert.Empty(t, entry.Name)
		assert.Empty(t, entry.ID)
		assert.Nil(t, entry.Links)
		assert.Nil(t, entry.Content)
	})
}

func TestPaging(t *testing.T) {
	t.Run("Success_CreatePaging", func(t *testing.T) {
		paging := Paging{
			Total:   250,
			PerPage: 50,
			Offset:  100,
		}

		assert.Equal(t, 250, paging.Total)
		assert.Equal(t, 50, paging.PerPage)
		assert.Equal(t, 100, paging.Offset)
	})

	t.Run("Success_FirstPage", func(t *testing.T) {
		paging := Paging{
			Total:   100,
			PerPage: 30,
			Offset:  0,
		}

		assert.Equal(t, 0, paging.Offset)
		assert.True(t, paging.Offset == 0)
	})

	t.Run("Success_LastPage", func(t *testing.T) {
		paging := Paging{
			Total:   95,
			PerPage: 30,
			Offset:  90,
		}

		assert.True(t, paging.Offset+paging.PerPage >= paging.Total)
	})

	t.Run("Success_ZeroValues", func(t *testing.T) {
		paging := Paging{}

		assert.Equal(t, 0, paging.Total)
		assert.Equal(t, 0, paging.PerPage)
		assert.Equal(t, 0, paging.Offset)
	})
}

func TestMessage(t *testing.T) {
	t.Run("Success_InfoMessage", func(t *testing.T) {
		message := Message{
			Type: "INFO",
			Text: "Operation completed successfully",
		}

		assert.Equal(t, "INFO", message.Type)
		assert.Equal(t, "Operation completed successfully", message.Text)
	})

	t.Run("Success_ErrorMessage", func(t *testing.T) {
		message := Message{
			Type: "ERROR",
			Text: "Failed to create index: permission denied",
		}

		assert.Equal(t, "ERROR", message.Type)
		assert.Contains(t, message.Text, "Failed to create index")
	})

	t.Run("Success_WarningMessage", func(t *testing.T) {
		message := Message{
			Type: "WARN",
			Text: "Index already exists, using existing index",
		}

		assert.Equal(t, "WARN", message.Type)
		assert.Contains(t, message.Text, "already exists")
	})

	t.Run("Success_EmptyMessage", func(t *testing.T) {
		message := Message{}

		assert.Empty(t, message.Type)
		assert.Empty(t, message.Text)
	})

	t.Run("Success_LongMessage", func(t *testing.T) {
		longText := "This is a very long error message that contains detailed information about what went wrong during the operation and provides context for debugging purposes."
		message := Message{
			Type: "ERROR",
			Text: longText,
		}

		assert.Equal(t, "ERROR", message.Type)
		assert.Equal(t, longText, message.Text)
		assert.True(t, len(message.Text) > 50)
	})
}

func TestAuthResponse(t *testing.T) {
	t.Run("Success_ValidAuthResponse", func(t *testing.T) {
		authResponse := AuthResponse{
			SessionKey: "abcdef123456789",
		}

		assert.Equal(t, "abcdef123456789", authResponse.SessionKey)
		assert.NotEmpty(t, authResponse.SessionKey)
	})

	t.Run("Success_EmptySessionKey", func(t *testing.T) {
		authResponse := AuthResponse{
			SessionKey: "",
		}

		assert.Empty(t, authResponse.SessionKey)
	})

	t.Run("Success_LongSessionKey", func(t *testing.T) {
		longKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"
		authResponse := AuthResponse{
			SessionKey: longKey,
		}

		assert.Equal(t, longKey, authResponse.SessionKey)
		assert.True(t, len(authResponse.SessionKey) > 50)
	})
}
