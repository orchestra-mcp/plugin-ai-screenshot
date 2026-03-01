package capture

import (
	"fmt"
	"os/exec"
	"runtime"
)

// CaptureScreen captures the entire screen to the given output file.
func CaptureScreen(output string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("screencapture", "-x", output).Run()
	case "linux":
		return exec.Command("gnome-screenshot", "-f", output).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CaptureRegion captures a specific region of the screen to the given output file.
func CaptureRegion(x, y, w, h int, output string) error {
	switch runtime.GOOS {
	case "darwin":
		rect := fmt.Sprintf("-R%d,%d,%d,%d", x, y, w, h)
		return exec.Command("screencapture", "-x", rect, output).Run()
	case "linux":
		return exec.Command("gnome-screenshot", "-a",
			fmt.Sprintf("--delay=0"),
			"-f", output).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CaptureWindow captures a window by title to the given output file.
func CaptureWindow(title string, output string) error {
	switch runtime.GOOS {
	case "darwin":
		// Use osascript to find the window by title and then screencapture by window ID
		script := fmt.Sprintf(
			`tell application "System Events" to set wid to id of first window of (first process whose name contains "%s")`,
			title,
		)
		idOut, err := exec.Command("osascript", "-e", script).Output()
		if err != nil {
			// Fallback: capture the frontmost window
			return exec.Command("screencapture", "-x", "-l0", output).Run()
		}
		windowID := string(idOut)
		if len(windowID) > 0 && windowID[len(windowID)-1] == '\n' {
			windowID = windowID[:len(windowID)-1]
		}
		return exec.Command("screencapture", "-x", "-l"+windowID, output).Run()
	case "linux":
		return exec.Command("gnome-screenshot", "-w", "-f", output).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CaptureInteractive launches an interactive capture (user selects region).
func CaptureInteractive(output string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("screencapture", "-i", output).Run()
	case "linux":
		return exec.Command("gnome-screenshot", "-a", "-f", output).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
