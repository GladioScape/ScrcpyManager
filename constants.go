package main

import "time"

// UI Constants
const (
	DefaultWindowWidth  = 900
	DefaultWindowHeight = 700
	FormColumns         = 2
	MaxAppsToDisplay    = 200
	AppGridDefaultCols  = 6
	IconItemWidth       = 140
	IconItemHeight      = 60
)

// Device Constants
const (
	DeviceRefreshInterval = 2 * time.Second
	DeviceScanTimeout     = 5 * time.Second
)

// App Detection Constants
const (
	AppCacheExpiry   = 5 * time.Minute
	MaxConcurrentOps = 6
	SizeBatchLimit   = 200
)

// Scrcpy Default Settings
const (
	DefaultBitrate       = 8000000 // 8 Mbps
	DefaultMaxFPS        = 60
	DefaultDisplayWidth  = 1080
	DefaultDisplayHeight = 2400
	MinBitrate           = 100000    // 100 Kbps
	MaxBitrate           = 100000000 // 100 Mbps
	MinFPS               = 1
	MaxFPS               = 240
)

// Process Management
const (
	ProcessStartTimeout = 10 * time.Second
	ProcessStopTimeout  = 5 * time.Second
	MaxRunningApps      = 20
)

// Bitrate Presets
var BitratePresets = map[string]int{
	"Performance (2 Mbps)": 2000000,
	"Balanced (8 Mbps)":    8000000,
	"Quality (20 Mbps)":    20000000,
	"Max (50 Mbps)":        50000000,
}

// Status Messages
const (
	StatusReady             = "Ready"
	StatusDeviceRefreshed   = "Devices refreshed"
	StatusNoDeviceSelected  = "No device selected"
	StatusLoadingApps       = "Loading apps..."
	StatusSettingsSaved     = "Settings saved successfully"
	StatusAppLaunched       = "App launched successfully"
	StatusAppAlreadyRunning = "App is already running"
	StatusAllAppsStopped    = "All apps stopped"
)
