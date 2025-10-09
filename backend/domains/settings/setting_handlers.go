package settings

import (
	"fmt"
	"os"
	"runtime"
	"videoarchiver/backend/domains/logging"
	"videoarchiver/backend/domains/runner"
)

// SettingHandler handles setting changes for specific keys
type SettingHandler interface {
	HandleSettingChange(key, oldValue, newValue string, logger *logging.LogService) error
}

// AutostartServiceHandler handles the autostart_service setting
type AutostartServiceHandler struct{}

func (h *AutostartServiceHandler) HandleSettingChange(key, oldValue, newValue string, logger *logging.LogService) error {
	if logger != nil {
		logger.Info(fmt.Sprintf("Setting changed: %s = %s (was: %s)", key, newValue, oldValue))
	}

	// Convert string values to boolean
	enable := newValue == "true"

	switch runtime.GOOS {
	case "windows":
		return h.handleWindowsAutostart(enable, logger)
	case "linux":
		return h.handleLinuxAutostart(enable, logger)
	default:
		if logger != nil {
			logger.Info(fmt.Sprintf("Autostart service not supported on %s", runtime.GOOS))
		}
		return nil
	}
}

func (h *AutostartServiceHandler) handleWindowsAutostart(enable bool, logger *logging.LogService) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	registryKey := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	valueName := "Video Archiver Daemon"
	command := fmt.Sprintf(`"%s" --mode daemon`, execPath)

	if enable {
		// Add registry key
		err = runner.RunAndWait("reg", "add", registryKey, "/v", valueName, "/t", "REG_SZ", "/d", command, "/f")
		if err != nil {
			return fmt.Errorf("failed to add registry key for autostart: %w", err)
		}
		if logger != nil {
			logger.Info("Registry key created for autostart service")
		}
	} else {
		// Remove registry key
		err = runner.RunAndWait("reg", "delete", registryKey, "/v", valueName, "/f")
		if err != nil {
			// Don't fail if the key doesn't exist
			if logger != nil {
				logger.Info("Registry key for autostart service was not present or failed to remove")
			}
		} else {
			if logger != nil {
				logger.Info("Registry key removed for autostart service")
			}
		}
	}

	return nil
}

func (h *AutostartServiceHandler) handleLinuxAutostart(enable bool, logger *logging.LogService) error {
	serviceName := "video-archiver.service"

	if enable {
		err := runner.RunAndWait("systemctl", "--user", "enable", serviceName)
		if err != nil {
			return fmt.Errorf("failed to enable systemd service: %w", err)
		}
		if logger != nil {
			logger.Info("Systemd service enabled for autostart")
		}
	} else {
		err := runner.RunAndWait("systemctl", "--user", "disable", serviceName)
		if err != nil {
			// Don't fail if the service is not installed
			if logger != nil {
				logger.Info("Systemd service was not enabled or failed to disable")
			}
		} else {
			if logger != nil {
				logger.Info("Systemd service disabled for autostart")
			}
		}
	}

	return nil
}

// AutoupdateYtdlpHandler handles the autoupdate_ytdlp setting
type AutoupdateYtdlpHandler struct{}

func (h *AutoupdateYtdlpHandler) HandleSettingChange(key, oldValue, newValue string, logger *logging.LogService) error {
	if logger != nil {
		logger.Info(fmt.Sprintf("Setting changed: %s = %s (was: %s)", key, newValue, oldValue))
	}

	// This handler just logs the change - the actual logic is in ytdlp_instancer.go
	// when it checks the setting before running updates
	return nil
}

// AllowDuplicatesHandler handles the allow_duplicates setting
type AllowDuplicatesHandler struct{}

func (h *AllowDuplicatesHandler) HandleSettingChange(key, oldValue, newValue string, logger *logging.LogService) error {
	if logger != nil {
		logger.Info(fmt.Sprintf("Setting changed: %s = %s (was: %s)", key, newValue, oldValue))
	}

	// This handler just logs the change - the actual logic is in download_service.go
	// when it checks the setting before performing duplicate checks
	return nil
}

// GetSettingHandlers returns a map of setting handlers
func GetSettingHandlers() map[string]SettingHandler {
	return map[string]SettingHandler{
		"autostart_service": &AutostartServiceHandler{},
		"autoupdate_ytdlp":  &AutoupdateYtdlpHandler{},
		"allow_duplicates":  &AllowDuplicatesHandler{},
	}
}
