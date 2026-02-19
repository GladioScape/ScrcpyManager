package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// ProcessInfo holds information about a running process
type ProcessInfo struct {
	PID         int
	Serial      string
	PackageName string
	AppName     string
	Cmd         *exec.Cmd
	StartTime   time.Time
	ctx         context.Context
	cancel      context.CancelFunc
}

// ProcessManager manages running scrcpy/app processes
type ProcessManager struct {
	mu        sync.RWMutex
	processes map[string]*ProcessInfo // key: serial:package
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*ProcessInfo),
	}
}

// StartProcess starts and tracks a new process
func (pm *ProcessManager) StartProcess(serial, packageName, appName string, cmd *exec.Cmd) error {
	key := pm.makeKey(serial, packageName)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already running
	if existing, ok := pm.processes[key]; ok {
		if existing.Cmd.Process != nil {
			return NewAppError("StartProcess", ErrAppAlreadyRunning,
				fmt.Sprintf("%s is already running", appName))
		}
		// Clean up stale entry
		delete(pm.processes, key)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ProcessStartTimeout)

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		return NewAppError("StartProcess", err, fmt.Sprintf("failed to start %s", appName))
	}

	// Track the process
	info := &ProcessInfo{
		PID:         cmd.Process.Pid,
		Serial:      serial,
		PackageName: packageName,
		AppName:     appName,
		Cmd:         cmd,
		StartTime:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
	}

	pm.processes[key] = info

	if appLogger != nil {
		appLogger.Info("Process started", "app", appName, "serial", serial, "pid", info.PID)
	}

	// Monitor process in background
	go pm.monitorProcess(key, info)

	return nil
}

// StopProcess stops a tracked process
func (pm *ProcessManager) StopProcess(serial, packageName string) error {
	key := pm.makeKey(serial, packageName)

	pm.mu.Lock()
	info, ok := pm.processes[key]
	if !ok {
		pm.mu.Unlock()
		return NewAppError("StopProcess", ErrProcessNotFound, "process not found")
	}
	delete(pm.processes, key)
	pm.mu.Unlock()

	return pm.killProcess(info)
}

// StopAllForDevice stops all processes for a specific device
func (pm *ProcessManager) StopAllForDevice(serial string) error {
	pm.mu.Lock()
	toStop := make([]*ProcessInfo, 0)
	for key, info := range pm.processes {
		if info.Serial == serial {
			toStop = append(toStop, info)
			delete(pm.processes, key)
		}
	}
	pm.mu.Unlock()

	var lastErr error
	for _, info := range toStop {
		if err := pm.killProcess(info); err != nil {
			lastErr = err
			if appLogger != nil {
				appLogger.Error("Failed to stop process", "app", info.AppName, "error", err)
			}
		}
	}

	return lastErr
}

// StopAll stops all tracked processes
func (pm *ProcessManager) StopAll() error {
	pm.mu.Lock()
	toStop := make([]*ProcessInfo, 0, len(pm.processes))
	for _, info := range pm.processes {
		toStop = append(toStop, info)
	}
	pm.processes = make(map[string]*ProcessInfo)
	pm.mu.Unlock()

	var lastErr error
	for _, info := range toStop {
		if err := pm.killProcess(info); err != nil {
			lastErr = err
			if appLogger != nil {
				appLogger.Error("Failed to stop process", "app", info.AppName, "error", err)
			}
		}
	}

	return lastErr
}

// IsRunning checks if a process is running
func (pm *ProcessManager) IsRunning(serial, packageName string) bool {
	key := pm.makeKey(serial, packageName)
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, ok := pm.processes[key]
	if !ok {
		return false
	}

	// Check if process is actually running
	if info.Cmd.Process == nil {
		return false
	}

	return true
}

// GetRunningProcesses returns a list of all running processes
func (pm *ProcessManager) GetRunningProcesses() []*ProcessInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*ProcessInfo, 0, len(pm.processes))
	for _, info := range pm.processes {
		result = append(result, info)
	}

	return result
}

// GetProcessCount returns the number of running processes
func (pm *ProcessManager) GetProcessCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.processes)
}

// monitorProcess monitors a process and cleans up when it exits
func (pm *ProcessManager) monitorProcess(key string, info *ProcessInfo) {
	// Wait for process to exit
	err := info.Cmd.Wait()

	// Cleanup
	info.cancel()

	pm.mu.Lock()
	delete(pm.processes, key)
	pm.mu.Unlock()

	if err != nil {
		if appLogger != nil {
			appLogger.Warning("Process exited with error", "app", info.AppName, "error", err)
		}
	} else {
		if appLogger != nil {
			appLogger.Info("Process exited normally", "app", info.AppName)
		}
	}
}

// killProcess kills a process gracefully
func (pm *ProcessManager) killProcess(info *ProcessInfo) error {
	if info.Cmd.Process == nil {
		return nil
	}

	if appLogger != nil {
		appLogger.Info("Stopping process", "app", info.AppName, "pid", info.PID)
	}

	// Cancel context
	info.cancel()

	// Try to kill the process
	err := info.Cmd.Process.Kill()
	if err != nil {
		return NewAppError("killProcess", err, fmt.Sprintf("failed to kill process %d", info.PID))
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- info.Cmd.Wait()
	}()

	select {
	case <-done:
		if appLogger != nil {
			appLogger.Info("Process stopped", "app", info.AppName)
		}
		return nil
	case <-time.After(ProcessStopTimeout):
		return fmt.Errorf("timeout waiting for process to stop")
	}
}

// makeKey creates a unique key for a process
func (pm *ProcessManager) makeKey(serial, packageName string) string {
	return fmt.Sprintf("%s:%s", serial, packageName)
}

// Global process manager instance
var processManager *ProcessManager

func init() {
	processManager = NewProcessManager()
}
