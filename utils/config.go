//secrets parsing and setting it so that its accessible throughout the codebase

//take whatevers is recqd and delete if not necessary for the envs loading
package utils

import (
	"context"
	"os"

	"github.com/cx-learning-platform/gen-ai-agent-hub/agent-sdk/pkg/agent"
	"github.com/cx-learning-platform/gen-ai-agent-hub/core-library/foundation/types"
	"github.com/cx-learning-platform/gen-ai-agent-hub/core-library/platform/observability"
	"github.com/cx-learning-platform/gen-ai-agent-hub/core-library/toolkit/tools"
)

// StandardConfig provides unified configuration for all agents
type StandardConfig struct {
	Server     ServerConfig           `json:"server"`
	Agent      AgentConfig            `json:"agent"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `env:"host" json:"host"`
	Port string `env:"APP_PORT" json:"port"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	ClientID         string            `env:"LCP_CLIENT_ID" json:"client_id"`
	ClientSecret     string            `env:"LCP_CLIENT_SECRET" json:"client_secret"`
	TokenURL         string            `env:"LCP_TOKEN_URL" json:"token_url"`
	ContextAgentBase string            `env:"CONTEXT_AGENT_BASE" json:"context_agent_base,omitempty"`
	Environment      agent.Environment `env:"LCP_ENVIRONMENT" json:"environment"`
	Temperature      float64           `json:"temperature"`
	ServiceName      string            `json:"service_name"`
}

// LoadStandardConfig loads configuration using the core library's environment loader
func LoadStandardConfig(logger observability.Logger) (*StandardConfig, error) {
	setup := tools.EnvLoaderTool()

	tool, err := setup.Execute(context.Background(), map[string]interface{}{
		"credentialsPath": os.Getenv("VAULT_PATH"),
	})
	if err != nil {
		return nil, err
	}

	config := &StandardConfig{}
	if err := tool.(*tools.Loader).Load(config); err != nil {
		return nil, err
	}

	// Set defaults if not provided
	if config.Agent.Temperature == 0 {
		config.Agent.Temperature = 0.7
	}

	return config, nil
}

// GetHTTPConfig returns HTTP configuration for the server
func (c *StandardConfig) GetHTTPConfig() *types.HTTPConfig {
	return &types.HTTPConfig{
		Port: c.Server.Port,
		Host: c.Server.Host,
	}
}

// GetExtension safely retrieves an extension value
func (c *StandardConfig) GetExtension(key string) (interface{}, bool) {
	if c.Extensions == nil {
		return nil, false
	}
	val, exists := c.Extensions[key]
	return val, exists
}

// SetExtension safely sets an extension value
func (c *StandardConfig) SetExtension(key string, value interface{}) {
	if c.Extensions == nil {
		c.Extensions = make(map[string]interface{})
	}
	c.Extensions[key] = value
}
