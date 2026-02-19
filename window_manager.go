package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// MonitorInfo holds information about a PC monitor
type MonitorInfo struct {
	ID      int
	Name    string
	Width   int
	Height  int
	X       int // Left position
	Y       int // Top position
	Primary bool
	Label   string
}

// WindowManager handles window positioning and tiling
type WindowManager struct {
	monitors     []MonitorInfo
	monitorsLock sync.RWMutex
}

var windowManager *WindowManager

func init() {
	windowManager = NewWindowManager()
}

// NewWindowManager creates a new window manager
func NewWindowManager() *WindowManager {
	wm := &WindowManager{
		monitors: make([]MonitorInfo, 0),
	}
	// Initial monitor detection
	wm.RefreshMonitors()
	return wm
}

// TileWindows arranges windows according to the configured tiling mode
// TODO: Implement advanced tiling in Phase 2
func (wm *WindowManager) TileWindows() {
	// Placeholder for future implementation
	// Will use Windows API to position scrcpy windows intelligently
}

// GetMonitors returns the list of detected monitors
func (wm *WindowManager) GetMonitors() []MonitorInfo {
	wm.monitorsLock.RLock()
	defer wm.monitorsLock.RUnlock()

	result := make([]MonitorInfo, len(wm.monitors))
	copy(result, wm.monitors)
	return result
}

// RefreshMonitors detects available PC monitors using PowerShell
func (wm *WindowManager) RefreshMonitors() {
	wm.monitorsLock.Lock()
	defer wm.monitorsLock.Unlock()

	monitors := make([]MonitorInfo, 0)

	// Use PowerShell to get monitor information via WMI
	// This gets actual screen dimensions and positions
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
$screens = [System.Windows.Forms.Screen]::AllScreens
$i = 0
foreach ($screen in $screens) {
    $bounds = $screen.Bounds
    $primary = $screen.Primary
    Write-Output "MONITOR:$i|$($screen.DeviceName)|$($bounds.Width)|$($bounds.Height)|$($bounds.X)|$($bounds.Y)|$primary"
    $i++
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.Output()

	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "MONITOR:") {
				parts := strings.Split(strings.TrimPrefix(line, "MONITOR:"), "|")
				if len(parts) >= 7 {
					id, _ := strconv.Atoi(parts[0])
					name := parts[1]
					width, _ := strconv.Atoi(parts[2])
					height, _ := strconv.Atoi(parts[3])
					x, _ := strconv.Atoi(parts[4])
					y, _ := strconv.Atoi(parts[5])
					primary := strings.ToLower(strings.TrimSpace(parts[6])) == "true"

					if width > 0 && height > 0 {
						label := fmt.Sprintf("Monitor %d: %dx%d", id+1, width, height)
						if primary {
							label += " (Primary)"
						}
						// Clean up device name
						cleanName := name
						if idx := strings.LastIndex(name, "\\"); idx >= 0 {
							cleanName = name[idx+1:]
						}
						if cleanName != "" {
							label = fmt.Sprintf("Monitor %d (%s): %dx%d", id+1, cleanName, width, height)
							if primary {
								label += " (Primary)"
							}
						}

						monitors = append(monitors, MonitorInfo{
							ID:      id,
							Name:    name,
							Width:   width,
							Height:  height,
							X:       x,
							Y:       y,
							Primary: primary,
							Label:   label,
						})
					}
				}
			}
		}
	}

	// Fallback: Try using wmic if PowerShell failed
	if len(monitors) == 0 {
		monitors = wm.detectMonitorsWMIC()
	}

	// Final fallback: Add a default monitor
	if len(monitors) == 0 {
		monitors = append(monitors, MonitorInfo{
			ID:      0,
			Name:    "Default",
			Width:   1920,
			Height:  1080,
			X:       0,
			Y:       0,
			Primary: true,
			Label:   "Monitor 1: 1920x1080 (Primary)",
		})
	}

	// Sort by ID
	sort.Slice(monitors, func(i, j int) bool {
		return monitors[i].ID < monitors[j].ID
	})

	wm.monitors = monitors

	if appLogger != nil {
		appLogger.Info("Detected PC monitors", "count", len(monitors))
		for _, m := range monitors {
			appLogger.Debug("Monitor info", "id", m.ID, "resolution", fmt.Sprintf("%dx%d", m.Width, m.Height), "position", fmt.Sprintf("%d,%d", m.X, m.Y))
		}
	}
}

// detectMonitorsWMIC uses wmic as fallback for monitor detection
func (wm *WindowManager) detectMonitorsWMIC() []MonitorInfo {
	monitors := make([]MonitorInfo, 0)

	// Get video controller info for resolution
	cmd := exec.Command("wmic", "path", "Win32_VideoController", "get", "CurrentHorizontalResolution,CurrentVerticalResolution,Name", "/format:csv")
	output, err := cmd.Output()

	if err == nil {
		lines := strings.Split(string(output), "\n")
		id := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "Node") {
				continue
			}
			// CSV format: Node,CurrentHorizontalResolution,CurrentVerticalResolution,Name
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				width, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				height, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
				name := strings.TrimSpace(parts[3])

				if width > 0 && height > 0 {
					monitors = append(monitors, MonitorInfo{
						ID:      id,
						Name:    name,
						Width:   width,
						Height:  height,
						X:       0,
						Y:       0,
						Primary: id == 0,
						Label:   fmt.Sprintf("Monitor %d: %dx%d", id+1, width, height),
					})
					id++
				}
			}
		}
	}

	// Alternative: Parse systeminfo for basic display info
	if len(monitors) == 0 {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			"Get-CimInstance -ClassName Win32_DesktopMonitor | Select-Object ScreenWidth, ScreenHeight | ConvertTo-Json")
		output, err := cmd.Output()
		if err == nil {
			// Parse JSON output for monitor dimensions
			re := regexp.MustCompile(`"ScreenWidth"\s*:\s*(\d+).*?"ScreenHeight"\s*:\s*(\d+)`)
			matches := re.FindAllStringSubmatch(string(output), -1)
			for i, match := range matches {
				if len(match) >= 3 {
					width, _ := strconv.Atoi(match[1])
					height, _ := strconv.Atoi(match[2])
					if width > 0 && height > 0 {
						monitors = append(monitors, MonitorInfo{
							ID:      i,
							Name:    fmt.Sprintf("Display %d", i+1),
							Width:   width,
							Height:  height,
							X:       0,
							Y:       0,
							Primary: i == 0,
							Label:   fmt.Sprintf("Monitor %d: %dx%d", i+1, width, height),
						})
					}
				}
			}
		}
	}

	return monitors
}

// GetMonitorByID returns monitor info by ID, or nil if not found
func (wm *WindowManager) GetMonitorByID(id int) *MonitorInfo {
	wm.monitorsLock.RLock()
	defer wm.monitorsLock.RUnlock()

	for i := range wm.monitors {
		if wm.monitors[i].ID == id {
			return &wm.monitors[i]
		}
	}
	return nil
}

// GetPrimaryMonitor returns the primary monitor info
func (wm *WindowManager) GetPrimaryMonitor() *MonitorInfo {
	wm.monitorsLock.RLock()
	defer wm.monitorsLock.RUnlock()

	for i := range wm.monitors {
		if wm.monitors[i].Primary {
			return &wm.monitors[i]
		}
	}
	if len(wm.monitors) > 0 {
		return &wm.monitors[0]
	}
	return nil
}
