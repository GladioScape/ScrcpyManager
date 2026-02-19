# Scrcpy Link (ScrcpyManager)

**A modern, cross-platform GUI application for managing Android devices using scrcpy**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Fyne](https://img.shields.io/badge/Fyne-v2-blue?style=flat)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](https://github.com/GladioScape/ScrcpyManager)

## Description

Scrcpy Link is a powerful desktop application that provides an intuitive graphical interface for managing Android devices via ADB (Android Debug Bridge) and scrcpy. It allows you to browse, launch, and manage Android applications directly from your computer, mirror your phone screen, and monitor running processes—all from a single, user-friendly interface.

## Features

### 📱 Device Management
- **Automatic Device Detection**: Automatically detects connected Android devices via USB or Wi-Fi
- **Real-time Monitoring**: Continuously monitors device connection status
- **Multiple Device Support**: Connect and switch between multiple Android devices
- **Device Information Display**: View device model and serial number

### 📦 App Management
- **App Browser**: Browse all installed launchable applications on your Android device
- **Search & Filter**: Quickly find apps by name or package name
- **Multiple View Modes**: 
  - Icons view (compact grid)
  - Grid view (with customizable columns)
  - List view (detailed information)
- **Sorting Options**: Sort apps by name (A-Z or Z-A) or size (ascending/descending)
- **App Size Display**: Optional display of application storage size
- **One-Click Launch**: Launch any Android app directly from your desktop

### 🖥️ Screen Mirroring
- **Phone Screen Mirroring**: Mirror your Android device screen using scrcpy
- **Configurable Quality Settings**:
  - Adjustable video bitrate (500 Kbps - 50 Mbps)
  - Configurable maximum FPS (1-120)
  - Custom display resolution

### 🔄 Running Apps Monitor
- **Process Tracking**: View all apps launched through Scrcpy Link
- **Easy Termination**: Stop running apps with a single click
- **Real-time Updates**: Automatically updates when apps start or stop

### ⚙️ Settings & Configuration
- **Customizable Paths**: Configure ADB and scrcpy executable paths
- **Display Preferences**: Set default display width and height
- **Performance Tuning**: Adjust bitrate and FPS for optimal performance
- **Persistent Settings**: All configurations are automatically saved

### 🎨 User Interface
- **Modern Dark Theme**: Clean, modern dark theme interface
- **Tabbed Navigation**: Easy navigation between Devices, Apps, Running, and Settings
- **Status Bar**: Real-time status updates and feedback
- **Async Operations**: Non-blocking UI with background task processing

## System Requirements

### Supported Operating Systems

| Operating System | Version | Status | Notes |
|-----------------|---------|--------|-------|
| **Windows 11** | All versions | ✅ Fully Supported | Recommended |
| **Windows 10** | Version 1809+ (October 2018 Update) | ✅ Fully Supported | Minimum required |
| **Windows 10** | Older builds (< 1809) | ⚠️ May Work | Not officially supported |
| **Windows 8.1** | - | ❌ Not Supported | Go runtime incompatible |
| **Windows 8** | - | ❌ Not Supported | Go runtime incompatible |
| **Windows 7** | - | ❌ Not Supported | Go runtime incompatible |
| **Windows Vista/XP** | - | ❌ Not Supported | Impossible to run |
| **macOS** | 10.15 (Catalina)+ | ✅ Fully Supported | |
| **Linux** | Modern distributions | ✅ Fully Supported | Requires OpenGL 2.0+ |

### Why Windows 10 (1809) Minimum?

This application is built with:
- **Go 1.21+**: The Go programming language dropped support for Windows 7, 8, and 8.1 starting with Go 1.21 (August 2023). Modern Go requires Windows 10 version 1809 or later.
- **Fyne v2**: The GUI framework uses modern Windows APIs and OpenGL rendering that require Windows 10+.
- **scrcpy**: The underlying screen mirroring tool recommends Windows 10+ for best compatibility.

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **CPU** | x64 processor (AMD64) | Modern multi-core processor |
| **RAM** | 4 GB | 8 GB or more |
| **Storage** | 100 MB free space | 500 MB+ (for logs and cache) |
| **Display** | 1280x720 | 1920x1080 or higher |
| **GPU** | OpenGL 2.0 compatible | OpenGL 3.0+ for better performance |
| **USB** | USB 2.0 port | USB 3.0 for faster file transfers |

### Architecture Support

| Architecture | Windows | macOS | Linux |
|-------------|---------|-------|-------|
| **x64 (AMD64)** | ✅ | ✅ | ✅ |
| **x86 (32-bit)** | ⚠️ Can be built | ❌ | ⚠️ Can be built |
| **ARM64** | ✅ | ✅ (Apple Silicon) | ✅ |

### Software Dependencies

| Dependency | Required | How to Install |
|------------|----------|----------------|
| **ADB** | ✅ Yes | [Android SDK Platform Tools](https://developer.android.com/studio/releases/platform-tools) or included with Android Studio |
| **scrcpy** | ✅ Yes | [scrcpy releases](https://github.com/Genymobile/scrcpy/releases) or via package manager |
| **Go 1.21+** | Only for building | [golang.org](https://golang.org/dl/) |

### Android Device Requirements

- **Android Version**: 5.0 (Lollipop) or higher
- **USB Debugging**: Must be enabled in Developer Options
- **USB Cable**: For wired connection (data cable, not charge-only)
- **Wi-Fi**: For wireless debugging (Android 11+ for native wireless ADB)

### Network Requirements (for Wireless ADB)

- Device and computer must be on the same network
- Port 5555 (or custom) must be accessible
- Firewall may need configuration to allow ADB connections

## Installation

### Pre-built Binaries

Download the latest release from the [Releases](https://github.com/GladioScape/ScrcpyManager/releases) page.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/GladioScape/ScrcpyManager.git
cd ScrcpyManager

# Install dependencies
go mod download

# Build the application
go build -o scrcpy-manager.exe

# Run
./scrcpy-manager.exe
```

## Usage

1. **Connect Your Device**
   - Enable USB debugging on your Android device
   - Connect via USB cable or set up wireless debugging
   - The device will appear automatically in the Devices tab

2. **Select a Device**
   - Click on a device in the Devices list to select it
   - The status bar will confirm your selection

3. **Browse Apps**
   - Navigate to the Apps tab
   - Browse, search, or filter installed applications
   - Click "Launch" to start an app with screen mirroring

4. **Monitor Running Apps**
   - Check the Running tab to see active app sessions
   - Stop any running app by clicking the stop button

5. **Configure Settings**
   - Adjust scrcpy settings in the Settings tab
   - Configure video quality, resolution, and paths

## Configuration

Configuration is stored in a JSON file and includes:

| Setting | Description | Default |
|---------|-------------|---------|
| `adbPath` | Path to ADB executable | `adb` |
| `scrcpyPath` | Path to scrcpy executable | `scrcpy` |
| `scrcpyBitrate` | Video bitrate in bps | `8000000` (8 Mbps) |
| `scrcpyMaxFPS` | Maximum frames per second | `60` |
| `defaultDisplayWidth` | Display width in pixels | `1920` |
| `defaultDisplayHeight` | Display height in pixels | `1080` |
| `appViewMode` | App display mode (`icons`/`grid`/`list`) | `icons` |
| `showPackageName` | Show package names in app list | `false` |
| `showAppSize` | Calculate and show app sizes | `false` |

## Project Structure

```
ScrcpyManager/
├── main_refactored.go    # Application entry point and main logic
├── ui_device_tab.go      # Device management UI
├── ui_apps_tab.go        # App browser UI
├── ui_running_tab.go     # Running apps monitor UI
├── ui_settings_tab.go    # Settings panel UI
├── ui_helpers.go         # UI utility functions
├── process_manager.go    # Process lifecycle management
├── window_manager.go     # Window management utilities
├── logger.go             # Logging system
├── config.go             # Configuration management
├── .signpath/            # Code signing configuration
└── README.md             # This file
```

## Code Signing

This project is configured for code signing using SignPath.io. See [SIGNPATH_SETUP.md](SIGNPATH_SETUP.md) for details on setting up code signing for releases.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [scrcpy](https://github.com/Genymobile/scrcpy) - The amazing screen mirroring tool
- [Fyne](https://fyne.io/) - Cross-platform GUI toolkit for Go
- [ADB](https://developer.android.com/studio/command-line/adb) - Android Debug Bridge

## Support

If you encounter any issues or have questions:
- Open an [issue](https://github.com/GladioScape/ScrcpyManager/issues) on GitHub
- Check existing issues for solutions

---

**Made with ❤️ by [GladioScape](https://github.com/GladioScape)**