package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	// Paths
	ScrcpyPath string `json:"scrcpy_path"`
	ADBPath    string `json:"adb_path"`

	// Virtual Display Settings
	DefaultDisplayWidth  int `json:"default_display_width"`
	DefaultDisplayHeight int `json:"default_display_height"`

	// Window Management
	EnableAutoTiling bool   `json:"enable_auto_tiling"`
	DefaultMonitorID int    `json:"default_monitor_id"`
	TilingMode       string `json:"tiling_mode"` // "grid", "side-by-side", "manual"

	// App Launch Settings
	MaxConcurrentApps int    `json:"max_concurrent_apps"`
	AppViewMode       string `json:"app_view_mode"`
	ShowPackageName   bool   `json:"show_package_name"`
	ShowAppSize       bool   `json:"show_app_size"`
	AppGridColumns    int    `json:"app_grid_columns"`

	// Scrcpy Options
	ScrcpyBitrate   int    `json:"scrcpy_bitrate"`
	ScrcpyMaxFPS    int    `json:"scrcpy_max_fps"`
	ScrcpyCodec     string `json:"scrcpy_codec"`
	StayAwake       bool   `json:"stay_awake"`
	TurnScreenOff   bool   `json:"turn_screen_off"`
	DisableAudio    bool   `json:"disable_audio"`
	AlwaysOnTop     bool   `json:"always_on_top"`
	Borderless      bool   `json:"borderless"`
	ExtraScrcpyArgs string `json:"extra_scrcpy_args"`
	TargetDisplayID int    `json:"target_display_id"`
}

var globalConfig Config

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		ScrcpyPath:           "scrcpy", // Assume in PATH
		ADBPath:              "adb",    // Assume in PATH
		DefaultDisplayWidth:  1080,
		DefaultDisplayHeight: 1920,
		EnableAutoTiling:     true,
		DefaultMonitorID:     0,
		TilingMode:           "grid",
		MaxConcurrentApps: 6,
		AppViewMode:       "icons",
		ShowPackageName:   false,
		ShowAppSize:       false,
		AppGridColumns:    6,
		ScrcpyBitrate:   8000000,
		ScrcpyMaxFPS:    60,
		ScrcpyCodec:     "h264",
		StayAwake:       true,
		TurnScreenOff:   false,
		DisableAudio:    false,
		AlwaysOnTop:     false,
		Borderless:      false,
		ExtraScrcpyArgs: "",
		TargetDisplayID: 0,
	}
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() error {
	configPath := getConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		globalConfig = DefaultConfig()
		return SaveConfig()
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &globalConfig)
	if err != nil {
		// If config is corrupt, use defaults
		globalConfig = DefaultConfig()
		return SaveConfig()
	}

	return nil
}

// SaveConfig saves current configuration to file
func SaveConfig() error {
	configPath := getConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	// For now, store in same directory as executable
	// Could be moved to user config directory later
	return "app_config.json"
}
