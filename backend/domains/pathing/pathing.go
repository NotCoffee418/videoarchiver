package pathing

import (
	"errors"
	"os"
	"path/filepath"
	"videoarchiver/data"
)

func GetWorkingFile(fileName string, parts ...string) (string, error) {
	workingDir, err := GetWorkingDir(parts...)
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{workingDir}, fileName)...), nil
}

func GetWorkingDir(parts ...string) (string, error) {
	var baseDir string

	if isWindows() {
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			return "", errors.New("LOCALAPPDATA environment variable is not set")
		}
	} else if isLinux() {
		home := os.Getenv("HOME")
		if home == "" {
			return "", errors.New("HOME environment variable is not set")
		}
		baseDir = filepath.Join(home, ".local", "share")
	} else {
		return "", errors.New("unsupported platform")
	}

	targetDir := filepath.Join(append([]string{baseDir, data.AppShortName}, parts...)...)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", err
	}

	return targetDir, nil
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func isLinux() bool {
	return os.PathSeparator == '/'
}
