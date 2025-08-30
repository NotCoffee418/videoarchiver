package daemonsignal

import (
	"strconv"
	"time"
	"videoarchiver/backend/domains/settings"
)

// Simple communication method between daemon and UI.
// When UI makes changes, this tells daemon to restart or start the next iteration early.
type DaemonSignalService struct {
	settingsSvc *settings.SettingsService
}

func NewDaemonSignalService(settingsSvc *settings.SettingsService) *DaemonSignalService {
	return &DaemonSignalService{settingsSvc: settingsSvc}
}

// Indicates to daemon that something has changes and we must start or restart loop
func (d *DaemonSignalService) TriggerChange() error {
	return d.settingsSvc.SetPreparsed("daemon_signal", strconv.FormatInt(time.Now().Unix(), 10))
}

// Check if a recent change signal has been triggered.
func (d *DaemonSignalService) IsChangeTriggered() (bool, error) {
	valStr, err := d.settingsSvc.GetSettingString("daemon_signal")
	if err != nil {
		return false, err
	}
	unixTime, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return false, err
	}
	if unixTime > time.Now().Add(-30*time.Second).Unix() {
		return true, nil
	}
	return false, nil
}

// Acknowledge the change signal in the daemon, clear it.
func (d *DaemonSignalService) ClearChangeTrigger() error {
	return d.settingsSvc.SetPreparsed("daemon_signal", "0")
}
