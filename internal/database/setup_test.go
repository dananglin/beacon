// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database_test

import (
	"testing"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
	bolt "go.etcd.io/bbolt"
)

func testDatabaseSetup(boltdb *bolt.DB) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("Ensuring that the database is not yet initialized.")

		initialised, err := database.Initialized(boltdb)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after checking whether or not the database is initialized.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		}

		if initialised {
			t.Fatalf(
				"FAILED test %s: The database appears to be initialized.",
				t.Name(),
			)
		} else {
			t.Logf("The database is not yet initialized as expected.")
		}

		t.Log("Now setting up the database.")

		profileID, err := utilities.ValidateAndCanonicalizeURL("https://pippins.example.me", false)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error validating the profile URL.\ngot %q",
				t.Name(),
				err.Error(),
			)
		}

		hashPassword, err := auth.HashPassword("fGEo1iGSsfqY")
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after attempting to hash the password.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		}

		timestamp := time.Now()

		profile := database.Profile{
			HashedPassword: hashPassword,
			CreatedAt:      timestamp,
			UpdatedAt:      timestamp,
			Information: database.ProfileInformation{
				Name:     "Pip Hubert",
				URL:      "https://pippins.example.me/about/me",
				PhotoURL: "https://pippins.example.me/about/me/profile.png",
			},
		}

		if err := database.Setup(boltdb, profileID, profile); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after setting up the database.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		} else {
			t.Log("Database was set up successfully.")
		}

		t.Log("Ensuring that the database is indeed initialised.")

		initialised, err = database.Initialized(boltdb)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after checking whether or not the database is initialised.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		}

		if !initialised {
			t.Errorf(
				"FAILED test %s: The database does not appear to be initialized.",
				t.Name(),
			)
		} else {
			t.Logf("The database appears to be initialized as expected.")
		}
	}
}
