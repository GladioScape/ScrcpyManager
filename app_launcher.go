package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// RunningApp represents a scrcpy instance running an app
type RunningApp struct {
	PackageName string
	AppName     string
	Serial      string
	Process     *exec.Cmd
	WindowTitle string
	StartTime   time.Time
}

// AppLauncher manages launching and stopping Android apps
type AppLauncher struct {
	runningApps map[string]*RunningApp // key: serial_packagename
	mutex       sync.RWMutex
	listeners   []func([]*RunningApp)
}

var appLauncher *AppLauncher

func init() {
	appLauncher = NewAppLauncher()
}

// NewAppLauncher creates a new app launcher
func NewAppLauncher() *AppLauncher {
	return &AppLauncher{
		runningApps: make(map[string]*RunningApp),
		listeners:   make([]func([]*RunningApp), 0),
	}
}

// LaunchApp launches an Android app in a separate scrcpy window
func (al *AppLauncher) LaunchApp(serial, packageName, appName string) error {
	key := fmt.Sprintf("%s_%s", serial, packageName)

	// Check if already running
	al.mutex.RLock()
	if _, exists := al.runningApps[key]; exists {
		al.mutex.RUnlock()
		return fmt.Errorf("app already running")
	}
	al.mutex.RUnlock()

	// Build scrcpy command
	windowTitle := fmt.Sprintf("%s - %s", appName, serial)
	displaySize := fmt.Sprintf("%dx%d", globalConfig.DefaultDisplayWidth, globalConfig.DefaultDisplayHeight)

	args := []string{
		"-s", serial,
		"--new-display=" + displaySize,
		"--start-app=" + packageName,
		"--window-title=" + windowTitle,
	}

	// Add optional flags
	if globalConfig.StayAwake {
		args = append(args, "--stay-awake")
	}
	if globalConfig.TurnScreenOff {
		args = append(args, "--turn-screen-off")
	}
	if globalConfig.DisableAudio {
		args = append(args, "--no-audio")
	}
	if globalConfig.AlwaysOnTop {
		args = append(args, "--always-on-top")
	}
	if globalConfig.Borderless {
		args = append(args, "--window-borderless")
	}
	if globalConfig.ScrcpyBitrate > 0 {
		args = append(args, "--video-bit-rate", fmt.Sprintf("%d", globalConfig.ScrcpyBitrate))
	}
	if globalConfig.ScrcpyMaxFPS > 0 {
		args = append(args, "--max-fps", fmt.Sprintf("%d", globalConfig.ScrcpyMaxFPS))
	}
	if globalConfig.ScrcpyCodec != "" {
		args = append(args, "--video-codec", globalConfig.ScrcpyCodec)
	}
	if globalConfig.TargetDisplayID > 0 {
		args = append(args, "--display", fmt.Sprintf("%d", globalConfig.TargetDisplayID))
	}
	if strings.TrimSpace(globalConfig.ExtraScrcpyArgs) != "" {
		extra := strings.Fields(globalConfig.ExtraScrcpyArgs)
		args = append(args, extra...)
	}

	fmt.Printf("Launching app: %s\nCommand: %s %s\n", appName, globalConfig.ScrcpyPath, strings.Join(args, " "))

	// Create command
	cmd := exec.Command(globalConfig.ScrcpyPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start scrcpy: %w", err)
	}

	// Track the running app
	runningApp := &RunningApp{
		PackageName: packageName,
		AppName:     appName,
		Serial:      serial,
		Process:     cmd,
		WindowTitle: windowTitle,
		StartTime:   time.Now(),
	}

	al.mutex.Lock()
	al.runningApps[key] = runningApp
	al.mutex.Unlock()

	// Monitor the process in a goroutine
	go al.monitorProcess(key, cmd)

	// Notify listeners
	al.notifyListeners()

	fmt.Printf("App launched successfully: %s\n", appName)
	return nil
}

