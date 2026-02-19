package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Application holds the main application state
type Application struct {
	fyneApp fyne.App
	window  fyne.Window
	ui      *UIManager
	helpers *UIHelpers
}

// NewApplication creates and initializes a new application
func NewApplication() (*Application, error) {
	// Initialize logger
	if err := InitLogger(LogInfo, true); err != nil {
		fmt.Printf("Warning: Failed to initialize logger: %v\n", err)
	}

	if appLogger != nil {
		appLogger.Info("Starting Scrcpy Manager")
	}

	// Load configuration
	if err := LoadConfig(); err != nil {
		if appLogger != nil {
			appLogger.Warning("Failed to load config, using defaults", "error", err)
		}
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		fmt.Println("Using default configuration")
	}

	// Validate configuration
	if err := validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Fyne app
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(theme.DarkTheme())

	window := fyneApp.NewWindow("Scrcpy Link")
	window.Resize(fyne.NewSize(DefaultWindowWidth, DefaultWindowHeight))

	helpers := NewUIHelpers(window)

	appInstance := &Application{
		fyneApp: fyneApp,
		window:  window,
		helpers: helpers,
	}

	// Initialize UI manager
	appInstance.ui = NewUIManager(window, helpers)

	return appInstance, nil
}

// Run starts the application
func (a *Application) Run() {
	// Build UI
	content := a.ui.BuildUI()
	a.window.SetContent(content)

	// Start device monitoring
	deviceManager.StartMonitoring()
	if appLogger != nil {
		appLogger.Info("Device monitoring started")
	}

	// Add event listeners
	deviceManager.AddListener(a.ui.OnDevicesChanged)
	appLauncher.AddListener(a.ui.OnRunningAppsChanged)

	// Initial device scan
	a.ui.RefreshDevices()

	// Setup cleanup on close
	a.window.SetOnClosed(func() {
		a.Shutdown()
	})

	// Show window and run
	if appLogger != nil {
		appLogger.Info("Showing main window")
	}
	a.window.ShowAndRun()
}

// Shutdown performs cleanup before exit
func (a *Application) Shutdown() {
	if appLogger != nil {
		appLogger.Info("Shutting down application")
	}

	// Stop all running processes
	if err := processManager.StopAll(); err != nil {
		if appLogger != nil {
			appLogger.Error("Error stopping processes", "error", err)
		}
	}

	// Stop device monitoring
	deviceManager.StopMonitoring()
	if appLogger != nil {
		appLogger.Info("Device monitoring stopped")
	}

	// Stop app launcher monitoring
	appLauncher.StopAllApps()
	if appLogger != nil {
		appLogger.Info("All apps stopped")
	}

	// Save configuration
	if err := SaveConfig(); err != nil {
		if appLogger != nil {
			appLogger.Error("Failed to save config", "error", err)
		}
	}

	// Close logger
	CloseLogger()
}

// validateConfig validates the loaded configuration
func validateConfig() error {
	if err := ValidateBitrate(globalConfig.ScrcpyBitrate); err != nil {
		// Use default if invalid
		globalConfig.ScrcpyBitrate = DefaultBitrate
		if appLogger != nil {
			appLogger.Warning("Invalid bitrate in config, using default", "bitrate", DefaultBitrate)
		}
	}

	if err := ValidateFPS(globalConfig.ScrcpyMaxFPS); err != nil {
		globalConfig.ScrcpyMaxFPS = DefaultMaxFPS
		if appLogger != nil {
			appLogger.Warning("Invalid FPS in config, using default", "fps", DefaultMaxFPS)
		}
	}

	if err := ValidateResolution(globalConfig.DefaultDisplayWidth, globalConfig.DefaultDisplayHeight); err != nil {
		globalConfig.DefaultDisplayWidth = DefaultDisplayWidth
		globalConfig.DefaultDisplayHeight = DefaultDisplayHeight
		if appLogger != nil {
			appLogger.Warning("Invalid resolution in config, using default")
		}
	}

	// Validate paths exist
	if globalConfig.ADBPath == "" {
		globalConfig.ADBPath = "adb"
	}

	if globalConfig.ScrcpyPath == "" {
		globalConfig.ScrcpyPath = "scrcpy"
	}

	return nil
}

