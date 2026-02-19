package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// RunningTabUI manages the running apps tab interface
type RunningTabUI struct {
	manager         *UIManager
	runningAppsList *widget.List
	runningApps     []*RunningApp
}

// NewRunningTabUI creates a new running apps tab UI
func NewRunningTabUI(manager *UIManager) *RunningTabUI {
	return &RunningTabUI{
		manager:     manager,
		runningApps: make([]*RunningApp, 0),
	}
}

// Build constructs the running apps tab UI
func (r *RunningTabUI) Build() fyne.CanvasObject {
	r.runningAppsList = widget.NewList(
		func() int {
			return len(r.runningApps)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("App Name")
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			deviceLabel := widget.NewLabel("Device")
			stopBtn := widget.NewButton("Stop", nil)
			stopBtn.Importance = widget.DangerImportance

			return container.NewBorder(
				nil, nil, nil,
				stopBtn,
				container.NewVBox(nameLabel, deviceLabel),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(r.runningApps) {
				return
			}
			app := r.runningApps[id]

			border := item.(*fyne.Container)
			vbox := border.Objects[0].(*fyne.Container)
			stopBtn := border.Objects[1].(*widget.Button)

			nameLabel := vbox.Objects[0].(*widget.Label)
			deviceLabel := vbox.Objects[1].(*widget.Label)

			nameLabel.SetText(app.AppName)
			deviceLabel.SetText(fmt.Sprintf("Device: %s", app.Serial))

			stopBtn.OnTapped = func() {
				r.stopApp(app)
			}
		},
	)

	stopAllBtn := widget.NewButton("Stop All", func() {
		r.stopAllApps()
	})
	stopAllBtn.Importance = widget.DangerImportance

	return container.NewBorder(
		container.NewVBox(stopAllBtn),
		nil, nil, nil,
		container.NewPadded(r.runningAppsList),
	)
}

// UpdateRunningApps updates the running apps list
func (r *RunningTabUI) UpdateRunningApps(apps []*RunningApp) {
	r.runningApps = apps
	if r.runningAppsList != nil {
		fyne.Do(func() {
			r.runningAppsList.Refresh()
		})
	}
}

// stopApp stops a specific app
func (r *RunningTabUI) stopApp(app *RunningApp) {
	r.manager.helpers.RunAsync("StopApp", func() error {
		if appLogger != nil {
			appLogger.Info("Stopping app", "app", app.AppName, "serial", app.Serial)
		}
		appLauncher.StopApp(app.Serial, app.PackageName)
		r.manager.UpdateStatus(fmt.Sprintf("Stopped: %s", app.AppName))
		return nil
	})
}

// stopAllApps stops all running apps
func (r *RunningTabUI) stopAllApps() {
	r.manager.helpers.RunAsync("StopAllApps", func() error {
		if appLogger != nil {
			appLogger.Info("Stopping all apps")
		}
		appLauncher.StopAllApps()
		r.manager.UpdateStatus(StatusAllAppsStopped)
		return nil
	})
}
