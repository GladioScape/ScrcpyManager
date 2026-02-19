package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SettingsTabUI manages the settings tab interface
type SettingsTabUI struct {
	manager *UIManager
}

// NewSettingsTabUI creates a new settings tab UI
func NewSettingsTabUI(manager *UIManager) *SettingsTabUI {
	return &SettingsTabUI{
		manager: manager,
	}
}

// Build constructs the settings tab UI
func (s *SettingsTabUI) Build() fyne.CanvasObject {
	// Display selection
	displaySelect := s.createDisplaySelect()
	refreshDisplaysBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		s.refreshDisplays(displaySelect)
	})

	// Buffer/Quality selection
	bufferSelect := s.createBufferSelect()

	// FPS entry
	fpsEntry := widget.NewEntry()
	fpsEntry.SetText(fmt.Sprintf("%d", globalConfig.ScrcpyMaxFPS))

	// Option checkboxes
	stayAwakeCheck := widget.NewCheck("Keep device awake", func(checked bool) {
		globalConfig.StayAwake = checked
	})
	stayAwakeCheck.Checked = globalConfig.StayAwake

	turnScreenOffCheck := widget.NewCheck("Turn screen off", func(checked bool) {
		globalConfig.TurnScreenOff = checked
	})
	turnScreenOffCheck.Checked = globalConfig.TurnScreenOff

	disableAudioCheck := widget.NewCheck("Disable audio", func(checked bool) {
		globalConfig.DisableAudio = checked
	})
	disableAudioCheck.Checked = globalConfig.DisableAudio

	alwaysOnTopCheck := widget.NewCheck("Always on top", func(checked bool) {
		globalConfig.AlwaysOnTop = checked
	})
	alwaysOnTopCheck.Checked = globalConfig.AlwaysOnTop

	borderlessCheck := widget.NewCheck("Borderless windows", func(checked bool) {
		globalConfig.Borderless = checked
	})
	borderlessCheck.Checked = globalConfig.Borderless

	extraArgsEntry := widget.NewEntry()
	extraArgsEntry.SetPlaceHolder("Custom scrcpy arguments")
	extraArgsEntry.SetText(globalConfig.ExtraScrcpyArgs)

	// Save button
	saveBtn := widget.NewButton("Save Settings", func() {
		s.saveSettings(fpsEntry.Text, extraArgsEntry.Text)
	})
	saveBtn.Importance = widget.HighImportance

	// Build form
	form := widget.NewForm(
		widget.NewFormItem("Display", container.NewBorder(nil, nil, nil, refreshDisplaysBtn, displaySelect)),
		widget.NewFormItem("Video Quality", bufferSelect),
		widget.NewFormItem("Max FPS", fpsEntry),
		widget.NewFormItem("", stayAwakeCheck),
		widget.NewFormItem("", turnScreenOffCheck),
		widget.NewFormItem("", disableAudioCheck),
		widget.NewFormItem("", alwaysOnTopCheck),
		widget.NewFormItem("", borderlessCheck),
		widget.NewFormItem("Extra Arguments", extraArgsEntry),
	)

	header := widget.NewLabelWithStyle("Application Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	return container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		container.NewPadded(saveBtn),
		nil, nil,
		container.NewVScroll(form),
	)
}

// createDisplaySelect creates the display selection dropdown for PC monitors
func (s *SettingsTabUI) createDisplaySelect() *widget.Select {
	// Get PC monitors from window manager
	monitors := windowManager.GetMonitors()

	// Build options list
	opts := make([]string, 0, len(monitors))
	monitorIDMap := make(map[string]int)

	for _, m := range monitors {
		opts = append(opts, m.Label)
		monitorIDMap[m.Label] = m.ID
	}

	// Fallback if no monitors detected
	if len(opts) == 0 {
		defaultLabel := fmt.Sprintf("Default (%dx%d)", globalConfig.DefaultDisplayWidth, globalConfig.DefaultDisplayHeight)
		opts = []string{defaultLabel}
		monitorIDMap[defaultLabel] = 0
	}

	displaySelect := widget.NewSelect(opts, nil)

	// Select current monitor or first one
	selectedLabel := ""
	for _, m := range monitors {
		if m.ID == globalConfig.DefaultMonitorID {
			selectedLabel = m.Label
			break
		}
	}
	if selectedLabel == "" && len(opts) > 0 {
		selectedLabel = opts[0]
	}
	displaySelect.SetSelected(selectedLabel)

	displaySelect.OnChanged = func(selected string) {
		if id, ok := monitorIDMap[selected]; ok {
			globalConfig.DefaultMonitorID = id
			// Also update the target display ID for scrcpy window positioning
			globalConfig.TargetDisplayID = id

			// Find the monitor and update default display dimensions
			for _, m := range monitors {
				if m.ID == id {
					globalConfig.DefaultDisplayWidth = m.Width
					globalConfig.DefaultDisplayHeight = m.Height
					if appLogger != nil {
						appLogger.Info("Selected PC monitor", "id", id, "resolution", fmt.Sprintf("%dx%d", m.Width, m.Height))
					}
					break
				}
			}
		}
	}

	return displaySelect
}

