// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"os"
	"testing"
)

func TestServer(t *testing.T) {
	t.Parallel()

	testdataDir := "testdata/data"

	testServer, err := NewServer("testdata/config.json")
	if err != nil {
		t.Fatalf("FAILED test %s: Unable to create the test server: %v", t.Name(), err)
	}

	defer func() {
		if err := os.RemoveAll(testdataDir); err != nil {
			t.Logf("WARNING: Error removing %s: %v", testdataDir, err)
		}
	}()

	t.Run("Test Server Metadata", testGetMetadata(testServer))
}
