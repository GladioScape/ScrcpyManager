package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DeviceTabUI manages the device tab interface
type DeviceTabUI struct {
	manager    *UIManager
	deviceList *widget.List
	devices    []*Device
}

// NewDeviceTabUI creates a new device tab UI
func NewDeviceTabUI(manager *UIManager) *DeviceTabUI {
	return &DeviceTabUI{
		manager: manager,
		devices: make([]*Device, 0),
	}
}

// Build constructs the device tab UI
func (d *DeviceTabUI) Build() fyne.CanvasObject {
	// Device list
	d.deviceList = widget.NewList(
		func() int {
			return len(d.devices)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.ComputerIcon())
			label := widget.NewLabel("Device")
			label.TextStyle = fyne.TextStyle{Bold: true}
			serialLabel := widget.NewLabel("Serial")
			serialLabel.TextStyle = fyne.TextStyle{Monospace: true}

			return container.NewBorder(
				nil, nil,
				icon,
				nil,
				container.NewVBox(label, serialLabel),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(d.devices) {
				return
			}
			dev := d.devices[id]

			border := item.(*fyne.Container)
			vbox := border.Objects[0].(*fyne.Container)

			nameLabel := vbox.Objects[0].(*widget.Label)
			serialLabel := vbox.Objects[1].(*widget.Label)

			nameLabel.SetText(dev.Model)
			serialLabel.SetText(dev.Serial)
		},
	)

	d.deviceList.OnSelected = func(id widget.ListItemID) {
		if id < len(d.devices) {
			selectedDev := d.devices[id]
			d.manager.SetSelectedDevice(selectedDev)
			d.manager.UpdateStatus(fmt.Sprintf("Selected: %s", selectedDev.Model))
		}
	}

	// Action buttons
	refreshBtn := widget.NewButton("Refresh Devices", func() {
		d.manager.RefreshDevices()
	})

	openScreenBtn := widget.NewButton("Open Phone Screen", func() {
		dev := d.manager.GetSelectedDevice()
		if dev == nil {
			d.manager.UpdateStatus(StatusNoDeviceSelected)
			return
		}

		d.manager.helpers.RunAsync("LaunchScreen", func() error {
			if err := appLauncher.LaunchScreen(dev.Serial); err != nil {
				return NewAppError("OpenScreen", err, "Failed to launch phone screen")
			}
			d.manager.UpdateStatus("Phone screen launched")
			return nil
		})
	})

	header := widget.NewLabelWithStyle("Connected Devices", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	return container.NewBorder(
		container.NewVBox(header, container.NewHBox(refreshBtn, openScreenBtn)),
		nil, nil, nil,
		container.NewPadded(d.deviceList),
	)
}

// UpdateDevices updates the device list
func (d *DeviceTabUI) UpdateDevices(devices []*Device) {
	d.devices = devices
	if d.deviceList != nil {
		fyne.Do(func() {
			d.deviceList.Refresh()
		})
	}
}

// GetDevices returns the current device list
func (d *DeviceTabUI) GetDevices() []*Device {
	return d.devices
}
