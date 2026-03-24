// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// Setup sets up the database by creating the 'profiles' bucket and
// writing the first profile to that bucket.
func Setup(boltdb *bolt.DB, profileID string, profile Profile) error {
	if err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := getBucketName()
		if _, err := tx.CreateBucket(bucket); err != nil {
			return fmt.Errorf(
				"error creating the bucket %q: %w",
				profilesBucketName,
				err,
			)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error creating a BoltDB bucket: %w", err)
	}

	if err := CreateProfile(boltdb, profileID, profile); err != nil {
		return fmt.Errorf("error creating the profile: %w", err)
	}

	return nil
}
