package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

func New(path string) (*bolt.DB, error) {
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

	if err := ensureBucket(boltdb); err != nil {
		return nil, fmt.Errorf(
			"unable to ensure that the required buckets are present in the database: %w",
			err,
		)
	}

	// Add the initialised key to the bucket if it does
	// not exist already.
	initialisedKey, err := initialisedKeyExists(boltdb)
	if err != nil {
		return nil, err
	}

	if !initialisedKey {
		if err := addInitialisedKey(boltdb); err != nil {
			return nil, err
		}
	}

	return boltdb, nil
}

func ensureBucket(boltdb *bolt.DB) error {
	err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := getBucketName()
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return fmt.Errorf(
				"unable to ensure the existence of the %q bucket: %w",
				string(bucket),
				err,
			)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf(
			"error ensuring the existence of the buckets in the database: %w",
			err,
		)
	}

	return nil
}

const bucketName string = "application"

func getBucketName() []byte {
	return []byte(bucketName)
}
