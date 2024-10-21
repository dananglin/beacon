package database_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/auth"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/database"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/utilities"
)

func TestDatabase(t *testing.T) {
	dbPath := filepath.Join("testdata", t.Name()+".golden")

	t.Log("Opening and initialising the database.")

	boltdb, err := database.New(dbPath)
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

	t.Log("Checking that the 'initialised' key is set to 'false'.")

	initialised, err := database.Initialised(boltdb)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error after checking whether or not the database is initialised: %v",
			t.Name(),
			err,
		)
	}

	if initialised {
		t.Fatalf(
			"FAILED test %s: The 'initialised' key is set to true",
			t.Name(),
		)
	} else {
		t.Logf("The 'initialised' key is set to false as expected.")
	}

	t.Logf("Creating the profile in the database.")

	inputWebsiteURL := "https://billjones.example.net"

	canonicalisedWebsiteURL, err := utilities.ValidateProfileURL(inputWebsiteURL)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error validating the profile URL: %v",
			t.Name(),
			err,
		)
	}

	hashedPassword, err := auth.HashPassword("test_p@$sW0rd")
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Unable to create the hashed password: %v",
			t.Name(),
			err,
		)
	}

	timestamp := time.Now()

	profile := database.Profile{
		HashedPassword: hashedPassword,
		CreatedAt:      timestamp,
		UpdatedAt:      timestamp,
		Information: database.ProfileInformation{
			Name:     "Bill Jones",
			URL:      "https://billjones.example.net/about/me",
			PhotoURL: "https://billjones.example.net/assets/images/profile.png",
			Email:    "hi@billjones.example.net",
		},
	}

	if err = database.UpdateProfile(boltdb, canonicalisedWebsiteURL, profile); err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error after adding the profile to the database: %v",
			t.Name(),
			err,
		)
	} else {
		t.Log("Successfully created the profile.")
	}

	t.Log("Checking that the profile exists in the database.")

	profileExists, err := database.ProfileExists(boltdb, canonicalisedWebsiteURL)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error after checking if the profile exists or not: %v",
			t.Name(),
			err,
		)
	}

	if !profileExists {
		t.Fatalf(
			"FAILED test %s: The profile for %q does not exist in the database",
			t.Name(),
			canonicalisedWebsiteURL,
		)
	} else {
		t.Log("The profile is present in the database.")
	}

	// Retrieve the profile's information from the database
	t.Log("Retrieving the profile's information from the database.")

	gotProfileInfo, err := database.GetProfileInformation(boltdb, canonicalisedWebsiteURL)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Unable to get the profile information from the database: %v",
			t.Name(),
			err,
		)
	} else {
		t.Log("Successfully received the profile's information from the database.")
	}

	if !reflect.DeepEqual(gotProfileInfo, profile.Information) {
		t.Errorf(
			"FAILED test %s: Unexpected profile information received from the database, want:\n%+v\ngot:\n%+v",
			t.Name(),
			profile.Information,
			gotProfileInfo,
		)
	} else {
		t.Logf(
			"PASSED test %s: Expected profile information received from the database, got:\n%+v",
			t.Name(),
			gotProfileInfo,
		)
	}

	t.Log("Setting the 'initialised' key to true.")

	if err := database.UpdateInitialisedKey(boltdb); err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error to update the initialised key: %v",
			t.Name(),
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
			t.Name(),
			err.Error(),
		)
	}

	if !initialised {
		t.Fatalf(
			"FAILED test %s: The 'initialised' key is set to false",
			t.Name(),
		)
	} else {
		t.Logf("The 'initialised' key is set to true as expected.")
	}
}
