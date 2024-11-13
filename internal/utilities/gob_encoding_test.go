// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type testStruct struct {
	StringVal string
	Number    int
	BoolVal   bool
}

func TestGobEncodeDecode(t *testing.T) {
	testCase := testStruct{
		StringVal: "c70b284b254ce74e224242d69cbc1e33",
		Number:    24,
		BoolVal:   true,
	}

	encodedString, err := utilities.GobEncode(testCase)
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

	if err := utilities.GobDecode(encodedString, &got); err != nil {
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
			"FAILED test %s: Unexpected value returned after GobDecode:\nwant: %+v\ngot:  %+v",
			t.Name(),
			testCase,
			got,
		)
	} else {
		t.Logf(
			"Expected value returned after GobDecode:\ngot:  %+v",
			got,
		)
	}
}
