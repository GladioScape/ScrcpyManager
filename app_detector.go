package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// AndroidApp represents an Android application
type AndroidApp struct {
	PackageName  string
	AppName      string
	IsLaunchable bool
	SizeKB       int    // best-effort size (apk + data)
	APKPath      string // primary apk path if discoverable
}

// AppDetector detects and caches Android apps for devices
type AppDetector struct {
	appCache map[string][]*AndroidApp // serial -> apps
	mutex    sync.RWMutex
}

var appDetector *AppDetector

func init() {
	appDetector = NewAppDetector()
}

// NewAppDetector creates a new app detector
func NewAppDetector() *AppDetector {
	return &AppDetector{
		appCache: make(map[string][]*AndroidApp),
	}
}

// GetAppsForDevice retrieves all launchable apps for a device
func (ad *AppDetector) GetAppsForDevice(serial string) ([]*AndroidApp, error) {
	// Check cache first
	ad.mutex.RLock()
	if apps, exists := ad.appCache[serial]; exists {
		ad.mutex.RUnlock()
		return apps, nil
	}
	ad.mutex.RUnlock()

	// Not in cache, detect apps
	apps, err := ad.detectApps(serial)
	if err != nil {
		return nil, err
	}

	// Cache the results
	ad.mutex.Lock()
	ad.appCache[serial] = apps
	ad.mutex.Unlock()

	return apps, nil
}

// RefreshApps forces a refresh of the app list for a device
func (ad *AppDetector) RefreshApps(serial string) ([]*AndroidApp, error) {
	// Clear cache for this device
	ad.mutex.Lock()
	delete(ad.appCache, serial)
	ad.mutex.Unlock()

	// Re-detect
	return ad.GetAppsForDevice(serial)
}

// detectApps detects all launchable apps on a device
func (ad *AppDetector) detectApps(serial string) ([]*AndroidApp, error) {
	fmt.Printf("Detecting apps for device: %s\n", serial)

	// Get list of all packages
	packages, err := ad.listPackages(serial)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	// Concurrency limit
	maxWorkers := 8
	jobs := make(chan string, len(packages))
	results := make(chan *AndroidApp, len(packages))
	var active int32

	// Workers
	worker := func() {
		defer atomic.AddInt32(&active, -1)
		for pkg := range jobs {
			// Check if launchable (may be slow; keep but concurrent)
			if !ad.isPackageLaunchable(serial, pkg) {
				continue
			}

			// Fast readable name without heavy dumpsys
			appName := cleanPackageName(pkg)

			app := &AndroidApp{
				PackageName:  pkg,
				AppName:      appName,
				IsLaunchable: true,
			}
			results <- app
		}
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		atomic.AddInt32(&active, 1)
		go worker()
	}

	// Enqueue jobs
	for _, p := range packages {
		jobs <- p
	}
	close(jobs)

	// Collect
	apps := make([]*AndroidApp, 0, len(packages))
	for atomic.LoadInt32(&active) > 0 || len(results) > 0 {
		select {
		case a := <-results:
			if a != nil {
				apps = append(apps, a)
			}
		default:
			// busy wait minimal
		}
	}

	close(results)

	fmt.Printf("Found %d launchable apps on device %s\n", len(apps), serial)
	return apps, nil
}

// listPackages lists all installed packages on a device (all apps)
func (ad *AppDetector) listPackages(serial string) ([]string, error) {
	cmd := exec.Command(globalConfig.ADBPath, "-s", serial, "shell", "pm", "list", "packages")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	packages := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Lines are in format "package:com.example.app"
		if strings.HasPrefix(line, "package:") {
			pkg := strings.TrimPrefix(line, "package:")
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// isPackageLaunchable checks if a package has a launcher activity
func (ad *AppDetector) isPackageLaunchable(serial, packageName string) bool {
	// Try to query the package for launcher activities
	// Using monkey with -c android.intent.category.LAUNCHER and count 0 to check if it exists
	cmd := exec.Command(globalConfig.ADBPath, "-s", serial, "shell",
		"cmd", "package", "resolve-activity", "--brief", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If there's output and no error, it's likely launchable
	outputStr := strings.TrimSpace(string(output))
	return outputStr != "" && !strings.Contains(outputStr, "No activity found")
}

// getAppName retrieves the human-readable name of an app
func (ad *AppDetector) getAppName(serial, packageName string) string {
	// Try to get app label using dumpsys
	cmd := exec.Command(globalConfig.ADBPath, "-s", serial, "shell",
		"dumpsys", "package", packageName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse for application label
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for labelRes or applicationInfo
		if strings.Contains(line, "labelRes=0x") {
			// We have a label resource, but extracting it requires more work
			// For MVP, we'll use package name
			continue
		}

		// Some apps have the label directly in the output
		if strings.HasPrefix(line, "label=") {
			label := strings.TrimPrefix(line, "label=")
			return strings.Trim(label, "\"")
		}
	}

	// Fallback: Clean up package name to make it more readable
	return cleanPackageName(packageName)
}

// getAppSizeKB tries to estimate the app size by summing APK(s) and data dir if accessible
func (ad *AppDetector) getAppSizeKB(serial, packageName string) (string, int) {
	// Find APK path(s)
	cmd := exec.Command(globalConfig.ADBPath, "-s", serial, "shell", "pm", "path", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", 0
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var apkPaths []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "package:") {
			apkPaths = append(apkPaths, strings.TrimPrefix(l, "package:"))
		}
	}
	if len(apkPaths) == 0 {
		return "", 0
	}
	// Sum APK sizes
	totalKB := 0
	for _, p := range apkPaths {
		du := exec.Command(globalConfig.ADBPath, "-s", serial, "shell", "du", "-s", "-k", p)
		duOut, err := du.Output()
		if err == nil {
			// output: "<kb>\t<path>"
			fields := strings.Fields(string(duOut))
			if len(fields) > 0 {
				if v, perr := parseIntSafe(fields[0]); perr == nil {
					totalKB += v
				}
			}
		}
	}
	// Try to include data dir if readable (best effort)
	dataDir := "/data/data/" + packageName
	duData := exec.Command(globalConfig.ADBPath, "-s", serial, "shell", "sh", "-c", "du -s -k "+dataDir+" 2>/dev/null || true")
	if duOut, err := duData.Output(); err == nil {
		fields := strings.Fields(string(duOut))
		if len(fields) > 0 {
			if v, perr := parseIntSafe(fields[0]); perr == nil {
				totalKB += v
			}
		}
	}
	return apkPaths[0], totalKB
}

func parseIntSafe(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// icon fetching removed per user request
// EnsureSizeFor computes and sets size for an app if not present.
// Returns true if size was computed and set.
func (ad *AppDetector) EnsureSizeFor(serial string, a *AndroidApp) bool {
	if a == nil || a.SizeKB > 0 {
		return false
	}
	if apkPath, size := ad.getAppSizeKB(serial, a.PackageName); size > 0 {
		ad.mutex.Lock()
		a.APKPath = apkPath
		a.SizeKB = size
		ad.mutex.Unlock()
		return true
	}
	return false
}

// cleanPackageName converts package name to a more readable format
func cleanPackageName(pkg string) string {
	// Remove common prefixes
	pkg = strings.TrimPrefix(pkg, "com.")
	pkg = strings.TrimPrefix(pkg, "org.")
	pkg = strings.TrimPrefix(pkg, "net.")

	// Split by dots and take the last meaningful part
	parts := strings.Split(pkg, ".")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// Capitalize first letter
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + name[1:]
		}
		return name
	}

	return pkg
}

// hide prefixes feature removed per user request
