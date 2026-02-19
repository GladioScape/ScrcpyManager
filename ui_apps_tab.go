package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AppsTabUI manages the apps tab interface
type AppsTabUI struct {
	manager            *UIManager
	appGrid            *fyne.Container
	appScrollContainer *container.Scroll
	searchEntry        *widget.Entry
	noSelectionLabel   *widget.Label

	apps          []*AndroidApp
	sortMode      string
	currentDevice *Device
}

// NewAppsTabUI creates a new apps tab UI
func NewAppsTabUI(manager *UIManager) *AppsTabUI {
	return &AppsTabUI{
		manager:  manager,
		apps:     make([]*AndroidApp, 0),
		sortMode: "name",
	}
}

// Build constructs the apps tab UI
func (a *AppsTabUI) Build() fyne.CanvasObject {
	// Search entry
	a.searchEntry = widget.NewEntry()
	a.searchEntry.SetPlaceHolder("Search apps...")
	a.searchEntry.OnChanged = func(s string) {
		a.updateAppGrid()
		if a.appScrollContainer != nil {
			a.appScrollContainer.ScrollToTop()
		}
	}
	a.searchEntry.Hide()

	// Toolbar buttons
	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		if a.searchEntry.Hidden {
			a.searchEntry.Show()
		} else {
			a.searchEntry.Hide()
			a.searchEntry.SetText("")
			a.updateAppGrid()
		}
	})

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		a.refreshApps()
	})

	viewBtn := a.createViewModeButton()
	sortBtn := a.createSortButton()
	filterBtn := a.createFilterButton()

	toolbarRow := container.NewHBox(searchBtn, refreshBtn, viewBtn, sortBtn, filterBtn)

	// App grid
	a.appGrid = container.NewMax()
	a.appScrollContainer = container.NewVScroll(a.appGrid)
	a.appScrollContainer.SetMinSize(fyne.NewSize(400, 500))

	// No selection label
	a.noSelectionLabel = widget.NewLabel("Select a device from the Devices tab to see its apps")
	a.noSelectionLabel.Alignment = fyne.TextAlignCenter
	a.noSelectionLabel.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewBorder(
		container.NewVBox(toolbarRow, a.searchEntry),
		nil, nil, nil,
		container.NewStack(
			a.appScrollContainer,
			container.NewCenter(a.noSelectionLabel),
		),
	)
}

// OnDeviceSelected is called when a device is selected
func (a *AppsTabUI) OnDeviceSelected(device *Device) {
	a.currentDevice = device
	a.noSelectionLabel.Hide()
	a.loadAppsForDevice(device)
}

// loadAppsForDevice loads apps for the specified device
func (a *AppsTabUI) loadAppsForDevice(dev *Device) {
	a.manager.UpdateStatus(fmt.Sprintf("Loading apps for %s...", dev.Model))

	go func() {
		detectedApps, err := appDetector.GetAppsForDevice(dev.Serial)
		if err != nil {
			if appLogger != nil {
				appLogger.Error("Failed to load apps", "device", dev.Serial, "error", err)
			}
			fyne.Do(func() {
				a.manager.UpdateStatus(fmt.Sprintf("Error loading apps: %v", err))
			})
			return
		}

		a.apps = detectedApps

		fyne.Do(func() {
			a.updateAppGrid()
			a.manager.UpdateStatus(fmt.Sprintf("Loaded %d apps for %s", len(a.apps), dev.Model))
		})

		// Lazy load sizes in background
		if dev != nil {
			a.lazyComputeSizes(dev.Serial, detectedApps, SizeBatchLimit)
		}
	}()
}

// refreshApps forces a refresh of apps
func (a *AppsTabUI) refreshApps() {
	if a.currentDevice == nil {
		a.manager.UpdateStatus(StatusNoDeviceSelected)
		return
	}

	a.manager.helpers.RunAsync("RefreshApps", func() error {
		appDetector.RefreshApps(a.currentDevice.Serial)
		a.loadAppsForDevice(a.currentDevice)
		return nil
	})
}

