// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	permDir  os.FileMode = 0o700
	permFile os.FileMode = 0o600
)

// absolutePath returns the absolute path of the given path.
func absolutePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting the path to the user's home directory: %w", err)
		}

		path = filepath.Join(homeDir, path[1:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error calculating the absolute path: %w", err)
	}

	return absPath, nil
}

// FileExists checks whether the file or directory at the given path exists.
func FileExists(path string) (bool, error) {
	absPath, err := absolutePath(path)
	if err != nil {
		return false, fmt.Errorf("error getting the absolute path to %s: %w", path, err)
	}

	_, err = os.Stat(absPath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("error checking if the file or directory is present: %w", err)
}

// MakeDir makes a new directory at the given path.
// The permission of the new directory is set to 0700.
func MakeDir(path string) error {
	absPath, err := absolutePath(path)
	if err != nil {
		return fmt.Errorf("error getting the absolute path to %s: %w", path, err)
	}

	err = os.MkdirAll(absPath, permDir)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", path, err)
	}

	return nil
}

type incorrectDirPermError struct {
	path string
}

func NewIncorrectDirPermError(path string) incorrectDirPermError {
	return incorrectDirPermError{path: path}
}

func (e incorrectDirPermError) Error() string {
	return fmt.Sprintf("the permission for the directory at %s is not set to 0700", e.path)
}

// CheckDirPerm checks to see if the permission of the directory at the given path is set to 0700.
func CheckDirPerm(path string) error {
	absPath, err := absolutePath(path)
	if err != nil {
		return fmt.Errorf("error getting the absolute path to %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("error retrieving the file info for %s: %w", path, err)
	}

	if info.Mode().Perm() != permDir {
		return NewIncorrectDirPermError(path)
	}

	return nil
}

type incorrectFilePermError struct {
	path string
}

func NewIncorrectFilePermError(path string) incorrectFilePermError {
	return incorrectFilePermError{path: path}
}

func (e incorrectFilePermError) Error() string {
	return fmt.Sprintf("the permission for the file at %s is not set to 0600", e.path)
}

// CheckFilePerm checks to see if the permission of the file at the given path is set to 0600.
func CheckFilePerm(path string) error {
	absPath, err := absolutePath(path)
	if err != nil {
		return fmt.Errorf("error getting the absolute path to %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("error retrieving the file info for %s: %w", path, err)
	}

	if info.Mode().Perm() != permFile {
		return NewIncorrectFilePermError(path)
	}

	return nil
}
