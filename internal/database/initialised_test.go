// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database_test

import (
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	bolt "go.etcd.io/bbolt"
)

func testInitialised(boltdb *bolt.DB, testName string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("Checking that the 'initialised' key is set to 'false'.")

		initialised, err := database.Initialised(boltdb)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after checking whether or not the database is initialised: %v",
				testName,
				err,
			)
		}

		if initialised {
			t.Fatalf(
				"FAILED test %s: The 'initialised' key is set to true",
				testName,
			)
		} else {
			t.Logf("The 'initialised' key is set to false as expected.")
		}

		t.Log("Setting the 'initialised' key to true.")

		if err := database.UpdateInitialisedKey(boltdb); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error to update the initialised key: %v",
				testName,
				err,
			)
		} else {
			t.Log("Successfully updated the 'initialised' key.")
		}

		t.Log("Ensuring that the 'initialised' key is set to true.")

		initialised, err = database.Initialised(boltdb)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after checking whether or not the database is initialised, got: %q",
				testName,
				err.Error(),
			)
		}

		if !initialised {
			t.Fatalf(
				"FAILED test %s: The 'initialised' key is set to false",
				testName,
			)
		} else {
			t.Logf("The 'initialised' key is set to true as expected.")
		}
	}
}
