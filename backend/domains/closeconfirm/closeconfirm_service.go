package closeconfirm

import (
	"context"
	"fmt"
	"videoarchiver/backend/domains/logging"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CloseConfirmService handles application close confirmation
type CloseConfirmService struct {
	ctx     context.Context
	enabled bool
	logger  *logging.LogService
}

// NewCloseConfirmService creates a new close confirmation service
func NewCloseConfirmService(ctx context.Context, logger *logging.LogService) *CloseConfirmService {
	return &CloseConfirmService{
		ctx:     ctx,
		enabled: false, // Default to disabled
		logger:  logger,
	}
}

// IsEnabled returns whether close confirmation is enabled
func (c *CloseConfirmService) IsEnabled() bool {
	return c.enabled
}

// SetEnabled sets whether close confirmation is enabled
func (c *CloseConfirmService) SetEnabled(enabled bool) {
	c.enabled = enabled
	if c.logger != nil {
		c.logger.Debug(fmt.Sprintf("Close confirmation enabled: %v", enabled))
	}
}

// ShouldConfirmClose checks if close confirmation dialog should be shown
// and shows it if needed. Returns true if application should close, false otherwise.
func (c *CloseConfirmService) ShouldConfirmClose() bool {
	if !c.enabled {
		// Confirmation disabled, allow close
		return true
	}

	// Show confirmation dialog with "No" as default
	result, err := runtime.MessageDialog(c.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Confirm Close",
		Message:       "Are you sure you want to close this application?",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})

	if err != nil {
		if c.logger != nil {
			c.logger.Error(fmt.Sprintf("Failed to show close confirmation dialog: %v", err))
		}
		// If dialog fails, allow close as fallback
		return true
	}

	// Only allow close if user explicitly clicked "Yes"
	return result == "Yes"
}
