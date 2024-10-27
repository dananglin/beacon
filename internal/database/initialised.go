// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const initialisedKeyName string = "initialised"

// UpdateInitialisedKey sets the initialised key to true.
func UpdateInitialisedKey(boltdb *bolt.DB) error {
	keyExists, err := initialisedKeyExists(boltdb)
	if err != nil {
		return err
	}

	if !keyExists {
		return errors.New("the initialised key is not present in the bucket")
	}

	bucketName := getBucketName()
	initialisedKey := getInitialisedKey()

	if err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return fmt.Errorf("the %s bucket does not exist", string(bucketName))
		}

		initialised := true
		buffer := new(bytes.Buffer)

		if err := gob.NewEncoder(buffer).Encode(initialised); err != nil {
			return fmt.Errorf(
				"unable to encode the initialised value: %w",
				err,
			)
		}

		if err := bucket.Put(initialisedKey, buffer.Bytes()); err != nil {
			return fmt.Errorf(
				"unable to add the initialised key and value to the bucket: %w",
				err,
			)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("unable to add the initialised key to the bucket: %w", err)
	}

	return nil
}

// Initialised returns true if the database has been initialised by the user.
func Initialised(boltdb *bolt.DB) (bool, error) {
	initialised := false
	bucketName := getBucketName()
	initialisedKey := getInitialisedKey()

	if err := boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return fmt.Errorf("the %s bucket does not exist", string(bucketName))
		}

		data := bucket.Get(initialisedKey)
		if data == nil {
			return errors.New("the initialised key is not present in the bucket")
		}

		buffer := bytes.NewBuffer(data)

		if err := gob.NewDecoder(buffer).Decode(&initialised); err != nil {
			return fmt.Errorf("unable to decode the value of the initialised key: %w", err)
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("unable get the value of the initialised key: %w", err)
	}

	return initialised, nil
}

// initialisedKeyExists returns true if the initialised key is present in the bucket.
func initialisedKeyExists(boltdb *bolt.DB) (bool, error) {
	keyExists := false
	bucketName := getBucketName()
	initialisedKey := getInitialisedKey()

	if err := boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return fmt.Errorf("the %s bucket does not exist", string(bucketName))
		}

		initialised := bucket.Get(initialisedKey)
		if initialised != nil {
			keyExists = true
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("unable to check if the initialised key exists in the bucket: %w", err)
	}

	return keyExists, nil
}

// addInitialisedKey adds the initialised key to the bucket and sets it to false.
func addInitialisedKey(boltdb *bolt.DB) error {
	keyExists, err := initialisedKeyExists(boltdb)
	if err != nil {
		return err
	}

	if keyExists {
		return errors.New("the initialised key is already present in the bucket")
	}

	bucketName := getBucketName()
	initialisedKey := getInitialisedKey()

	if err := boltdb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			return fmt.Errorf("the %s bucket does not exist", string(bucketName))
		}

		notInitialised := false
		buffer := new(bytes.Buffer)

		if err := gob.NewEncoder(buffer).Encode(notInitialised); err != nil {
			return fmt.Errorf(
				"unable to encode the notInitialised value: %w",
				err,
			)
		}

		if err := bucket.Put(initialisedKey, buffer.Bytes()); err != nil {
			return fmt.Errorf(
				"unable to add the initialised key and value to the bucket: %w",
				err,
			)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("unable to add the initialised key to the bucket: %w", err)
	}

	return nil
}

func getInitialisedKey() []byte {
	return []byte(initialisedKeyName)
}
