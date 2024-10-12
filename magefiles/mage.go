//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	app                  = "binary"
	defaultInstallPrefix = "/usr/local"
	envInstallPrefix     = "PROJECT_INSTALL_PREFIX"
	envTestVerbose       = "PROJECT_TEST_VERBOSE"
	envTestCover         = "PROJECT_TEST_COVER"
	envBuildRebuildAll   = "PROJECT_BUILD_REBUILD_ALL"
	envBuildVerbose      = "PROJECT_BUILD_VERBOSE"
)

var (
	Default = Build
	binary  = "./__build/" + app
)

// Test run the go tests.
// To enable verbose mode set PROJECT_TEST_VERBOSE=1.
// To enable coverage mode set PROJECT_TEST_COVER=1.
func Test() error {
	goTest := sh.RunCmd("go", "test")

	args := []string{"./..."}

	if os.Getenv(envTestVerbose) == "1" {
		args = append(args, "-v")
	}

	if os.Getenv(envTestCover) == "1" {
		args = append(args, "-cover")
	}

	return goTest(args...)
}

// Lint runs golangci-lint against the code.
func Lint() error {
	return sh.RunV("golangci-lint", "run", "--color", "always")
}

// Build build the executable.
// To rebuild packages that are already up-to-date set PROJECT_BUILD_REBUILD_ALL=1
// To enable verbose mode set PROJECT_BUILD_VERBOSE=1
func Build() error {
	main := "main.go"
	flags := ldflags()
	build := sh.RunCmd("go", "build")
	args := []string{"-ldflags=" + flags, "-o", binary}

	if os.Getenv(envBuildRebuildAll) == "1" {
		args = append(args, "-a")
	}

	if os.Getenv(envBuildVerbose) == "1" {
		args = append(args, "-v")
	}

	args = append(args, main)

	return build(args...)
}

// Install install the executable.
func Install() error {
	mg.Deps(Build)

	installPrefix := os.Getenv(envInstallPrefix)

	if installPrefix == "" {
		installPrefix = defaultInstallPrefix
	}

	dest := filepath.Join(installPrefix, "bin", app)

	if err := sh.Copy(dest, binary); err != nil {
		return fmt.Errorf("unable to install %s; %w", dest, err)
	}

	fmt.Printf("%s successfully installed to %s\n", app, dest)

	return nil
}

// Clean clean the workspace.
func Clean() error {
	if err := sh.Rm(binary); err != nil {
		return err
	}

	if err := sh.Run("go", "clean", "./..."); err != nil {
		return err
	}

	return nil
}

// ldflags returns the build flags.
func ldflags() string {
	ldflagsfmt := "-s -w -X main.binaryVersion=%s -X main.gitCommit=%s -X main.goVersion=%s -X main.buildTime=%s"
	buildTime := time.Now().UTC().Format(time.RFC3339)

	return fmt.Sprintf(ldflagsfmt, version(), gitCommit(), runtime.Version(), buildTime)
}

// version returns the latest git tag using git describe.
func version() string {
	version, err := sh.Output("git", "describe", "--tags")
	if err != nil {
		version = "N/A"
	}

	return version
}

// gitCommit returns the current git commit
func gitCommit() string {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		commit = "N/A"
	}

	return commit
}
