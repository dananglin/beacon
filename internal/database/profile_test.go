// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database_test

import (
	"reflect"
	"testing"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
	bolt "go.etcd.io/bbolt"
)

func testProfile(boltdb *bolt.DB, testName string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Logf("Creating the profile in the database.")

		website := "https://billjones.example.net"

		profileID, err := utilities.ValidateAndCanonicalizeURL(website)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error validating the profile URL: %v",
				testName,
				err,
			)
		}

		hashedPassword, err := auth.HashPassword("test_p@$sW0rd")
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to create the hashed password: %v",
				testName,
				err,
			)
		}

		profile := database.Profile{
			HashedPassword: hashedPassword,
			CreatedAt:      time.Time{},
			UpdatedAt:      time.Time{},
			Information: database.ProfileInformation{
				Name:     "Bill Jones",
				URL:      "https://billjones.example.net/about/me",
				PhotoURL: "https://billjones.example.net/assets/images/profile.png",
				Email:    "hi@billjones.example.net",
			},
		}

		if err = database.CreateProfile(boltdb, profileID, profile); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after adding the profile to the database: %v",
				testName,
				err,
			)
		} else {
			t.Log("Successfully created the profile.")
		}

		t.Log("Checking that the profile exists in the database.")

		profileExists, err := database.ProfileExists(boltdb, profileID)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after checking if the profile exists or not: %v",
				testName,
				err,
			)
		}

		if !profileExists {
			t.Fatalf(
				"FAILED test %s: The profile for %q does not exist in the database",
				testName,
				profileID,
			)
		} else {
			t.Log("The profile is present in the database.")
		}

		gotProfile, err := database.GetProfile(boltdb, profileID)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after retrieving the profile from the database: %v",
				testName,
				err,
			)
		}

		if gotProfile.CreatedAt.IsZero() {
			t.Errorf(
				"FAILED test %s: The profile's 'CreatedAt' field is set to its zero value",
				testName,
			)
		}

		if gotProfile.UpdatedAt.IsZero() {
			t.Errorf(
				"FAILED test %s: The profile's 'UpdatedAt' field is set to its zero value",
				testName,
			)
		}

		t.Log("Retrieving the profile's information from the database.")

		gotProfileInfo, err := database.GetProfileInformation(boltdb, profileID)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to get the profile information from the database: %v",
				testName,
				err,
			)
		} else {
			t.Log("Successfully received the profile's information from the database.")
		}

		if !reflect.DeepEqual(gotProfileInfo, profile.Information) {
			t.Errorf(
				"FAILED test %s: Unexpected profile information received from the database, want:\n%+v\ngot:\n%+v",
				testName,
				profile.Information,
				gotProfileInfo,
			)
		} else {
			t.Logf(
				"Expected profile information received from the database, got:\n%+v",
				gotProfileInfo,
			)
		}

		gotTokenVersion, err := database.GetProfileTokenVersion(boltdb, profileID)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to get the profile's token version from the database: %v",
				testName,
				err,
			)
		} else {
			t.Logf("Successfully received the profile's token version from the database.")
		}

		if gotTokenVersion != 0 {
			t.Errorf(
				"FAILED test %s: Unexpected token version received from the database: want 0, got %d",
				testName,
				gotTokenVersion,
			)
		} else {
			t.Logf(
				"Expected token version received from the database: got %d",
				gotTokenVersion,
			)
		}

		t.Log("Updating the profile's information")

		newProfileInformation := database.ProfileInformation{
			Name:     "Bill Jones",
			URL:      "https://billjones.example.net/about/bill",
			PhotoURL: "https://billjones.example.net/about/bill/profile.png",
			Email:    "hi@billjones.example.net",
		}

		if err := database.UpdateProfileInformation(boltdb, profileID, newProfileInformation); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after updating the profile information in the database: %v",
				testName,
				err,
			)
		}

		gotProfile, err = database.GetProfile(boltdb, profileID)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error after retrieving the profile from the database: %v",
				testName,
				err,
			)
		}

		if !reflect.DeepEqual(gotProfile.Information, newProfileInformation) {
			t.Errorf(
				"FAILED test %s: Unexpected profile information received from the database, want:\n%+v\ngot:\n%+v",
				testName,
				newProfileInformation,
				gotProfile.Information,
			)
		} else {
			t.Logf(
				"Expected profile information received from the database, got:\n%+v",
				gotProfile.Information,
			)
		}

		if !gotProfile.UpdatedAt.After(gotProfile.CreatedAt) {
			t.Errorf(
				"FAILED test %s: The profile's 'UpdatedAt' field has not been updated",
				testName,
			)
		}
	}
}
