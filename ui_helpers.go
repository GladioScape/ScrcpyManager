package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// UIHelpers provides common UI utility functions
type UIHelpers struct {
	window fyne.Window
}

// NewUIHelpers creates a new UI helpers instance
func NewUIHelpers(window fyne.Window) *UIHelpers {
	return &UIHelpers{window: window}
}

// ShowError displays an error dialog
func (h *UIHelpers) ShowError(err error) {
	if err == nil {
		return
	}

	Log(LogError, "Showing error dialog: %v", err)

	var message string
	if appErr, ok := err.(*AppError); ok {
		message = appErr.Message
		if message == "" {
			message = appErr.Err.Error()
		}
	} else {
		message = err.Error()
	}

	dialog.ShowError(err, h.window)
}

// ShowInfo displays an information dialog
func (h *UIHelpers) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, h.window)
}

// ShowConfirm displays a confirmation dialog
func (h *UIHelpers) ShowConfirm(title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, h.window)
}

// ShowSuccess displays a success message in the status bar
func (h *UIHelpers) ShowSuccess(message string) {
	updateStatus(message)
	Log(LogInfo, "%s", message)
}

// UpdateStatus updates the status bar with a message
func (h *UIHelpers) UpdateStatus(message string) {
	updateStatus(message)
}

// SafeRefresh safely refreshes a widget on the UI thread
func (h *UIHelpers) SafeRefresh(obj fyne.CanvasObject) {
	if obj == nil {
		return
	}
	fyne.CurrentApp().Driver().CanvasForObject(obj).Refresh(obj)
}

// RunAsync runs a function asynchronously and handles errors
func (h *UIHelpers) RunAsync(operation string, fn func() error) {
	go func() {
		Log(LogDebug, "Starting async operation: %s", operation)
		if err := fn(); err != nil {
			Log(LogError, "Async operation failed: %s, error: %v", operation, err)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: err.Error(),
			})
		} else {
			Log(LogDebug, "Async operation completed: %s", operation)
		}
	}()
}

// WithLoading shows a loading indicator while executing a function
func (h *UIHelpers) WithLoading(message string, fn func() error) {
	updateStatus(message + "...")

	go func() {
		err := fn()
		if err != nil {
			h.ShowError(err)
		}
	}()
}