// UIManager manages all UI components
type UIManager struct {
	window  fyne.Window
	helpers *UIHelpers

	// UI Components
	deviceTab   *DeviceTabUI
	appsTab     *AppsTabUI
	runningTab  *RunningTabUI
	settingsTab *SettingsTabUI
	statusLabel *widget.Label

	// State
	selectedDevice *Device
}

// NewUIManager creates a new UI manager
func NewUIManager(window fyne.Window, helpers *UIHelpers) *UIManager {
	manager := &UIManager{
		window:  window,
		helpers: helpers,
	}

	// Initialize tab components
	manager.deviceTab = NewDeviceTabUI(manager)
	manager.appsTab = NewAppsTabUI(manager)
	manager.runningTab = NewRunningTabUI(manager)
	manager.settingsTab = NewSettingsTabUI(manager)

	return manager
}

// BuildUI constructs the main UI
func (m *UIManager) BuildUI() fyne.CanvasObject {
	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Devices", m.deviceTab.Build()),
		container.NewTabItem("Apps", m.appsTab.Build()),
		container.NewTabItem("Running", m.runningTab.Build()),
		container.NewTabItem("Settings", m.settingsTab.Build()),
	)

	// Status bar
	m.statusLabel = widget.NewLabel(StatusReady)
	m.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	statusBar := container.NewBorder(
		nil, nil, nil, nil,
		container.NewPadded(m.statusLabel),
	)

	// Main layout
	return container.NewBorder(
		nil,       // top
		statusBar, // bottom
		nil, nil,  // left, right
		tabs, // center
	)
}

// SetSelectedDevice sets the currently selected device
func (m *UIManager) SetSelectedDevice(device *Device) {
	m.selectedDevice = device
	if appLogger != nil {
		appLogger.Info("Device selected", "serial", device.Serial, "model", device.Model)
	}

	// Notify apps tab
	m.appsTab.OnDeviceSelected(device)
}

// GetSelectedDevice returns the currently selected device
func (m *UIManager) GetSelectedDevice() *Device {
	return m.selectedDevice
}

// UpdateStatus updates the status bar message
func (m *UIManager) UpdateStatus(message string) {
	if appLogger != nil {
		appLogger.Info("Status", "message", message)
	}
	if m.statusLabel != nil {
		m.statusLabel.SetText(message)
	}
}

// RefreshDevices triggers a device refresh
func (m *UIManager) RefreshDevices() {
	if appLogger != nil {
		appLogger.Debug("Refreshing devices")
	}
	deviceManager.scanDevices()
	devices := deviceManager.GetDevices()
	m.deviceTab.UpdateDevices(devices)
	m.UpdateStatus(StatusDeviceRefreshed)
}

// OnDevicesChanged is called when devices change
func (m *UIManager) OnDevicesChanged(devices []*Device) {
	if appLogger != nil {
		appLogger.Debug("Devices changed", "count", len(devices))
	}
	m.deviceTab.UpdateDevices(devices)
}

// OnRunningAppsChanged is called when running apps change
func (m *UIManager) OnRunningAppsChanged(apps []*RunningApp) {
	if appLogger != nil {
		appLogger.Debug("Running apps changed", "count", len(apps))
	}
	m.runningTab.UpdateRunningApps(apps)
}

// Global variable for backward compatibility
var (
	mainWindow      fyne.Window
	selectedDevice  *Device
	deviceList      *widget.List
	appGrid         *fyne.Container
	runningAppsList *widget.List
	statusLabel     *widget.Label

	appSearchEntry   *widget.Entry
	viewModeSelect   *widget.Select
	showPackageCheck *widget.Check
	showAppSizeCheck *widget.Check

	noSelectionLabel   *widget.Label
	appScrollContainer *container.Scroll

	appSortMode string

	devices     []*Device
	apps        []*AndroidApp
	runningApps []*RunningApp
)

func main() {
	// Create and run application
	app, err := NewApplication()
	if err != nil {
		Log(LogError, "Failed to load configuration: %v", err)
		os.Exit(1)
	}

	app.Run()
}

// updateStatus updates the global status label (for backward compatibility)
func updateStatus(message string) {
	if appLogger != nil {
		appLogger.Info("Status", "message", message)
	}
	if statusLabel != nil {
		fyne.Do(func() {
			statusLabel.SetText(message)
		})
	}
}
