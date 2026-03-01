package aiscreenshot

import (
	"github.com/orchestra-mcp/plugin-ai-screenshot/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all screenshot tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)
}
