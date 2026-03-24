// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database

import (
	"fmt"
	"path/filepath"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
	bolt "go.etcd.io/bbolt"
)

// Open opens the BoltDB database at the given path.
func Open(path string) (*bolt.DB, error) {
	dir := filepath.Dir(path)

	dirExists, err := utilities.FileExists(dir)
	if err != nil {
		return nil, fmt.Errorf("error checking if the database's directory exists: %w", err)
	}

	if dirExists {
		err = utilities.CheckDirPerm(dir)
		if err != nil {
			return nil, fmt.Errorf("error checking the directory permission of %s: %w", dir, err)
		}
	} else {
		err := utilities.MakeDir(dir)
		if err != nil {
			return nil, fmt.Errorf("error creating %s: %w", dir, err)
		}
	}

	fileExists, err := utilities.FileExists(path)
	if err != nil {
		return nil, fmt.Errorf("error checking if the database file exists: %w", err)
	}

	if fileExists {
		err = utilities.CheckFilePerm(path)
		if err != nil {
			return nil, fmt.Errorf("error checking the file permission of %s: %w", path, err)
		}
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
