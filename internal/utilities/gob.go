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
	"io"
)

var (
	ErrCannotEncodeNil = errors.New("cannot encode nil values")
	ErrCannotDecodeNil = errors.New("cannot decode nil values")
)

// GobBase64Encode encodes data of an arbitrary type using the gob library
// into a base64 encoded string.
func GobBase64Encode(data any) (string, error) {
	dataBytes, err := gobEncode(data)
	if err != nil {
		return "", fmt.Errorf("error using the Gob encoder: %w", err)
	}

	return base64.StdEncoding.EncodeToString(dataBytes), nil
}

// GobEncode encodes data of an arbitrary type using the gob library and
// returns the resulting slice of bytes.
func GobEncode(data any) ([]byte, error) {
	dataBytes, err := gobEncode(data)
	if err != nil {
		return []byte{}, fmt.Errorf("error using the Gob encoder: %w", err)
	}

	return dataBytes, nil
}

func gobEncode(data any) ([]byte, error) {
	if data == nil {
		return []byte{}, ErrCannotEncodeNil
	}

	buffer := new(bytes.Buffer)

	if err := gob.NewEncoder(buffer).Encode(data); err != nil {
		return []byte{}, fmt.Errorf("unable to encode the data: %w", err)
	}

	return buffer.Bytes(), nil
}

// GobBase64Decode decodes a Base64 decoded string into a value of an arbitrary type using the
// gob library.
func GobBase64Decode(encodedString string, data any) error {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(encodedString)))

	_, err := base64.StdEncoding.Decode(dst, []byte(encodedString))
	if err != nil {
		return fmt.Errorf("error decoding the Base64 encoded string: %w", err)
	}

	if err := gobDecode(bytes.NewBuffer(dst), data); err != nil {
		return fmt.Errorf("error using the Gob decoder: %w", err)
	}

	return nil
}

// GobDecode decodes data into a value of an arbitrary type using the gob library.
func GobDecode(reader io.Reader, data any) error {
	if err := gobDecode(reader, data); err != nil {
		return fmt.Errorf("error using the Gob decoder: %w", err)
	}

	return nil
}

func gobDecode(reader io.Reader, data any) error {
	if data == nil {
		return ErrCannotDecodeNil
	}

	if err := gob.NewDecoder(reader).Decode(data); err != nil {
		return fmt.Errorf("error gob decoding the data: %w", err)
	}

	return nil
}
