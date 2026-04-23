package mcps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MCPS map[string]ServerConfig `yaml:"mcps"`
}

type ServerConfig struct {
	URL   string                `yaml:"url"`
	Tools map[string]ToolConfig `yaml:"tools"`
}

type ToolConfig struct {
	Description string            `yaml:"description,omitempty"`
	Arguments   map[string]string `yaml:"arguments,omitempty"`
	KeepAsIs    bool              `yaml:"keep_as_is,omitempty"`
}

func ConfigPath(rootDir string) string {
	return rootDir + "/.nixdevkit/mcps.yml"
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse mcps config: %w", err)
	}
	return &cfg, nil
}

func RegisterProxiedTools(ctx context.Context, srv *server.MCPServer, cfg *Config) error {
	for name, scfg := range cfg.MCPS {
		if err := registerUpstream(ctx, srv, name, scfg); err != nil {
			return err
		}
	}
	return nil
}

func registerUpstream(ctx context.Context, srv *server.MCPServer, name string, scfg ServerConfig) error {
	c, err := client.NewStreamableHttpClient(scfg.URL)
	if err != nil {
		return fmt.Errorf("connect upstream %s: %w", name, err)
	}
	if err := c.Start(ctx); err != nil {
		return fmt.Errorf("start upstream %s: %w", name, err)
	}
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "nixdevkit-proxy",
				Version: "0.1.0",
			},
		},
	}); err != nil {
		return fmt.Errorf("initialize upstream %s: %w", name, err)
	}

	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("list tools from %s: %w", name, err)
	}

	for _, tool := range toolsResult.Tools {
		tcfg, ok := scfg.Tools[tool.Name]
		if !ok {
			continue
		}

		proxyTool := tool
		if !tcfg.KeepAsIs {
			if tcfg.Description != "" {
				proxyTool.Description = tcfg.Description
			}
			for argName, desc := range tcfg.Arguments {
				setArgDescription(&proxyTool, argName, desc)
			}
		}

		cl := c
		srv.AddTool(proxyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return cl.CallTool(ctx, req)
		})
	}
	return nil
}

func setArgDescription(tool *mcp.Tool, argName, desc string) {
	props := tool.InputSchema.Properties
	if props == nil {
		return
	}
	prop, ok := props[argName]
	if !ok {
		return
	}
	switch p := prop.(type) {
	case map[string]any:
		p["description"] = desc
	default:
		b, err := json.Marshal(prop)
		if err != nil {
			return
		}
		var m map[string]any
		if json.Unmarshal(b, &m) != nil {
			return
		}
		m["description"] = desc
		props[argName] = m
	}
}
