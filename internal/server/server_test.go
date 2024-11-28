// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"os"
	"testing"
)

func TestServer(t *testing.T) {
	t.Parallel()

	testServer, err := NewServer("testdata/Config.json.golden")
	if err != nil {
		t.Fatalf("FAILED test %s: Unable to create the test server: %v", t.Name(), err)
	}

	defer func() {
		os.Remove("testdata/Database.db.golden")
	}()

	t.Run("Test Server Metadata", testGetMetadata(testServer))
}
