// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Open opens the BoltDB database at the given path.
func Open(path string) (*bolt.DB, error) {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("unable to create directory %q: %w", dir, err)
	}

	opts := bolt.Options{
		Timeout: 1 * time.Second,
	}

	boltdb, err := bolt.Open(path, 0o600, &opts)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to open the database at %q: %w",
			path,
			err,
		)
	}

	return boltdb, nil
}

// Setup sets up the database by creating the 'profiles' bucket and
// writing the first profile to that bucket.
func Setup(boltdb *bolt.DB, profileID string, profile Profile) error {
	if err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := getBucketName()
		if _, err := tx.CreateBucket(bucket); err != nil {
			return fmt.Errorf(
				"unable to create the bucket %q: %w",
				profilesBucketName,
				err,
			)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error creating the BoltDB bucket: %w", err)
	}

	if err := saveProfile(boltdb, profileID, profile); err != nil {
		return fmt.Errorf("error saving the profile: %w", err)
	}

	return nil
}

// Initialized checks to see if the database is initialized or not.
// The database is initialized if the 'profiles' bucket exists and that
// there is at least one profile stored in the bucket.
func Initialized(boltdb *bolt.DB) (bool, error) {
	initialized := false

	if err := boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(getBucketName())
		if bucket == nil {
			return nil
		}

		if bucket.Stats().BucketN > 0 {
			initialized = true
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("error checking if the database is initialized or not: %w", err)
	}

	return initialized, nil
}
