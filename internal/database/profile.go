package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BucketNotExistError struct {
	bucket string
}

func (e BucketNotExistError) Error() string {
	return "the '" + e.bucket + "' bucket does not exist"
}

type ProfileNotExistError struct {
	profileID string
}

func (e ProfileNotExistError) Error() string {
	return "the profile for '" + e.profileID + "' does not exist"
}

type ProfileAlreadyExistError struct {
	profileID string
}

func (e ProfileAlreadyExistError) Error() string {
	return "the profile for '" + e.profileID + "' is already present in the database"
}

type Profile struct {
	CreatedAt      time.Time
	UpdatedAt      time.Time
	TokenVersion   int
	HashedPassword string
	Information    ProfileInformation
}

type ProfileInformation struct {
	Name     string
	URL      string
	PhotoURL string
	Email    string
}

// CreateProfile creates a new profile in the database.
func CreateProfile(boltdb *bolt.DB, profileID string, profile Profile) error {
	profileExists, err := ProfileExists(boltdb, profileID)
	if err != nil {
		return fmt.Errorf("unable to check if the profile already exists in the database: %w", err)
	}

	if profileExists {
		return ProfileAlreadyExistError{profileID: profileID}
	}

	timestamp := time.Now()
	profile.CreatedAt = timestamp
	profile.UpdatedAt = timestamp

	profile.TokenVersion = 0

	return saveProfile(boltdb, profileID, profile)
}

// UpdateProfileInformation updates an existing profile's information.
func UpdateProfileInformation(boltdb *bolt.DB, profileID string, newProfileInfo ProfileInformation) error {
	profile, err := getProfile(boltdb, profileID)
	if err != nil {
		return fmt.Errorf("unable to get the profile from the database: %w", err)
	}

	timestamp := time.Now()
	profile.Information = newProfileInfo
	profile.UpdatedAt = timestamp

	if err := saveProfile(boltdb, profileID, profile); err != nil {
		return fmt.Errorf("unable to save the updated profile to the database: %w", err)
	}

	return nil
}

func saveProfile(boltdb *bolt.DB, profileID string, profile Profile) error {
	bucketName := getBucketName()

	err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return BucketNotExistError{bucket: string(bucketName)}
		}

		key := []byte(profileID)

		buffer := new(bytes.Buffer)
		if err := gob.NewEncoder(buffer).Encode(profile); err != nil {
			return fmt.Errorf(
				"unable to encode the user data: %w",
				err,
			)
		}

		if err := bucket.Put(key, buffer.Bytes()); err != nil {
			return fmt.Errorf(
				"unable to update the user in the %s bucket: %w",
				string(bucketName),
				err,
			)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error updating the user in the database: %w", err)
	}

	return nil
}

// ProfileExists checks if a profile exists for a given website.
func ProfileExists(boltdb *bolt.DB, profileID string) (bool, error) {
	profileExists := false
	bucketName := getBucketName()
	key := []byte(profileID)

	if err := boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return BucketNotExistError{bucket: string(bucketName)}
		}

		profile := bucket.Get(key)
		if profile != nil {
			profileExists = true
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("unable to check of the profile exists in the bucket: %w", err)
	}

	return profileExists, nil
}

// GetProfile returns the profile for a given profile ID.
func GetProfile(boltdb *bolt.DB, profileID string) (Profile, error) {
	return getProfile(boltdb, profileID)
}

// GetProfileInformation returns the profile information for a given profile ID.
func GetProfileInformation(boltdb *bolt.DB, profileID string) (ProfileInformation, error) {
	profile, err := getProfile(boltdb, profileID)
	if err != nil {
		return ProfileInformation{}, fmt.Errorf("error getting profile: %w", err)
	}

	return profile.Information, nil
}

// GetProfileTokenVersion returns the token version for a given profile ID.
func GetProfileTokenVersion(boltdb *bolt.DB, profileID string) (int, error) {
	profile, err := getProfile(boltdb, profileID)
	if err != nil {
		return 0, fmt.Errorf("error getting profile: %w", err)
	}

	return profile.TokenVersion, nil
}

func getProfile(boltdb *bolt.DB, profileID string) (Profile, error) {
	bucketName := getBucketName()
	key := []byte(profileID)

	var profile Profile

	if err := boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return BucketNotExistError{bucket: string(bucketName)}
		}

		data := bucket.Get(key)
		if data == nil {
			return ProfileNotExistError{profileID: profileID}
		}

		buffer := bytes.NewBuffer(data)

		if err := gob.NewDecoder(buffer).Decode(&profile); err != nil {
			return fmt.Errorf("unable to decode the profile: %w", err)
		}

		return nil
	}); err != nil {
		return Profile{}, fmt.Errorf("unable to get the profile from the database: %w", err)
	}

	return profile, nil
}
