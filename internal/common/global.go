package common

import (
	"pornboss/internal/manager"

	"gorm.io/gorm"
)

// Shared application-wide dependencies.
var (
	DB                *gorm.DB
	ScreenshotManager *manager.ScreenshotManager
	CoverManager      *manager.CoverManager
	AppConfig         *Config
)
