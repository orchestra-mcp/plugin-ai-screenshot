package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ListCapturesSchema returns the JSON Schema for the list_captures tool.
func ListCapturesSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to scan for captures (defaults to system temp directory)",
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of captures to return (default 20)",
			},
		},
	})
	return s
}

// ListCaptures returns a tool handler that lists recent screenshot captures.
func ListCaptures() func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		dir := helpers.GetString(req.Arguments, "directory")
		if dir == "" {
			dir = os.TempDir()
		}
		dir, _ = filepath.Abs(dir)

		limit := helpers.GetInt(req.Arguments, "limit")
		if limit <= 0 {
			limit = 20
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return helpers.ErrorResult("read_dir_error", fmt.Sprintf("failed to read directory %s: %v", dir, err)), nil
		}

		type captureInfo struct {
			Name     string `json:"name"`
			Path     string `json:"path"`
			Size     int64  `json:"size_bytes"`
			Modified string `json:"modified"`
		}

		var captures []captureInfo
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			lower := strings.ToLower(name)
			if !strings.HasSuffix(lower, ".png") && !strings.HasSuffix(lower, ".jpg") && !strings.HasSuffix(lower, ".jpeg") {
				continue
			}
			// Filter to files that look like screenshots
			if !strings.Contains(lower, "screenshot") && !strings.HasPrefix(lower, "screen") && !strings.HasPrefix(lower, "capture") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			captures = append(captures, captureInfo{
				Name:     name,
				Path:     filepath.Join(dir, name),
				Size:     info.Size(),
				Modified: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Sort by modification time (newest first)
		sort.Slice(captures, func(i, j int) bool {
			return captures[i].Modified > captures[j].Modified
		})

		if len(captures) > limit {
			captures = captures[:limit]
		}

		resp, err := helpers.JSONResult(map[string]any{
			"directory": dir,
			"count":     len(captures),
			"captures":  captures,
		})
		if err != nil {
			return helpers.ErrorResult("result_error", err.Error()), nil
		}
		return resp, nil
	}
}
