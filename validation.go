package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// Validator provides input validation functions
type Validator struct{}

var validator = &Validator{}

// ValidateBitrate validates bitrate format (e.g., "8M", "2000000")
func (v *Validator) ValidateBitrate(bitrate int) error {
	if bitrate < 100000 {
		return NewAppError("ValidateBitrate", ErrInvalidBitrate,
			"bitrate must be at least 100000 (100 Kbps)")
	}
	if bitrate > 100000000 {
		return NewAppError("ValidateBitrate", ErrInvalidBitrate,
			"bitrate must not exceed 100000000 (100 Mbps)")
	}
	return nil
}

// ValidateFPS validates frames per second value
func (v *Validator) ValidateFPS(fps int) error {
	if fps < 1 || fps > 240 {
		return NewAppError("ValidateFPS", ErrInvalidFPS,
			"fps must be between 1 and 240")
	}
	return nil
}

// ValidateDisplayID validates display ID
func (v *Validator) ValidateDisplayID(displayID int) error {
	if displayID < 0 || displayID > 10 {
		return NewAppError("ValidateDisplayID", ErrInvalidDisplayID,
			"display ID must be between 0 and 10")
	}
	return nil
}

// ValidateResolution validates width and height
func (v *Validator) ValidateResolution(width, height int) error {
	if width < 100 || width > 7680 {
		return fmt.Errorf("width must be between 100 and 7680")
	}
	if height < 100 || height > 4320 {
		return fmt.Errorf("height must be between 100 and 4320")
	}
	return nil
}

// ValidateSerialFormat validates device serial format
func (v *Validator) ValidateSerialFormat(serial string) error {
	if serial == "" {
		return fmt.Errorf("serial cannot be empty")
	}
	// Basic validation - serials are typically alphanumeric with colons/dots
	matched, _ := regexp.MatchString(`^[A-Za-z0-9.:_-]+$`, serial)
	if !matched {
		return fmt.Errorf("invalid serial format")
	}
	return nil
}

// ValidatePackageName validates Android package name format
func (v *Validator) ValidatePackageName(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	// Android package name format: com.example.app
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`, pkg)
	if !matched {
		return fmt.Errorf("invalid package name format")
	}
	return nil
}

// ParseBitrateString parses bitrate strings like "8M" or "2000000"
func (v *Validator) ParseBitrateString(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	// Check for suffix (K, M)
	matched, _ := regexp.MatchString(`^\d+[KM]?$`, s)
	if !matched {
		return 0, fmt.Errorf("invalid bitrate format (use format like '8M' or '2000000')")
	}

	suffix := s[len(s)-1]
	if suffix == 'K' || suffix == 'M' {
		numStr := s[:len(s)-1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, err
		}
		if suffix == 'K' {
			return num * 1000, nil
		}
		return num * 1000000, nil
	}

	return strconv.Atoi(s)
}

// Convenience functions
func ValidateBitrate(bitrate int) error {
	return validator.ValidateBitrate(bitrate)
}

func ValidateFPS(fps int) error {
	return validator.ValidateFPS(fps)
}

func ValidateDisplayID(displayID int) error {
	return validator.ValidateDisplayID(displayID)
}

func ValidateResolution(width, height int) error {
	return validator.ValidateResolution(width, height)
}

func ValidateSerialFormat(serial string) error {
	return validator.ValidateSerialFormat(serial)
}

func ValidatePackageName(pkg string) error {
	return validator.ValidatePackageName(pkg)
}
