package internal

import (
	"github.com/orchestra-mcp/plugin-ai-screenshot/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ToolsPlugin registers all screenshot tools.
type ToolsPlugin struct{}

// RegisterTools registers all 6 screenshot tools with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	builder.RegisterTool("capture_screen",
		"Capture a full-screen screenshot",
		tools.CaptureScreenSchema(), tools.CaptureScreen())

	builder.RegisterTool("capture_region",
		"Capture a specific region of the screen",
		tools.CaptureRegionSchema(), tools.CaptureRegion())

	builder.RegisterTool("capture_window",
		"Capture a specific window by title",
		tools.CaptureWindowSchema(), tools.CaptureWindow())

	builder.RegisterTool("capture_interactive",
		"Capture an interactive selection (user selects region)",
		tools.CaptureInteractiveSchema(), tools.CaptureInteractive())

	builder.RegisterTool("annotate_screenshot",
		"Annotate a screenshot with overlays (rectangles, arrows, text)",
		tools.AnnotateScreenshotSchema(), tools.AnnotateScreenshot())

	builder.RegisterTool("list_captures",
		"List recent screenshot captures in a directory",
		tools.ListCapturesSchema(), tools.ListCaptures())
}