// createBufferSelect creates the buffer/quality selection dropdown
func (s *SettingsTabUI) createBufferSelect() *widget.Select {
	currentBuffer := "Balanced (8 Mbps)"
	for label, val := range BitratePresets {
		if val == globalConfig.ScrcpyBitrate {
			currentBuffer = label
			break
		}
	}

	bufferSelect := widget.NewSelect([]string{
		"Performance (2 Mbps)",
		"Balanced (8 Mbps)",
		"Quality (20 Mbps)",
		"Max (50 Mbps)",
	}, func(selected string) {
		if val, ok := BitratePresets[selected]; ok {
			if err := ValidateBitrate(val); err == nil {
				globalConfig.ScrcpyBitrate = val
			} else if appLogger != nil {
				appLogger.Warning("Invalid bitrate", "error", err)
			}
		}
	})
	bufferSelect.SetSelected(currentBuffer)

	return bufferSelect
}

// refreshDisplays refreshes the available PC monitors
func (s *SettingsTabUI) refreshDisplays(displaySelect *widget.Select) {
	s.manager.helpers.RunAsync("RefreshDisplays", func() error {
		// Refresh PC monitors
		windowManager.RefreshMonitors()
		monitors := windowManager.GetMonitors()

		opts := make([]string, 0, len(monitors))
		monitorIDMap := make(map[string]int)

		for _, m := range monitors {
			opts = append(opts, m.Label)
			monitorIDMap[m.Label] = m.ID
		}

		if len(opts) == 0 {
			def := fmt.Sprintf("Default (%dx%d)", globalConfig.DefaultDisplayWidth, globalConfig.DefaultDisplayHeight)
			opts = []string{def}
			monitorIDMap[def] = 0
		}

		// Update the select widget on the main thread
		fyne.Do(func() {
			displaySelect.Options = opts
			displaySelect.Refresh()

			// Select the current monitor or first one
			selectedLabel := ""
			for _, m := range monitors {
				if m.ID == globalConfig.DefaultMonitorID {
					selectedLabel = m.Label
					break
				}
			}
			if selectedLabel == "" && len(opts) > 0 {
				selectedLabel = opts[0]
			}
			displaySelect.SetSelected(selectedLabel)
		})

		s.manager.UpdateStatus(fmt.Sprintf("Found %d PC monitor(s)", len(opts)))

		return nil
	})
}

// saveSettings saves the current settings
func (s *SettingsTabUI) saveSettings(fpsText, extraArgs string) {
	// Validate and parse FPS
	var fps int
	if _, err := fmt.Sscanf(fpsText, "%d", &fps); err != nil {
		s.manager.helpers.ShowError(NewAppError("SaveSettings", err, "Invalid FPS value"))
		return
	}

	if err := ValidateFPS(fps); err != nil {
		s.manager.helpers.ShowError(err)
		return
	}

	globalConfig.ScrcpyMaxFPS = fps
	globalConfig.ExtraScrcpyArgs = extraArgs

	// Save configuration
	if err := SaveConfig(); err != nil {
		s.manager.helpers.ShowError(NewAppError("SaveSettings", err, "Failed to save settings"))
		return
	}

	s.manager.UpdateStatus(StatusSettingsSaved)
	if appLogger != nil {
		appLogger.Info("Settings saved", "fps", fps)
	}
}

// Note: PC monitor detection is now handled by windowManager in window_manager.go
// The getDeviceDisplaysForSettings function has been removed as we now use PC monitors
// instead of Android device displays for the display selection setting.
