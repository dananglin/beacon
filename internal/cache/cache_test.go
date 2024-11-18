// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package cache_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/cache"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type TestData struct {
	Name    string
	Number  int
	BoolVal bool
}

func TestCache(t *testing.T) {
	cleanupInterval := 1 * time.Second

	testCache := cache.NewCache(cleanupInterval)

	t.Run("Test Add Entry", testAddEntry(testCache))
	t.Run("Test Delete Entry", testDeleteEntry(testCache))
	t.Run("Test Expired Entry", testExpiredEntry(testCache))
	t.Run("Test Cache Cleanup", testCleanup(testCache))
}

func testAddEntry(testCache *cache.Cache) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := TestData{
			Name:    "add",
			Number:  1037,
			BoolVal: true,
		}

		dataBytes, err := utilities.GobEncode(data)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to encode the test data: %v",
				t.Name(),
				err,
			)
		}

		key := t.Name()
		expiresAt := time.Now().Add(1 * time.Minute)

		testCache.Add(key, dataBytes, expiresAt)

		t.Log("Added the test data to the cache.")

		entry, exists := testCache.Get(key)
		if !exists {
			t.Fatalf(
				"FAILED test %s: The key %q was not found after adding it to the cache",
				t.Name(),
				key,
			)
		} else {
			t.Log("Retrieved the data from the cache.")
		}

		var got TestData

		if err := utilities.GobDecode(bytes.NewBuffer(entry.Value()), &got); err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to decode the data from the cache: %v",
				t.Name(),
				err,
			)
		}

		if entry.Expired() {
			t.Errorf(
				"FAILED test %s: The cache entry appears to have expired",
				t.Name(),
			)
		} else {
			t.Log("The cached data has not expired as expected.")
		}

		if !reflect.DeepEqual(data, got) {
			t.Errorf(
				"FAILED test %s: Unexpected data retrieved from the cache\nwant: %+v\ngot: %+v",
				t.Name(),
				data,
				got,
			)
		} else {
			t.Logf(
				"Expected data retrieved from the cache\ngot: %+v",
				got,
			)
		}
	}
}

func testDeleteEntry(testCache *cache.Cache) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := TestData{
			Name:    "delete",
			Number:  1256,
			BoolVal: false,
		}

		dataBytes, err := utilities.GobEncode(data)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to encode the test data: %v",
				t.Name(),
				err,
			)
		}

		key := t.Name()
		expiresAt := time.Now().Add(1 * time.Minute)

		testCache.Add(key, dataBytes, expiresAt)

		t.Log("Added the test data to the cache.")

		testCache.Delete(key)

		_, exists := testCache.Get(key)
		if exists {
			t.Errorf(
				"FAILED test %s: The key %q was found after deleting it from the cache",
				t.Name(),
				key,
			)
		} else {
			t.Logf(
				"The key %q was not found after deleting it from the cache.",
				key,
			)
		}
	}
}

func testExpiredEntry(testCache *cache.Cache) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := TestData{
			Name:    "expired",
			Number:  45101,
			BoolVal: true,
		}

		dataBytes, err := utilities.GobEncode(data)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to encode the test data: %v",
				t.Name(),
				err,
			)
		}

		key := t.Name()
		expiresAt := time.Now().Add(1 * time.Millisecond)

		testCache.Add(key, dataBytes, expiresAt)

		t.Log("Added the test data to the cache.")

		time.Sleep(5 * time.Millisecond)

		entry, exists := testCache.Get(key)
		if !exists {
			t.Fatalf(
				"FAILED test %s: The key %q was not found after adding it to the cache",
				t.Name(),
				key,
			)
		} else {
			t.Log("Retrieved the data from the cache.")
		}

		if !entry.Expired() {
			t.Errorf(
				"FAILED test %s: The expired cache entry does not appear to have expired",
				t.Name(),
			)
		} else {
			t.Logf("The expired cache entry has expired as expected.")
		}
	}
}

func testCleanup(testCache *cache.Cache) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		data := TestData{
			Name:    "cleanup",
			Number:  67433,
			BoolVal: false,
		}

		dataBytes, err := utilities.GobEncode(data)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unable to encode the test data: %v",
				t.Name(),
				err,
			)
		}

		key := t.Name()
		expiresAt := time.Now().Add(1 * time.Millisecond)

		testCache.Add(key, dataBytes, expiresAt)

		t.Log("Added the test data to the cache.")
		t.Log("Waiting for the next cache clean-up")
		time.Sleep(2 * time.Second)

		_, exists := testCache.Get(key)
		if exists {
			t.Errorf(
				"FAILED test %s: The expired entry for %q was found after the cache clean-up",
				t.Name(),
				key,
			)
		} else {
			t.Logf(
				"The expired entry for %q was not found after the cache clean-up.",
				key,
			)
		}
	}
}
