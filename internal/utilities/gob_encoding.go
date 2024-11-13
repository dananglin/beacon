// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
)

// GobEncode encodes data of an arbitrary type using the gob library
// into a base64 encoded string.
func GobEncode(data any) (string, error) {
	if data == nil {
		return "", errors.New("cannot encode nil values")
	}

	buffer := new(bytes.Buffer)

	if err := gob.NewEncoder(buffer).Encode(data); err != nil {
		return "", fmt.Errorf("unable to encode the data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// GobDecode decodes a Base64 decoded string into an arbitrary data type using the
// gob library.
func GobDecode(encodedString string, data any) error {
	if data == nil {
		return errors.New("cannot decode nil values")
	}

	dst := make([]byte, base64.StdEncoding.DecodedLen(len(encodedString)))

	_, err := base64.StdEncoding.Decode(dst, []byte(encodedString))
	if err != nil {
		return fmt.Errorf("error decoding the Base64 encoded string: %w", err)
	}

	buffer := bytes.NewBuffer(dst)

	if err := gob.NewDecoder(buffer).Decode(data); err != nil {
		return fmt.Errorf("error gob decoding the data: %w", err)
	}

	return nil
}
