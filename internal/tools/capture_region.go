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

// CaptureRegionSchema returns the JSON Schema for the capture_region tool.
func CaptureRegionSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"x": map[string]any{
				"type":        "number",
				"description": "X coordinate of the top-left corner",
			},
			"y": map[string]any{
				"type":        "number",
				"description": "Y coordinate of the top-left corner",
			},
			"width": map[string]any{
				"type":        "number",
				"description": "Width of the capture region in pixels",
			},
			"height": map[string]any{
				"type":        "number",
				"description": "Height of the capture region in pixels",
			},
			"output_path": map[string]any{
				"type":        "string",
				"description": "File path to save the screenshot (defaults to temp file)",
			},
		},
		"required": []any{"x", "y", "width", "height"},
	})
	return s
}

// CaptureRegion returns a tool handler that captures a specific screen region.
func CaptureRegion() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		x := helpers.GetInt(req.Arguments, "x")
		y := helpers.GetInt(req.Arguments, "y")
		w := helpers.GetInt(req.Arguments, "width")
		h := helpers.GetInt(req.Arguments, "height")

		if w <= 0 || h <= 0 {
			return helpers.ErrorResult("validation_error", "width and height must be positive"), nil
		}

		outputPath := helpers.GetString(req.Arguments, "output_path")
		if outputPath == "" {
			tmp, err := os.CreateTemp("", "screenshot-region-*.png")
			if err != nil {
				return helpers.ErrorResult("temp_file_error", err.Error()), nil
			}
			outputPath = tmp.Name()
			tmp.Close()
		}

		outputPath, _ = filepath.Abs(outputPath)

		if err := capture.CaptureRegion(x, y, w, h, outputPath); err != nil {
			return helpers.ErrorResult("capture_error", fmt.Sprintf("failed to capture region: %v", err)), nil
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read screenshot: %v", err)), nil
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		resp, err := helpers.JSONResult(map[string]any{
			"file_path":    outputPath,
			"size_bytes":   len(data),
			"region":       map[string]any{"x": x, "y": y, "width": w, "height": h},
			"image_base64": encoded,
		})
		if err != nil {
			return helpers.ErrorResult("result_error", err.Error()), nil
		}
		return resp, nil
	}
}
