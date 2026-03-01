package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// AnnotateScreenshotSchema returns the JSON Schema for the annotate_screenshot tool.
func AnnotateScreenshotSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"image_path": map[string]any{
				"type":        "string",
				"description": "Path to the screenshot image to annotate",
			},
			"annotations": map[string]any{
				"type":        "string",
				"description": "JSON string of annotations array, e.g. [{\"type\":\"rectangle\",\"x\":10,\"y\":10,\"width\":100,\"height\":50,\"color\":\"red\"},{\"type\":\"text\",\"x\":20,\"y\":80,\"text\":\"Bug here\"}]",
			},
			"output_path": map[string]any{
				"type":        "string",
				"description": "File path to save the annotated image (defaults to temp file)",
			},
		},
		"required": []any{"image_path", "annotations"},
	})
	return s
}

// AnnotateScreenshot returns a tool handler that annotates a screenshot.
// This is a placeholder implementation that returns the original image with annotation metadata.
func AnnotateScreenshot() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "image_path", "annotations"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		imagePath := helpers.GetString(req.Arguments, "image_path")
		annotationsJSON := helpers.GetString(req.Arguments, "annotations")

		// Validate the image exists
		imagePath, _ = filepath.Abs(imagePath)
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return helpers.ErrorResult("not_found", fmt.Sprintf("image not found: %s", imagePath)), nil
		}

		// Parse annotations to validate JSON
		var annotations []map[string]any
		if err := json.Unmarshal([]byte(annotationsJSON), &annotations); err != nil {
			return helpers.ErrorResult("validation_error", fmt.Sprintf("invalid annotations JSON: %v", err)), nil
		}

		// Determine output path
		outputPath := helpers.GetString(req.Arguments, "output_path")
		if outputPath == "" {
			tmp, err := os.CreateTemp("", "screenshot-annotated-*.png")
			if err != nil {
				return helpers.ErrorResult("temp_file_error", err.Error()), nil
			}
			outputPath = tmp.Name()
			tmp.Close()
		}
		outputPath, _ = filepath.Abs(outputPath)

		// Placeholder: copy original image to output path with annotation metadata
		// Actual image editing would require an image processing library.
		data, err := os.ReadFile(imagePath)
		if err != nil {
			return helpers.ErrorResult("read_error", fmt.Sprintf("failed to read image: %v", err)), nil
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return helpers.ErrorResult("write_error", fmt.Sprintf("failed to write annotated image: %v", err)), nil
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		resp, err := helpers.JSONResult(map[string]any{
			"file_path":        outputPath,
			"source_path":      imagePath,
			"size_bytes":       len(data),
			"annotation_count": len(annotations),
			"annotations":      annotations,
			"note":             "Placeholder: original image copied with annotation metadata. Image overlay rendering requires an image processing library.",
			"image_base64":     encoded,
		})
		if err != nil {
			return helpers.ErrorResult("result_error", err.Error()), nil
		}
		return resp, nil
	}
}
