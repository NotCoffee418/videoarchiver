package main

// Version information - set at build time via ldflags
var (
	// Version is the application version, typically set from git tag during build
	Version = "development"

	// BuildDate is when the binary was built
	BuildDate = "unknown"

	// GitCommit is the git commit hash
	GitCommit = "unknown"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	if Version == "development" {
		return "videoarchiver (development build)"
	}
	return "videoarchiver " + Version
}
