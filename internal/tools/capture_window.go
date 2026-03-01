package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-ai-screenshot/internal/capture"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// CaptureWindowSchema returns the JSON Schema for the capture_window tool.
func CaptureWindowSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"window_title": map[string]any{
				"type":        "string",
				"description": "Title (or partial title) of the window to capture",
			},
			"output_path": map[string]any{
				"type":        "string",
				"description": "File path to save the screenshot (defaults to temp file)",
			},
		},
		"required": []any{"window_title"},
	})
	return s
}

// CaptureWindow returns a tool handler that captures a specific window by title.
func CaptureWindow() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "window_title"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		title := helpers.GetString(req.Arguments, "window_title")

		outputPath := helpers.GetString(req.Arguments, "output_path")
		if outputPath == "" {
			tmp, err := os.CreateTemp("", "screenshot-window-*.png")
			if err != nil {
				return helpers.ErrorResult("temp_file_error", err.Error()), nil
			}
			outputPath = tmp.Name()
			tmp.Close()
		}

		outputPath, _ = filepath.Abs(outputPath)

		if err := capture.CaptureWindow(title, outputPath); err != nil {
			return helpers.ErrorResult("capture_error", fmt.Sprintf("failed to capture window %q: %v", title, err)), nil
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read screenshot: %v", err)), nil
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		resp, err := helpers.JSONResult(map[string]any{
			"file_path":    outputPath,
			"window_title": title,
			"size_bytes":   len(data),
			"image_base64": encoded,
		})
		if err != nil {
			return helpers.ErrorResult("result_error", err.Error()), nil
		}
		return resp, nil
	}
}
