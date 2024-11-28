// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/discovery"
)

func TestGetMetadataFromHTML(t *testing.T) {
	t.Parallel()

	t.Run("Discovery via HTML", testGetMetadataFromHTML)
	t.Run("Discovery via JSON", testGetMetadataFromJSON)
	t.Run("Bad status code from client", testBadStatusCode)
}

func testGetMetadataFromHTML(t *testing.T) {
	t.Parallel()

	testHTMLWebsite := `
<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Test Webpage</title>
        <link rel="redirect_uri" href="/callback">
        <link rel="redirect_uri" href="http://callback.testclient.example">
    </head>

    <body>
      <header>
        <div class="h-app">
          <img src="assets/logo.png" class="u-logo">
          <a href="/" class="u-url p-name">Test Client Website</a>
        </div>
      </header>
      <h1>Test Webpage</h1>
      <p>A simple paragraph</p>
    </body>
</html>
`

	testClient := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(testHTMLWebsite))
	}))
	defer testClient.Close()

	gotMetadata, err := discovery.FetchClientMetadata(context.Background(), testClient.URL)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error after attempting to get the client's metadata.\ngot: %q",
			t.Name(),
			err.Error(),
		)
	}

	wantMetadata := discovery.ClientIDMetadata{
		ClientID:   testClient.URL,
		ClientName: "Test Client Website",
		ClientURI:  testClient.URL + "/",
		LogoURI:    testClient.URL + "/assets/logo.png",
		RedirectURIs: []string{
			testClient.URL + "/callback",
			"http://callback.testclient.example",
		},
	}

	if !reflect.DeepEqual(wantMetadata, gotMetadata) {
		t.Errorf(
			"FAILED test %s: unexpected metadata received.\nwant: %+v\ngot: %+v",
			t.Name(),
			wantMetadata,
			gotMetadata,
		)
	} else {
		t.Logf(
			"Expected metadata received.\ngot: %+v",
			gotMetadata,
		)
	}
}

func testGetMetadataFromJSON(t *testing.T) {
	t.Parallel()

	testClientMetadataString := `
{
	"client_id": "https://testclient.example/",
	"client_name": "Test Client Website",
	"client_uri": "https://testclient.example/",
	"logo_uri": "https://testclient.example/static/logo.png",
	"redirect_uris": [
		"https://testclient.example/redirect",
		"https://testclient.example/callback",
		"https://callback.testclient.example/"
	]
}
`

	testClient := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(testClientMetadataString))
	}))
	defer testClient.Close()

	gotMetadata, err := discovery.FetchClientMetadata(context.Background(), testClient.URL)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error after attempting to get the client's metadata.\ngot: %q",
			t.Name(),
			err.Error(),
		)
	}

	wantMetadata := discovery.ClientIDMetadata{
		ClientID:   "https://testclient.example/",
		ClientName: "Test Client Website",
		ClientURI:  "https://testclient.example/",
		LogoURI:    "https://testclient.example/static/logo.png",
		RedirectURIs: []string{
			"https://testclient.example/redirect",
			"https://testclient.example/callback",
			"https://callback.testclient.example/",
		},
	}

	if !reflect.DeepEqual(wantMetadata, gotMetadata) {
		t.Errorf(
			"FAILED test %s: unexpected metadata received.\nwant: %+v\ngot: %+v",
			t.Name(),
			wantMetadata,
			gotMetadata,
		)
	} else {
		t.Logf(
			"Expected metadata received.\ngot: %+v",
			gotMetadata,
		)
	}
}

func testBadStatusCode(t *testing.T) {
	t.Parallel()

	testClient := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer testClient.Close()

	_, err := discovery.FetchClientMetadata(context.Background(), testClient.URL)
	if err == nil {
		t.Fatalf(
			"FAILED test %s: Did not receive an error for bad status code.",
			t.Name(),
		)
	}

	wantErr := discovery.BadStatusResponseError{}

	if !errors.As(err, &wantErr) {
		t.Errorf(
			"FAILED test %s: Unexpected error received for bad status code.\nwant something like: %q\ngot: %q",
			t.Name(),
			wantErr.Error(),
			err.Error(),
		)
	} else {
		t.Logf(
			"Expected error received for bad status code.\ngot: %q",
			err.Error(),
		)
	}
}
