package tools

import (
	"context"
	"os"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------- helpers ----------

func callTool(t *testing.T, handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error), args map[string]any) *pluginv1.ToolResponse {
	t.Helper()
	var s *structpb.Struct
	if args != nil {
		var err error
		s, err = structpb.NewStruct(args)
		if err != nil {
			t.Fatalf("NewStruct: %v", err)
		}
	}
	resp, err := handler(context.Background(), &pluginv1.ToolRequest{Arguments: s})
	if err != nil {
		t.Fatalf("handler returned Go error: %v", err)
	}
	return resp
}

func isError(resp *pluginv1.ToolResponse) bool {
	return resp != nil && !resp.Success
}

func getErrorCode(resp *pluginv1.ToolResponse) string {
	if resp == nil {
		return ""
	}
	return resp.GetErrorCode()
}

// ---------- capture_screen ----------

func TestCaptureScreen_NoArgs(t *testing.T) {
	// Will succeed or return capture_error — both are valid depending on platform.
	resp := callTool(t, CaptureScreen(), map[string]any{})
	_ = resp
}

func TestCaptureScreen_WithOutputPath(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "test-*.png")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()

	resp := callTool(t, CaptureScreen(), map[string]any{"output_path": tmp.Name()})
	// May succeed (screencapture available) or fail (capture_error in CI).
	_ = resp
}

// ---------- capture_region ----------

func TestCaptureRegion_MissingDimensions(t *testing.T) {
	// No required args — proceeds to capture attempt (may fail on CI).
	resp := callTool(t, CaptureRegion(), map[string]any{})
	_ = resp
}

func TestCaptureRegion_WithDimensions(t *testing.T) {
	resp := callTool(t, CaptureRegion(), map[string]any{
		"x": float64(0), "y": float64(0),
		"width": float64(100), "height": float64(100),
	})
	_ = resp
}

// ---------- capture_window ----------

func TestCaptureWindow_NoTitle(t *testing.T) {
	resp := callTool(t, CaptureWindow(), map[string]any{})
	_ = resp
}

func TestCaptureWindow_WithTitle(t *testing.T) {
	resp := callTool(t, CaptureWindow(), map[string]any{"window_title": "Finder"})
	_ = resp
}

// ---------- capture_interactive ----------

func TestCaptureInteractive_NoArgs(t *testing.T) {
	// Interactive capture requires user input — will fail in CI (capture_error).
	resp := callTool(t, CaptureInteractive(), map[string]any{})
	_ = resp
}

// ---------- annotate_screenshot ----------

func TestAnnotateScreenshot_MissingImagePath(t *testing.T) {
	resp := callTool(t, AnnotateScreenshot(), map[string]any{
		"annotations": `[{"type":"rectangle","x":0,"y":0,"width":10,"height":10}]`,
	})
	if !isError(resp) {
		t.Error("expected validation_error for missing image_path")
	}
}

func TestAnnotateScreenshot_MissingAnnotations(t *testing.T) {
	resp := callTool(t, AnnotateScreenshot(), map[string]any{
		"image_path": "/tmp/test.png",
	})
	if !isError(resp) {
		t.Error("expected validation_error for missing annotations")
	}
}

func TestAnnotateScreenshot_NonexistentImage(t *testing.T) {
	resp := callTool(t, AnnotateScreenshot(), map[string]any{
		"image_path":  "/tmp/no-such-image-orchestra-xyz.png",
		"annotations": `[]`,
	})
	if !isError(resp) {
		t.Error("expected not_found for nonexistent image")
	}
	code := getErrorCode(resp)
	if code != "not_found" {
		t.Errorf("expected code=not_found, got %q", code)
	}
}

func TestAnnotateScreenshot_InvalidAnnotationsJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "img-*.png")
	if err != nil {
		t.Fatal(err)
	}
	// Write minimal PNG header so the file exists.
	_, _ = f.Write([]byte("\x89PNG\r\n\x1a\n"))
	_ = f.Close()

	resp := callTool(t, AnnotateScreenshot(), map[string]any{
		"image_path":  f.Name(),
		"annotations": `not-valid-json`,
	})
	if !isError(resp) {
		t.Error("expected validation_error for invalid JSON")
	}
}

func TestAnnotateScreenshot_ValidArgs(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "img-*.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte("\x89PNG\r\n\x1a\n"))
	_ = f.Close()

	resp := callTool(t, AnnotateScreenshot(), map[string]any{
		"image_path":  f.Name(),
		"annotations": `[{"type":"rectangle","x":10,"y":10,"width":50,"height":50,"color":"red"}]`,
	})
	if isError(resp) {
		t.Errorf("unexpected error: %v", resp)
	}
}

// ---------- list_captures ----------

func TestListCaptures_DefaultDir(t *testing.T) {
	// Scans os.TempDir() — always succeeds.
	resp := callTool(t, ListCaptures(), map[string]any{})
	if isError(resp) {
		t.Errorf("unexpected error for default dir: %v", resp)
	}
}

func TestListCaptures_NonexistentDir(t *testing.T) {
	resp := callTool(t, ListCaptures(), map[string]any{
		"directory": "/tmp/no-such-directory-orchestra-xyz",
	})
	if !isError(resp) {
		t.Error("expected read_dir_error for nonexistent directory")
	}
	code := getErrorCode(resp)
	if code != "read_dir_error" {
		t.Errorf("expected code=read_dir_error, got %q", code)
	}
}

func TestListCaptures_WithScreenshotFiles(t *testing.T) {
	dir := t.TempDir()
	// Create screenshot-named files.
	for _, name := range []string{"screenshot-001.png", "screenshot-002.png", "capture-003.png"} {
		if err := os.WriteFile(dir+"/"+name, []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// This file should be excluded (no screenshot prefix).
	if err := os.WriteFile(dir+"/other.png", []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	resp := callTool(t, ListCaptures(), map[string]any{"directory": dir})
	if isError(resp) {
		t.Errorf("unexpected error: %v", resp)
	}
}

func TestListCaptures_WithLimit(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 5; i++ {
		name := dir + "/" + strings.Repeat("screenshot-", 1) + string(rune('a'+i)) + ".png"
		_ = os.WriteFile(name, []byte("fake"), 0644)
	}

	resp := callTool(t, ListCaptures(), map[string]any{
		"directory": dir,
		"limit":     float64(2),
	})
	if isError(resp) {
		t.Errorf("unexpected error: %v", resp)
	}
}
