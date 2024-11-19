// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database_test

import (
	"os"
	"path/filepath"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
)

func TestDatabase(t *testing.T) {
	dbPath := filepath.Join("testdata", t.Name()+".golden")

	t.Log("Opening the database.")

	boltdb, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Unable to open the database: %v",
			t.Name(),
			err,
		)
	} else {
		t.Log("Successfully opened the database.")
	}

	// Close and delete the database at the end of the test.
	defer func() {
		_ = boltdb.Close()
		_ = os.Remove(dbPath)
	}()

	t.Run("Test Database Setup", testDatabaseSetup(boltdb))
	t.Run("Test Profile Lifecycle", testProfile(boltdb, t.Name()+" (Profile)"))
}
