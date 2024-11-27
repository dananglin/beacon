// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/server"
)

func testGetMetadata(srv *server.Server) func(t *testing.T) {
	return func(t *testing.T) {
		writer := httptest.NewRecorder()

		srv.GetMetadata(writer, nil)

		response := writer.Result()
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			t.Fatalf(
				"FAILED test %s: Unexpected status code received.\nwant: %d, got: %d",
				t.Name(),
				http.StatusOK,
				response.StatusCode,
			)
		}

		if response.Header.Get("Content-Type") != "application/json" {
			t.Fatalf(
				"FAILED test %s: Unexpected content type received.\nwant: %q\ngot: %s",
				t.Name(),
				"application/json",
				response.Header.Get("Content-Type"),
			)
		}

		var got server.Metadata

		if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error decoding the JSON data.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		}

		want := server.Metadata{
			Issuer:                                 "https://indieauth.test.example/",
			AuthorizationEndpoint:                  "https://indieauth.test.example/indieauth/authorize",
			TokenEndpoint:                          "https://indieauth.test.example/indieauth/token",
			ServiceDocumentation:                   "https://indieauth.spec.indieweb.org",
			CodeChallengeMethodsSupported:          []string{"S256"},
			GrantTypesSupported:                    []string{"authorization_code"},
			ResponseTypesSupported:                 []string{"code"},
			ScopesSupported:                        []string{"profile", "email"},
			AuthorizationResponseISSParamSupported: true,
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf(
				"FAILED test %s: Unexpected results returned.\nwant: %+v\ngot: %+v",
				t.Name(),
				want,
				got,
			)
		} else {
			t.Logf(
				"Expected results returned.\ngot: %+v",
				got,
			)
		}
	}
}
