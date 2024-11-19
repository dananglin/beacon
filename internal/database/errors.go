// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package database

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
