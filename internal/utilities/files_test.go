// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"errors"
	"os"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func TestMakeDir(t *testing.T) {
	path := "testdata/" + t.Name()

	defer func() {
		if err := os.Remove(path); err != nil {
			t.Logf("WARNING: error removing %s: %v", path, err)
		}
	}()

	err := utilities.MakeDir(path)
	if err != nil {
		t.Fatalf("FAILED test %s: Received an error making the directory at %s: %v", t.Name(), path, err)
	}

	t.Logf("Successfully created the directory at %s", path)

	err = utilities.CheckDirPerm(path)
	if err != nil {
		t.Fatalf("FAILED test %s: Received an error checking if the directory is set to the right permission: %v", t.Name(), err)
	}

	t.Log("The directory is set to the right permission.")
}

func TestCheckDirPerm(t *testing.T) {
	goodDirPath := "testdata/" + t.Name() + "_good"
	badDirPath := "testdata/" + t.Name() + "_bad"

	defer func() {
		for _, path := range []string{goodDirPath, badDirPath} {
			if err := os.Remove(path); err != nil {
				t.Logf("WARNING: error removing %s: %v", path, err)
			}
		}
	}()

	err := os.Mkdir(goodDirPath, 0o700)
	if err != nil {
		t.Fatalf("FAILED test %s: Error creating %s: %v", t.Name(), goodDirPath, err)
	}

	t.Logf("Successfully created the directory at %s", goodDirPath)

	err = os.Mkdir(badDirPath, 0o777)
	if err != nil {
		t.Fatalf("FAILED test %s: Error creating %s: %v", t.Name(), badDirPath, err)
	}

	t.Logf("Successfully created the directory at %s", badDirPath)

	err = utilities.CheckDirPerm(goodDirPath)
	if err != nil {
		t.Errorf("FAILED test %s: Received an error checking if the directory at %s is set to the right permission: %v", t.Name(), goodDirPath, err)
	} else {
		t.Logf("The CheckDirPerm() function confirms that the directory created at %s is set to the right permission", goodDirPath)
	}

	err = utilities.CheckDirPerm(badDirPath)
	if err == nil {
		t.Errorf("FAILED test %s: Did not receive an error after checking if the directory at %s was set to the right permission", t.Name(), badDirPath)
	} else {
		wantErr := utilities.NewIncorrectDirPermError(badDirPath)
		if !errors.Is(err, wantErr) {
			t.Errorf(
				"FAILED test %s: Received an unexpected error after checking if the directory at %s was set to the right permission.\nwant: %v\n got: %v",
				t.Name(),
				badDirPath,
				wantErr,
				err,
			)
		} else {
			t.Logf(
				"Received expected error after checking if the directory at %s was set to the right permission.\ngot: %v",
				badDirPath,
				err,
			)
		}
	}
}

func TestCheckFilePerm(t *testing.T) {
	goodFilePath := "testdata/" + t.Name() + ".good"
	badFilePath := "testdata/" + t.Name() + ".bad"

	defer func() {
		for _, path := range []string{goodFilePath, badFilePath} {
			if err := os.Remove(path); err != nil {
				t.Logf("WARNING: error removing %s: %v", path, err)
			}
		}
	}()

	goodFile, err := os.OpenFile(goodFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatalf("FAILED test %s: Error creating the file at %s: %v", t.Name(), goodFilePath, err)
	}
	_ = goodFile.Close()

	badFile, err := os.OpenFile(badFilePath, os.O_RDWR|os.O_CREATE, 0o660)
	if err != nil {
		t.Fatalf("FAILED test %s: Error creating the file at %s: %v", t.Name(), badFilePath, err)
	}
	_ = badFile.Close()

	err = utilities.CheckFilePerm(goodFilePath)
	if err != nil {
		t.Errorf("FAILED test %s: Received an error after checking if the file at %s was created with the right permission: %v", t.Name(), goodFilePath, err)
	} else {
		t.Logf("The CheckFilePerm() function confirms that the file at %s is set to the right permission", goodFilePath)
	}

	err = utilities.CheckFilePerm(badFilePath)
	if err == nil {
		t.Errorf("FAILED test %s: Did not receive an error after checking if the directory at %s was set to the right permission", t.Name(), badFilePath)
	} else {
		wantErr := utilities.NewIncorrectFilePermError(badFilePath)
		if !errors.Is(err, wantErr) {
			t.Errorf(
				"FAILED test %s: Received an unexpected error after checking if the file at %s was set to the right permission.\nwant: %v\n got: %v",
				t.Name(),
				badFilePath,
				wantErr,
				err,
			)
		} else {
			t.Logf(
				"Received expected error after checking if the file at %s was set to the right permission.\ngot: %v",
				badFilePath,
				err,
			)
		}
	}
}

func TestFileExists(t *testing.T) {
	path := "testdata/" + t.Name() + ".good"

	defer func() {
		if err := os.Remove(path); err != nil {
			t.Logf("WARNING: error removing %s: %v", path, err)
		}
	}()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("FAILED test %s: Error creating the file at %s: %v", t.Name(), path, err)
	}
	_ = file.Close()

	t.Log("Successfully created the file at " + path)

	exists, err := utilities.FileExists(path)
	if err != nil {
		t.Fatalf("FAILED test %s: Error checking if the file exists: %v", t.Name(), err)
	}

	if !exists {
		t.Errorf("FAILED test %s: The function FileExists() claimed that the file created at %s is not present", t.Name(), path)
	} else {
		t.Logf("The function FileExists() confirms that the file created at %s is present", path)
	}

	badPath := "testdata/" + t.Name() + ".bad"

	exists, err = utilities.FileExists(badPath)
	if err != nil {
		t.Fatalf("FAILED test %s: Error checking if the file exists: %v", t.Name(), err)
	}

	if exists {
		t.Errorf("FAILED test %s: The function FileExists() claimed that the non-existing file at %s is present", t.Name(), badPath)
	} else {
		t.Logf("The function FileExists() confirms that the non-existing file at %s is not present", badPath)
	}
}
