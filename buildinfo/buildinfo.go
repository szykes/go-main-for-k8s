package buildinfo

import (
	"log/slog"
)

var (
	appVersion     = "dev"
	commitHash     = "unknown"
	buildTimestamp = "unknown"
)

func Log(isDebugModeEnabled bool) {
	if isDebugModeEnabled {
		slog.Info("DEBUG mode is enabled")
		slog.Info("Build info", "app_version", appVersion, "commit_hash", commitHash, "build_timestamp", buildTimestamp)
	} else {
		slog.Info("Build info", "app_version", appVersion)
	}
}
