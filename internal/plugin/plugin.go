package plugin

import (
	"context"
	"net/rpc"
	
	"github.com/hashicorp/go-plugin"
	"github.com/taokim/muno/internal/interfaces"
)

// MunoPlugin is the implementation of plugin.Plugin for MUNO plugins
type MunoPlugin struct {
	Impl interfaces.Plugin
}

// Server returns the RPC server for this plugin
func (p *MunoPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

// Client returns the RPC client for this plugin
func (p *MunoPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// RPCServer is the RPC server that RPCClient talks to
type RPCServer struct {
	Impl interfaces.Plugin
}

// Metadata returns plugin metadata
func (s *RPCServer) Metadata(args interface{}, resp *interfaces.PluginMetadata) error {
	*resp = s.Impl.Metadata()
	return nil
}

// Commands returns plugin commands
func (s *RPCServer) Commands(args interface{}, resp *[]interfaces.CommandDefinition) error {
	*resp = s.Impl.Commands()
	return nil
}

// Execute executes a command
func (s *RPCServer) Execute(args *ExecuteArgs, resp *interfaces.Result) error {
	result, err := s.Impl.Execute(context.Background(), args.Command, args.Args, args.Environment)
	if err != nil {
		return err
	}
	*resp = result
	return nil
}

// Initialize initializes the plugin
func (s *RPCServer) Initialize(config map[string]interface{}, resp *interface{}) error {
	return s.Impl.Initialize(config)
}

// Cleanup performs cleanup
func (s *RPCServer) Cleanup(args interface{}, resp *interface{}) error {
	return s.Impl.Cleanup()
}

// HealthCheck performs health check
func (s *RPCServer) HealthCheck(args interface{}, resp *interface{}) error {
	return s.Impl.HealthCheck(context.Background())
}

// ExecuteArgs holds arguments for Execute RPC
type ExecuteArgs struct {
	Command     string
	Args        []string
	Environment interfaces.PluginEnvironment
}

// RPCClient is an implementation that talks over RPC
type RPCClient struct {
	client *rpc.Client
}

// Metadata returns plugin metadata
func (c *RPCClient) Metadata() interfaces.PluginMetadata {
	var resp interfaces.PluginMetadata
	err := c.client.Call("Plugin.Metadata", new(interface{}), &resp)
	if err != nil {
		return interfaces.PluginMetadata{}
	}
	return resp
}

// Commands returns plugin commands
func (c *RPCClient) Commands() []interfaces.CommandDefinition {
	var resp []interfaces.CommandDefinition
	err := c.client.Call("Plugin.Commands", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return resp
}

// Execute executes a command
func (c *RPCClient) Execute(ctx context.Context, cmd string, args []string, env interfaces.PluginEnvironment) (interfaces.Result, error) {
	var resp interfaces.Result
	err := c.client.Call("Plugin.Execute", &ExecuteArgs{
		Command:     cmd,
		Args:        args,
		Environment: env,
	}, &resp)
	return resp, err
}

// Initialize initializes the plugin
func (c *RPCClient) Initialize(config map[string]interface{}) error {
	var resp interface{}
	return c.client.Call("Plugin.Initialize", config, &resp)
}

// Cleanup performs cleanup
func (c *RPCClient) Cleanup() error {
	var resp interface{}
	return c.client.Call("Plugin.Cleanup", new(interface{}), &resp)
}

// HealthCheck performs health check
func (c *RPCClient) HealthCheck(ctx context.Context) error {
	var resp interface{}
	return c.client.Call("Plugin.HealthCheck", new(interface{}), &resp)
}