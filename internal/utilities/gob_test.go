// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"bytes"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type testStruct struct {
	StringVal string
	Number    int
	BoolVal   bool
}

func TestGob(t *testing.T) {
	testCase := testStruct{
		StringVal: "c70b284b254ce74e224242d69cbc1e33",
		Number:    24,
		BoolVal:   true,
	}

	t.Run("Test GobBase64EncodeDecode", testGobBase64EncodeDecode(testCase))
	t.Run("Test GobEncodeDecode", testGobEncodeDecode(testCase))
}

func testGobBase64EncodeDecode(testCase testStruct) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		encodedString, err := utilities.GobBase64Encode(testCase)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error encoding the test case: %v",
				t.Name(),
				err,
			)
		} else {
			t.Logf(
				"Successfully encoded the test case: got %s",
				encodedString,
			)
		}

		got := testStruct{}

		if err := utilities.GobBase64Decode(encodedString, &got); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error decoding the encoded string: %v",
				t.Name(),
				err,
			)
		} else {
			t.Log("Successfully decoded the encoded string.")
		}

		if !reflect.DeepEqual(testCase, got) {
			t.Errorf(
				"FAILED test %s: Unexpected value returned after decoding:\nwant: %+v\ngot:  %+v",
				t.Name(),
				testCase,
				got,
			)
		} else {
			t.Logf(
				"Expected value returned after decoding:\ngot:  %+v",
				got,
			)
		}
	}
}

func testGobEncodeDecode(testCase testStruct) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		dataBytes, err := utilities.GobEncode(testCase)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error encoding the test case: %v",
				t.Name(),
				err,
			)
		} else {
			t.Log("Successfully encoded the test case.")
		}

		got := testStruct{}

		if err := utilities.GobDecode(bytes.NewBuffer(dataBytes), &got); err != nil {
			t.Fatalf(
				"FAILED test %s: Received an error decoding the data: %v",
				t.Name(),
				err,
			)
		} else {
			t.Log("Successfully decoded the data.")
		}

		if !reflect.DeepEqual(testCase, got) {
			t.Errorf(
				"FAILED test %s: Unexpected value returned after decoding:\nwant: %+v\ngot:  %+v",
				t.Name(),
				testCase,
				got,
			)
		} else {
			t.Logf(
				"Expected value returned after decoding:\ngot:  %+v",
				got,
			)
		}
	}
}
