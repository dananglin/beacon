// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	projectName          = "beacon"
	defaultAppName       = projectName
	defaultInstallPrefix = "/usr/local"
	defaultDockerfile    = "Dockerfile"

	envInstallPrefix    = "BEACON_INSTALL_PREFIX"
	envTestVerbose      = "BEACON_TEST_VERBOSE"
	envTestCover        = "BEACON_TEST_COVER"
	envBuildRebuildAll  = "BEACON_BUILD_REBUILD_ALL"
	envBuildVerbose     = "BEACON_BUILD_VERBOSE"
	envFailOnFormatting = "BEACON_FAIL_ON_FORMATTING"
	envAppName          = "BEACON_APP_NAME"
	envDockerImageName  = "BEACON_DOCKER_IMAGE_NAME"
	envDockerfile       = "BEACON_DOCKERFILE"
)

var Default = Build

// Test run the go tests.
// To enable verbose mode set BEACON_TEST_VERBOSE=1.
// To enable coverage mode set BEACON_TEST_COVER=1.
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

// Gosec runs gosec against the code.
func Gosec() error {
	return sh.RunV("gosec", "./...")
}

// Staticcheck runs staticcheck against the code.
func Staticcheck() error {
	return sh.RunV("staticcheck", "./...")
}

// Gofmt checks the code for formatting.
// To fail on formatting set BEACON_FAIL_ON_FORMATTING=1
func Gofmt() error {
	output, err := sh.Output("go", "fmt", "./...")
	if err != nil {
		return err
	}

	files := make([]string, 0)

	if output != "" {
		files = strings.Split(output, "\n")
	}

	filesToFormat := ""

	for _, file := range files {
		filesToFormat += "\n- " + file
	}

	if os.Getenv(envFailOnFormatting) != "1" {
		fmt.Println(filesToFormat)

		return nil
	}

	if filesToFormat != "" {
		return fmt.Errorf("The following files needs to be formatted: %s", filesToFormat)
	}

	return nil
}

// Govet runs go vet against the code.
func Govet() error {
	return sh.RunV("go", "vet", "./...")
}

// Build build the executable.
// To rebuild packages that are already up-to-date set BEACON_BUILD_REBUILD_ALL=1
// To enable verbose mode set BEACON_BUILD_VERBOSE=1
func Build() error {
	fmt.Println("Building the binary...")

	main := "./cmd/" + projectName
	flags := ldflags()
	build := sh.RunCmd("go", "build")
	binary := filepath.Join("./__build", appName())
	args := []string{"-ldflags=" + flags, "-o", binary}

	if os.Getenv(envBuildRebuildAll) == "1" {
		args = append(args, "-a")
	}

	if os.Getenv(envBuildVerbose) == "1" {
		args = append(args, "-v")
	}

	args = append(args, main)

	if err := build(args...); err != nil {
		return fmt.Errorf("error building the binary: %w", err)
	}

	fmt.Printf("Successfully built the binary to %q.\n", binary)

	return nil
}

// Install install the executable.
func Install() error {
	mg.Deps(Build)

	installPrefix := os.Getenv(envInstallPrefix)
	app := appName()
	binary := filepath.Join("./__build", app)

	if installPrefix == "" {
		installPrefix = defaultInstallPrefix
	}

	dest := filepath.Join(installPrefix, "bin", appName())

	fmt.Printf("Installing the binary to %s\n", dest)

	if err := sh.Copy(dest, binary); err != nil {
		return fmt.Errorf("unable to install %s; %w", dest, err)
	}

	fmt.Printf("%s successfully installed to %s\n", app, dest)

	return nil
}

// Clean clean the workspace.
func Clean() error {
	fmt.Println("Cleaning the workspace...")

	binary := filepath.Join("./__build", appName())

	if err := sh.Rm(binary); err != nil {
		return err
	}

	if err := sh.Run("go", "clean", "./..."); err != nil {
		return err
	}

	fmt.Println("Workspace cleaned.")

	return nil
}

// Docker builds the docker image.
// Use BEACON_DOCKER_IMAGE_NAME to specify the docker image name.
func Docker() error {
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv(envBuildRebuildAll, "1")

	mg.SerialDeps(Clean, Build)

	fmt.Println("Building the docker image...")

	imageName := os.Getenv(envDockerImageName)
	if imageName == "" {
		timestamp := time.Now().UTC().Unix()
		imageName = fmt.Sprintf("localhost/%s:dev-%d", appName(), timestamp)
	}

	if err := sh.Run("docker", "build", "-t", imageName, "-f", dockerfile(), "."); err != nil {
		return fmt.Errorf("error building the docker image: %w", err)
	}

	fmt.Printf("Successfully built the docker image %q.\n", imageName)

	return nil
}

// ldflags returns the build flags.
func ldflags() string {
	var (
		infoPackage              = "codeflow.dananglin.me.uk/apollo/beacon/internal/info"
		binaryVersionVar         = infoPackage + "." + "BinaryVersion"
		gitCommitVar             = infoPackage + "." + "GitCommit"
		goVersionVar             = infoPackage + "." + "GoVersion"
		buildTimeVar             = infoPackage + "." + "BuildTime"
		applicationNameVar       = infoPackage + "." + "ApplicationName"
		applicationTitledNameVar = infoPackage + "." + "ApplicationTitledName"

		buildTime  = time.Now().UTC().Format(time.RFC3339)
		ldflagsfmt = "-s -w -X %s=%s -X %s=%s -X %s=%s -X %s=%s -X %s=%s -X %s=%s"
	)

	return fmt.Sprintf(
		ldflagsfmt,
		binaryVersionVar, version(),
		gitCommitVar, gitCommit(),
		goVersionVar, runtime.Version(),
		buildTimeVar, buildTime,
		applicationNameVar, appName(),
		applicationTitledNameVar, appTitledName(),
	)
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

// appName returns the application's name.
// The value of BEACON_APP_NAME is return if the environment variable is set
// otherwise the default name is returned.
func appName() string {
	appName := os.Getenv(envAppName)

	if appName == "" {
		return defaultAppName
	}

	return appName
}

func appTitledName() string {
	runes := []rune(appName())
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

func dockerfile() string {
	dockerfile := os.Getenv(envDockerfile)

	if dockerfile == "" {
		return defaultDockerfile
	}

	return dockerfile
}