// updateAppGrid updates the app grid display
func (a *AppsTabUI) updateAppGrid() {
	if a.appGrid == nil {
		return
	}
	a.appGrid.Objects = nil

	if len(a.apps) == 0 {
		label := widget.NewLabel("No launchable apps found")
		label.Alignment = fyne.TextAlignCenter
		a.appGrid.Add(label)
		a.appGrid.Refresh()
		return
	}

	// Filter apps
	filtered := a.filterApps(a.apps, a.searchEntry.Text)

	// Sort apps
	a.sortApps(filtered)

	// Limit display
	maxApps := len(filtered)
	if maxApps > MaxAppsToDisplay {
		maxApps = MaxAppsToDisplay
	}
	filtered = filtered[:maxApps]

	// Display based on view mode
	switch globalConfig.AppViewMode {
	case "list":
		vbox := container.NewVBox()
		for _, app := range filtered {
			vbox.Add(a.makeAppCard(app))
		}
		a.appGrid.Add(vbox)
	case "grid":
		cols := globalConfig.AppGridColumns
		if cols <= 0 {
			cols = AppGridDefaultCols
		}
		grid := container.NewGridWithColumns(cols)
		for _, app := range filtered {
			grid.Add(a.makeCompactCard(app))
		}
		a.appGrid.Add(grid)
	default: // icons
		itemSize := fyne.NewSize(IconItemWidth, IconItemHeight)
		iconGrid := container.NewGridWrap(itemSize)
		for _, app := range filtered {
			iconGrid.Add(a.makeIconItem(app))
		}
		a.appGrid.Add(iconGrid)
	}

	a.appGrid.Refresh()

	if a.appScrollContainer != nil {
		a.appScrollContainer.ScrollToTop()
	}
}

// sortApps sorts the app list based on current sort mode
func (a *AppsTabUI) sortApps(list []*AndroidApp) {
	switch a.sortMode {
	case "name":
		sort.Slice(list, func(i, j int) bool {
			return strings.ToLower(list[i].AppName) < strings.ToLower(list[j].AppName)
		})
	case "name-desc":
		sort.Slice(list, func(i, j int) bool {
			return strings.ToLower(list[i].AppName) > strings.ToLower(list[j].AppName)
		})
	case "size":
		sort.Slice(list, func(i, j int) bool {
			return list[i].SizeKB < list[j].SizeKB
		})
	case "size-desc":
		sort.Slice(list, func(i, j int) bool {
			return list[i].SizeKB > list[j].SizeKB
		})
	default:
		sort.Slice(list, func(i, j int) bool {
			return strings.ToLower(list[i].AppName) < strings.ToLower(list[j].AppName)
		})
	}
}

// filterApps filters apps based on search text
func (a *AppsTabUI) filterApps(list []*AndroidApp, search string) []*AndroidApp {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return list
	}

	out := make([]*AndroidApp, 0, len(list))
	for _, app := range list {
		name := strings.ToLower(app.AppName)
		pkg := strings.ToLower(app.PackageName)
		if strings.Contains(name, search) || strings.Contains(pkg, search) {
			out = append(out, app)
		}
	}
	return out
}

// makeAppCard creates a detailed app card
func (a *AppsTabUI) makeAppCard(app *AndroidApp) fyne.CanvasObject {
	nameText := app.AppName
	if app.SizeKB > 0 {
		nameText = fmt.Sprintf("%s  [%s]", app.AppName, formatSize(app.SizeKB))
	}

	nameLabel := widget.NewLabel(nameText)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Wrapping = fyne.TextTruncate

	var center fyne.CanvasObject
	if globalConfig.ShowPackageName {
		pkgLabel := widget.NewLabel(app.PackageName)
		pkgLabel.TextStyle = fyne.TextStyle{Monospace: true, Italic: true}
		pkgLabel.Wrapping = fyne.TextTruncate
		center = container.NewVBox(nameLabel, pkgLabel)
	} else {
		center = nameLabel
	}

	launchBtn := widget.NewButton("Launch", func() {
		a.launchApp(app)
	})
	launchBtn.Importance = widget.HighImportance

	bg := canvas.NewRectangle(color.RGBA{R: 40, G: 40, B: 50, A: 255})
	content := container.NewBorder(nil, nil, nil, launchBtn, center)

	return container.NewPadded(container.NewStack(bg, content))
}

// makeCompactCard creates a compact app card
func (a *AppsTabUI) makeCompactCard(app *AndroidApp) fyne.CanvasObject {
	titleText := app.AppName
	if app.SizeKB > 0 {
		titleText = fmt.Sprintf("%s [%s]", app.AppName, formatSize(app.SizeKB))
	}
	title := widget.NewLabel(titleText)
	title.Wrapping = fyne.TextTruncate

	btn := widget.NewButton("Launch", func() {
		a.launchApp(app)
	})

	return container.NewBorder(nil, nil, nil, btn, title)
}

// makeIconItem creates an icon-style app button
func (a *AppsTabUI) makeIconItem(app *AndroidApp) fyne.CanvasObject {
	label := app.AppName
	if app.SizeKB > 0 {
		label = fmt.Sprintf("%s\n%s", app.AppName, formatSize(app.SizeKB))
	}

	btn := widget.NewButton(label, func() {
		a.launchApp(app)
	})

	return btn
}

