package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Device represents a connected Android device
type Device struct {
	Serial       string
	Model        string
	AndroidVer   string
	IsWireless   bool
	IsAuthorized bool
	LastSeen     time.Time
}

// DeviceManager manages connected Android devices
type DeviceManager struct {
	devices      map[string]*Device
	mutex        sync.RWMutex
	listeners    []func([]*Device)
	stopMonitor  chan bool
	isMonitoring bool
}

var deviceManager *DeviceManager

func init() {
	deviceManager = NewDeviceManager()
}

// NewDeviceManager creates a new device manager
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices:     make(map[string]*Device),
		listeners:   make([]func([]*Device), 0),
		stopMonitor: make(chan bool),
	}
}

// StartMonitoring begins monitoring for device connections/disconnections
func (dm *DeviceManager) StartMonitoring() {
	if dm.isMonitoring {
		return
	}

	dm.isMonitoring = true
	go dm.monitorLoop()
}

// StopMonitoring stops the device monitoring
func (dm *DeviceManager) StopMonitoring() {
	if !dm.isMonitoring {
		return
	}

	dm.stopMonitor <- true
	dm.isMonitoring = false
}

// monitorLoop continuously checks for device changes
func (dm *DeviceManager) monitorLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Initial scan
	dm.scanDevices()

	for {
		select {
		case <-ticker.C:
			dm.scanDevices()
		case <-dm.stopMonitor:
			return
		}
	}
}

// scanDevices scans for connected devices using ADB
func (dm *DeviceManager) scanDevices() {
	cmd := exec.Command(globalConfig.ADBPath, "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error scanning devices: %v\n", err)
		return
	}

	newDevices := parseDeviceList(string(output))

	dm.mutex.Lock()

	// Track which devices are still connected
	currentSerials := make(map[string]bool)

	// Add or update devices
	changed := false
	for _, dev := range newDevices {
		currentSerials[dev.Serial] = true

		if existing, exists := dm.devices[dev.Serial]; exists {
			// Update existing device
			existing.LastSeen = time.Now()
			if existing.Model != dev.Model || existing.AndroidVer != dev.AndroidVer {
				existing.Model = dev.Model
				existing.AndroidVer = dev.AndroidVer
				changed = true
			}
		} else {
			// New device
			dev.LastSeen = time.Now()
			dm.devices[dev.Serial] = dev
			changed = true
			fmt.Printf("Device connected: %s (%s)\n", dev.Serial, dev.Model)
		}
	}

	// Remove disconnected devices
	for serial := range dm.devices {
		if !currentSerials[serial] {
			delete(dm.devices, serial)
			changed = true
			fmt.Printf("Device disconnected: %s\n", serial)
		}
	}

	dm.mutex.Unlock()

	// Notify listeners if devices changed
	if changed {
		dm.notifyListeners()
	}
}

// parseDeviceList parses the output of "adb devices -l"
func parseDeviceList(output string) []*Device {
	devices := make([]*Device, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header and empty lines
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}

		// Parse line: "serial device/unauthorized model:... device:..."
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		serial := parts[0]
		status := parts[1]

		device := &Device{
			Serial:       serial,
			IsWireless:   strings.Contains(serial, ":"),
			IsAuthorized: status == "device",
		}

		// Parse additional info (model, device name, etc.)
		for i := 2; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "model:") {
				device.Model = strings.TrimPrefix(parts[i], "model:")
				device.Model = strings.ReplaceAll(device.Model, "_", " ")
			}
		}

		// If no model found, use serial
		if device.Model == "" {
			device.Model = serial
		}

		devices = append(devices, device)
	}

	return devices
}

// GetDevices returns a list of all connected devices
func (dm *DeviceManager) GetDevices() []*Device {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	devices := make([]*Device, 0, len(dm.devices))
	for _, dev := range dm.devices {
		devices = append(devices, dev)
	}

	return devices
}

// GetDevice returns a specific device by serial
func (dm *DeviceManager) GetDevice(serial string) *Device {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	return dm.devices[serial]
}

// AddListener adds a callback that will be called when devices change
func (dm *DeviceManager) AddListener(callback func([]*Device)) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.listeners = append(dm.listeners, callback)
}

// notifyListeners notifies all listeners of device changes
func (dm *DeviceManager) notifyListeners() {
	devices := dm.GetDevices()

	dm.mutex.RLock()
	listeners := make([]func([]*Device), len(dm.listeners))
	copy(listeners, dm.listeners)
	dm.mutex.RUnlock()

	for _, listener := range listeners {
		go listener(devices)
	}
}
