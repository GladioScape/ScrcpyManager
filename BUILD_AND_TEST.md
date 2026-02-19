# Build and Test Instructions

## Quick Start

### Option 1: Test Refactored Version (Side-by-side)

Keep both versions to compare:

```bash
# Build original version
go build -o scrcpy-manager-original.exe main.go device_manager.go app_detector.go app_launcher.go config.go window_manager.go

# Build refactored version (temporarily rename main files)
move main.go main_original_temp.go
move main_refactored.go main.go
go build -o scrcpy-manager-v2.exe
move main.go main_refactored.go
move main_original_temp.go main.go

# Now you have both:
# - scrcpy-manager-original.exe (old version)
# - scrcpy-manager-v2.exe (new refactored version)
```

### Option 2: Full Migration

```bash
# Backup original
copy main.go main_backup.go

# Use refactored version
del main.go
ren main_refactored.go main.go

# Build
go build -o scrcpy-manager-v2.exe

# Run
scrcpy-manager-v2.exe
```

## Pre-Flight Checklist

Before running the refactored version:

- [ ] ADB is installed and in PATH (or configured in app_config.json)
- [ ] Scrcpy is installed and in PATH (or configured in app_config.json)
- [ ] At least one Android device connected via USB or WiFi
- [ ] USB debugging enabled on device
- [ ] Go 1.19+ installed

## Testing Checklist

### 1. Startup Test
- [ ] Application starts without errors
- [ ] Main window appears with correct size (900x700)
- [ ] All 4 tabs are visible: Devices, Apps, Running, Settings
- [ ] Status bar shows "Ready"
- [ ] Log file created in `%TEMP%\ScrcpyManager\`

### 2. Device Detection
- [ ] Connect a device
- [ ] Click "Refresh Devices"
- [ ] Device appears in list with model name and serial
- [ ] Select device - status shows "Selected: [device name]"
- [ ] Click "Open Phone Screen" - scrcpy window opens

### 3. Apps Tab
- [ ] Select a device
- [ ] "Select a device..." message disappears
- [ ] Apps load and display
- [ ] Search icon shows/hides search bar
- [ ] Search filters apps correctly
- [ ] View mode switches (Icons/Grid/List) work
- [ ] Sort options work (Name A→Z, Z→A, Size)
- [ ] Filter options toggle package name visibility
- [ ] Launch button starts app
- [ ] Running apps show in Running tab

### 4. Running Apps Tab
- [ ] Launched apps appear in list
- [ ] Shows app name and device serial
- [ ] Stop button terminates app
- [ ] Stop All button terminates all apps
- [ ] List updates in real-time

### 5. Settings Tab
- [ ] Display dropdown populated
- [ ] Refresh displays button works
- [ ] Video quality presets selectable
- [ ] FPS entry accepts valid values (1-240)
- [ ] Checkboxes toggle properly
- [ ] Extra arguments field accepts text
- [ ] Save Settings button saves config
- [ ] Settings persist after restart

### 6. Error Handling
- [ ] Invalid FPS shows error dialog (try entering 0 or 999)
- [ ] Launching already-running app shows status message
- [ ] No device selected shows appropriate message
- [ ] ADB not found shows error
- [ ] Scrcpy not found shows error

### 7. Process Management
- [ ] Launched processes appear in task manager
- [ ] Closing app stops all scrcpy processes
- [ ] No orphaned processes remain after exit
- [ ] Process count accurate in Running tab

### 8. Logging
- [ ] Check log file: `type %TEMP%\ScrcpyManager\scrcpy-manager-*.log`
- [ ] Verify startup messages logged
- [ ] Verify device operations logged
- [ ] Verify errors logged with details

## Troubleshooting

### Build Errors

**Error: "undefined: fyne"**
```bash
go get fyne.io/fyne/v2
go mod tidy
```

**Error: "undefined: LogInfo"**
- Ensure all new files are in the same directory
- Run `go mod tidy`

**Error: "multiple main packages"**
- Only one main.go should exist
- Rename main_refactored.go to main.go if migrating

### Runtime Errors

**Error: "adb not found"**
```bash
# Check ADB path
where adb

# Or configure in app_config.json
{
  "adb_path": "C:\\path\\to\\platform-tools\\adb.exe"
}
```

**Error: "scrcpy not found"**
```bash
# Check scrcpy path
where scrcpy

# Or configure in app_config.json
{
  "scrcpy_path": "C:\\path\\to\\scrcpy\\scrcpy.exe"
}
```

**UI freezing**
- Check logs for blocking operations
- Verify async operations are working
- May indicate process deadlock - check process manager

**Processes not stopping**
- Check process manager logs
- Verify ProcessManager is being called
- Check task manager for orphaned scrcpy.exe

## Performance Testing

### Memory Test
```bash
# Monitor memory usage over time
# 1. Launch app
# 2. Connect/disconnect devices multiple times
# 3. Launch apps repeatedly
# 4. Check memory doesn't grow unbounded
```

### Stress Test
```bash
# 1. Connect multiple devices (if available)
# 2. Launch apps on all devices
# 3. Verify all processes tracked
# 4. Stop all - verify cleanup
# 5. Check logs for errors
```

## Log Analysis

View logs in real-time:
```powershell
# PowerShell
Get-Content -Path "$env:TEMP\ScrcpyManager\scrcpy-manager-*.log" -Wait -Tail 50
```

```cmd
# CMD (install tail or use PowerShell)
powershell -Command "Get-Content -Path \"$env:TEMP\ScrcpyManager\scrcpy-manager-*.log\" -Wait -Tail 50"
```

Search for errors:
```powershell
Select-String -Path "$env:TEMP\ScrcpyManager\*.log" -Pattern "ERROR"
```

## Comparison Test

Run both versions side-by-side:

| Feature | Original | Refactored | Notes |
|---------|----------|------------|-------|
| Startup time | ___ ms | ___ ms | Should be similar |
| Device scan | ___ ms | ___ ms | May be faster |
| App load | ___ ms | ___ ms | Should be faster |
| Memory usage | ___ MB | ___ MB | Should be lower |
| Process cleanup | Manual | Automatic | Big improvement |
| Error handling | None | Yes | Much better |
| Logging | No | Yes | New feature |

## Rollback Plan

If issues occur:

```bash
# Stop the new version
taskkill /F /IM scrcpy-manager-v2.exe

# Restore original
del main.go
ren main_backup.go main.go

# Rebuild original
go build -o scrcpy-manager.exe

# Run original
scrcpy-manager.exe
```

## Success Criteria

The refactored version is successful if:

✅ All features work as before
✅ No crashes or freezes
✅ Processes cleaned up properly
✅ Error messages are helpful
✅ Logs provide useful debugging info
✅ Performance is same or better
✅ Memory usage is stable

## Next Steps After Testing

1. **If successful:**
   - Delete original version
   - Rename main_refactored.go to main.go
   - Commit changes to git
   - Update README.md

2. **If issues found:**
   - Check logs for errors
   - Review REFACTORING_GUIDE.md
   - Fix issues incrementally
   - Re-test

3. **Production deployment:**
   - Build release version: `go build -ldflags="-w -s" -o scrcpy-manager.exe`
   - Test on clean system
   - Create installer if needed

## Support

If you encounter issues:

1. Check logs: `%TEMP%\ScrcpyManager\*.log`
2. Review REFACTORING_GUIDE.md
3. Compare with original implementation
4. Use `/reportbug` command for Cline issues