// LaunchScreen starts a standard scrcpy device screen mirror (no --start-app)
func (al *AppLauncher) LaunchScreen(serial string) error {
	// Build args for full screen mirror
	args := []string{"-s", serial}
	if globalConfig.ScrcpyBitrate > 0 {
		args = append(args, "--video-bit-rate", fmt.Sprintf("%d", globalConfig.ScrcpyBitrate))
	}
	if globalConfig.ScrcpyMaxFPS > 0 {
		args = append(args, "--max-fps", fmt.Sprintf("%d", globalConfig.ScrcpyMaxFPS))
	}
	if globalConfig.StayAwake {
		args = append(args, "--stay-awake")
	}
	if globalConfig.TurnScreenOff {
		args = append(args, "--turn-screen-off")
	}
	if globalConfig.DisableAudio {
		args = append(args, "--no-audio")
	}
	if globalConfig.TargetDisplayID > 0 {
		args = append(args, "--display", fmt.Sprintf("%d", globalConfig.TargetDisplayID))
	}
	if strings.TrimSpace(globalConfig.ExtraScrcpyArgs) != "" {
		extra := strings.Fields(globalConfig.ExtraScrcpyArgs)
		args = append(args, extra...)
	}

	fmt.Printf("Launching phone screen for %s\nCommand: %s %s\n", serial, globalConfig.ScrcpyPath, strings.Join(args, " "))
	cmd := exec.Command(globalConfig.ScrcpyPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start scrcpy screen: %w", err)
	}

	key := fmt.Sprintf("%s_%s", serial, "__screen__")
	runningApp := &RunningApp{
		PackageName: "__screen__",
		AppName:     "Phone Screen",
		Serial:      serial,
		Process:     cmd,
		WindowTitle: fmt.Sprintf("scrcpy-screen-%s", serial),
		StartTime:   time.Now(),
	}
	al.mutex.Lock()
	al.runningApps[key] = runningApp
	al.mutex.Unlock()

	go al.monitorProcess(key, cmd)
	al.notifyListeners()
	return nil
}

// StopApp stops a running app
func (al *AppLauncher) StopApp(serial, packageName string) error {
	key := fmt.Sprintf("%s_%s", serial, packageName)

	al.mutex.Lock()
	runningApp, exists := al.runningApps[key]
	if !exists {
		al.mutex.Unlock()
		return fmt.Errorf("app not running")
	}

	// Kill the process
	if runningApp.Process != nil && runningApp.Process.Process != nil {
		err := runningApp.Process.Process.Kill()
		if err != nil {
			al.mutex.Unlock()
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	delete(al.runningApps, key)
	al.mutex.Unlock()

	// Notify listeners
	al.notifyListeners()

	fmt.Printf("App stopped: %s\n", runningApp.AppName)
	return nil
}

// monitorProcess monitors a scrcpy process and cleans up when it exits
func (al *AppLauncher) monitorProcess(key string, cmd *exec.Cmd) {
	// Wait for process to finish
	cmd.Wait()

	// Remove from running apps
	al.mutex.Lock()
	delete(al.runningApps, key)
	al.mutex.Unlock()

	// Notify listeners
	al.notifyListeners()

	fmt.Printf("Process exited: %s\n", key)
}

// GetRunningApps returns a list of all running apps
func (al *AppLauncher) GetRunningApps() []*RunningApp {
	al.mutex.RLock()
	defer al.mutex.RUnlock()

	apps := make([]*RunningApp, 0, len(al.runningApps))
	for _, app := range al.runningApps {
		apps = append(apps, app)
	}

	return apps
}

// IsAppRunning checks if a specific app is running
func (al *AppLauncher) IsAppRunning(serial, packageName string) bool {
	key := fmt.Sprintf("%s_%s", serial, packageName)

	al.mutex.RLock()
	defer al.mutex.RUnlock()

	_, exists := al.runningApps[key]
	return exists
}

// AddListener adds a callback for when running apps change
func (al *AppLauncher) AddListener(callback func([]*RunningApp)) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	al.listeners = append(al.listeners, callback)
}

// notifyListeners notifies all listeners of running app changes
func (al *AppLauncher) notifyListeners() {
	apps := al.GetRunningApps()

	al.mutex.RLock()
	listeners := make([]func([]*RunningApp), len(al.listeners))
	copy(listeners, al.listeners)
	al.mutex.RUnlock()

	for _, listener := range listeners {
		go listener(apps)
	}
}

// StopAllApps stops all running apps
func (al *AppLauncher) StopAllApps() {
	al.mutex.Lock()

	for key, app := range al.runningApps {
		if app.Process != nil && app.Process.Process != nil {
			app.Process.Process.Kill()
		}
		delete(al.runningApps, key)
	}

	al.mutex.Unlock()

	// Notify listeners
	al.notifyListeners()
}
