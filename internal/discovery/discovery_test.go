// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery_test

import (
	"net/url"
	"os"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/discovery"
)

func TestGetMetadataFromHTML(t *testing.T) {
	filename := "testdata/TestWebsite.golden"

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: unable to open %s: %v",
			t.Name(),
			filename,
			err,
		)
	}
	defer file.Close()

	clientID := "https://test.website.net/"

	parsedClientID, err := url.Parse(clientID)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: unable to parse %s: %v",
			t.Name(),
			clientID,
			err,
		)
	}

	gotMetadata := discovery.GetMetadataFromHTML(file, clientID, parsedClientID)

	wantMetadata := discovery.ClientIDMetadata{
		ClientID:   "https://test.website.net/",
		ClientName: "Test Website",
		ClientURI:  "https://test.website.net/",
		LogoURI:    "https://test.website.net/assets/logo.png",
		RedirectURIs: []string{
			"https://test.website.net/redirect",
			"https://redirect.website.net/",
		},
	}

	if !reflect.DeepEqual(wantMetadata, gotMetadata) {
		t.Errorf(
			"FAILED test %s: unexpected metadata received:\nwant: %+v\ngot:  %+v",
			t.Name(),
			wantMetadata,
			gotMetadata,
		)
	} else {
		t.Logf(
			"Expected metadata received:\ngot:  %+v",
			gotMetadata,
		)
	}
}
