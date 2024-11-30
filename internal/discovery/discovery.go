// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"willnorris.com/go/microformats"
)

type UnsupportedContentTypeError struct {
	contentType string
}

func (e UnsupportedContentTypeError) Error() string {
	return "unsupported content type '" + e.contentType + "'"
}

type BadStatusResponseError struct {
	code   int
	status string
}

func (e BadStatusResponseError) Error() string {
	return fmt.Sprintf("received a bad status from the client: (%d) %s", e.code, e.status)
}

type ClientIDMetadata struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	ClientURI    string   `json:"client_uri"`
	LogoURI      string   `json:"logo_uri"`
	RedirectURIs []string `json:"redirect_uris"`
}

func FetchClientMetadata(ctx context.Context, clientID, issuer string) (ClientIDMetadata, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, clientID, nil)
	if err != nil {
		return ClientIDMetadata{}, fmt.Errorf("error received after creating the HTTP request: %w", err)
	}

	request.Header.Set(
		"User-Agent",
		fmt.Sprintf("%s/%s (+%s)", info.ApplicationTitledName, info.BinaryVersion, issuer),
	)

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return ClientIDMetadata{}, fmt.Errorf("error getting the response from the client: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return ClientIDMetadata{}, BadStatusResponseError{
			code:   response.StatusCode,
			status: response.Status,
		}
	}

	metadata := ClientIDMetadata{}
	gotContentType := response.Header.Get("Content-Type")

	switch gotContentType {
	case "application/json":
		if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
			return ClientIDMetadata{}, fmt.Errorf("unable to decode the JSON data: %w", err)
		}
	case "text/html", "text/html; charset=UTF-8", "text/html; charset=utf-8":
		parsedClientID, err := url.Parse(clientID)
		if err != nil {
			return ClientIDMetadata{}, fmt.Errorf("unable to parse the client ID: %w", err)
		}

		metadata = GetMetadataFromHTML(response.Body, clientID, parsedClientID)
	default:
		return ClientIDMetadata{}, UnsupportedContentTypeError{contentType: gotContentType}
	}

	return metadata, nil
}

func GetMetadataFromHTML(reader io.Reader, clientID string, parsedClientID *url.URL) ClientIDMetadata {
	data := microformats.Parse(reader, parsedClientID)

	metadata := ClientIDMetadata{
		ClientID:     clientID,
		ClientName:   "",
		ClientURI:    "",
		LogoURI:      "",
		RedirectURIs: make([]string, 0),
	}

	if len(data.Items) == 0 {
		return metadata
	}

	for _, item := range data.Items {
		for _, formatType := range item.Type {
			if formatType == "h-app" || formatType == "h-x-app" {
				if values, ok := item.Properties["name"]; ok {
					if len(values) > 0 {
						if name, ok := values[0].(string); ok {
							metadata.ClientName = name
						}
					}
				}

				if values, ok := item.Properties["logo"]; ok {
					if len(values) > 0 {
						switch logo := values[0].(type) {
						case string:
							metadata.LogoURI = logo
						case map[string]string:
							if logoURL, ok := logo["value"]; ok {
								metadata.LogoURI = logoURL
							}
						}
					}
				}

				if values, ok := item.Properties["url"]; ok {
					if len(values) > 0 {
						if value, ok := values[0].(string); ok {
							metadata.ClientURI = value
						}
					}
				}

				break
			}
		}
	}

	if redirectURIs, ok := data.Rels["redirect_uri"]; ok {
		metadata.RedirectURIs = redirectURIs
	}

	return metadata
}