// launchApp launches an app with error handling
func (a *AppsTabUI) launchApp(app *AndroidApp) {
	dev := a.manager.GetSelectedDevice()
	if dev == nil {
		a.manager.UpdateStatus(StatusNoDeviceSelected)
		return
	}

	if appLauncher.IsAppRunning(dev.Serial, app.PackageName) {
		a.manager.UpdateStatus(fmt.Sprintf("%s is already running", app.AppName))
		return
	}

	a.manager.helpers.RunAsync("LaunchApp", func() error {
		err := appLauncher.LaunchApp(dev.Serial, app.PackageName, app.AppName)
		if err != nil {
			return NewAppError("LaunchApp", err, fmt.Sprintf("Failed to launch %s", app.AppName))
		}
		a.manager.UpdateStatus(fmt.Sprintf("Launched: %s", app.AppName))
		return nil
	})
}

// lazyComputeSizes computes app sizes in background
func (a *AppsTabUI) lazyComputeSizes(serial string, list []*AndroidApp, limit int) {
	count := 0
	sem := make(chan struct{}, MaxConcurrentOps)
	updated := false

	for _, app := range list {
		if app.SizeKB > 0 {
			continue
		}
		if count >= limit {
			break
		}
		count++
		sem <- struct{}{}
		go func(app *AndroidApp) {
			defer func() { <-sem }()
			if appDetector.EnsureSizeFor(serial, app) {
				updated = true
			}
		}(app)
	}

	// Wait for all to complete
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	if updated {
		fyne.Do(func() {
			a.updateAppGrid()
		})
	}
}

// Helper UI creation methods
func (a *AppsTabUI) createViewModeButton() *widget.Button {
	viewBtn := widget.NewButtonWithIcon("", theme.ListIcon(), nil)
	viewMenu := fyne.NewMenu("",
		fyne.NewMenuItem("Icons", func() { globalConfig.AppViewMode = "icons"; a.updateAppGrid() }),
		fyne.NewMenuItem("Grid", func() { globalConfig.AppViewMode = "grid"; a.updateAppGrid() }),
		fyne.NewMenuItem("List", func() { globalConfig.AppViewMode = "list"; a.updateAppGrid() }),
	)
	viewBtn.OnTapped = func() {
		pop := widget.NewPopUpMenu(viewMenu, a.manager.window.Canvas())
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(viewBtn)
		pop.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+viewBtn.Size().Height))
	}
	return viewBtn
}

func (a *AppsTabUI) createSortButton() *widget.Button {
	sortBtn := widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), nil)
	sortBtn.OnTapped = func() {
		sortMenu := fyne.NewMenu("",
			fyne.NewMenuItem("Name A→Z", func() { a.sortMode = "name"; a.updateAppGrid() }),
			fyne.NewMenuItem("Name Z→A", func() { a.sortMode = "name-desc"; a.updateAppGrid() }),
			fyne.NewMenuItem("Size ↑", func() { a.sortMode = "size"; a.updateAppGrid() }),
			fyne.NewMenuItem("Size ↓", func() { a.sortMode = "size-desc"; a.updateAppGrid() }),
		)
		pop := widget.NewPopUpMenu(sortMenu, a.manager.window.Canvas())
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(sortBtn)
		pop.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+sortBtn.Size().Height))
	}
	return sortBtn
}

func (a *AppsTabUI) createFilterButton() *widget.Button {
	filterBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), nil)
	filterBtn.OnTapped = func() {
		showPackageCheck := widget.NewCheck("Package name", func(b bool) {
			globalConfig.ShowPackageName = b
			a.updateAppGrid()
		})
		showPackageCheck.Checked = globalConfig.ShowPackageName

		showAppSizeCheck := widget.NewCheck("App size (slower)", func(b bool) {
			globalConfig.ShowAppSize = b
			if a.currentDevice != nil {
				a.refreshApps()
			} else {
				a.updateAppGrid()
			}
		})
		showAppSizeCheck.Checked = globalConfig.ShowAppSize

		content := container.NewVBox(showPackageCheck, showAppSizeCheck)
		pop := widget.NewPopUp(content, a.manager.window.Canvas())
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(filterBtn)
		pop.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+filterBtn.Size().Height))
	}
	return filterBtn
}

// formatSize formats size in KB to human-readable format
func formatSize(sizeKB int) string {
	if sizeKB <= 0 {
		return ""
	}
	if sizeKB < 1024 {
		return fmt.Sprintf("%d KB", sizeKB)
	}
	mb := float64(sizeKB) / 1024.0
	if mb < 1024 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	gb := mb / 1024.0
	return fmt.Sprintf("%.2f GB", gb)
}
