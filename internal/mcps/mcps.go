package mcps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MCPS map[string]ServerConfig `yaml:"mcps"`
}

type ServerConfig struct {
	URL     string                `yaml:"url"`
	Headers map[string]string     `yaml:"headers,omitempty"`
	Tools   map[string]ToolConfig `yaml:"tools"`
}

type ToolConfig struct {
	Rename      string                   `yaml:"rename,omitempty"`
	Description string                   `yaml:"description,omitempty"`
	Arguments   map[string]ArgumentConfig `yaml:"arguments,omitempty"`
	KeepAsIs    bool                     `yaml:"keep_as_is,omitempty"`
}

type ArgumentConfig struct {
	Rename      string `yaml:"rename,omitempty"`
	Description string `yaml:"description,omitempty"`
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
	var opts []transport.StreamableHTTPCOption
	if len(scfg.Headers) > 0 {
		opts = append(opts, transport.WithHTTPHeaders(scfg.Headers))
	}
	c, err := client.NewStreamableHttpClient(scfg.URL, opts...)
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
		upstreamName := tool.Name
		renameMap := map[string]string{}

		if !tcfg.KeepAsIs {
			if tcfg.Rename != "" {
				proxyTool.Name = tcfg.Rename
			}
			if tcfg.Description != "" {
				proxyTool.Description = tcfg.Description
			}
			for argName, acfg := range tcfg.Arguments {
				if acfg.Description != "" {
					setArgDescription(&proxyTool, argName, acfg.Description)
				}
				if acfg.Rename != "" {
					renameMap[argName] = acfg.Rename
				}
			}
			if len(renameMap) > 0 {
				renameArgs(&proxyTool, renameMap)
			}
		}

		cl := c
		srv.AddTool(proxyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			req.Params.Name = upstreamName
			reverseRenameArgs(&req, renameMap)
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

func renameArgs(tool *mcp.Tool, renameMap map[string]string) {
	props := tool.InputSchema.Properties
	if props == nil {
		return
	}
	for oldName, newName := range renameMap {
		prop, ok := props[oldName]
		if !ok {
			continue
		}
		delete(props, oldName)
		props[newName] = prop
	}
	for i, r := range tool.InputSchema.Required {
		if newName, ok := renameMap[r]; ok {
			tool.InputSchema.Required[i] = newName
		}
	}
}

func reverseRenameArgs(req *mcp.CallToolRequest, renameMap map[string]string) {
	if req.Params.Arguments == nil || len(renameMap) == 0 {
		return
	}
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return
	}
	for oldName, newName := range renameMap {
		if v, exists := args[newName]; exists {
			delete(args, newName)
			args[oldName] = v
		}
	}
}
