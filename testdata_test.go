// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	_ "embed" // initialize embed
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

//go:embed testdata/*.json
var testDocumentJSONBytes []byte

func testDocumentJSON(t *testing.T) any {
	t.Helper()

	var document any
	require.NoError(t, json.Unmarshal(testDocumentJSONBytes, &document))

	return document
}

func testStructJSONDoc(t *testing.T) testStructJSON {
	t.Helper()

	var document testStructJSON
	require.NoError(t, json.Unmarshal(testDocumentJSONBytes, &document))

	return document
}

func testStructJSONPtr(t *testing.T) *testStructJSON {
	t.Helper()

	document := testStructJSONDoc(t)

	return &document
}

// number of items in the test document
func testDocumentNBItems() int {
	return 11
}

// number of objects nodes in the test document
func testNodeObjNBItems() int {
	return 4
}

type testStructJSON struct {
	Foo []string `json:"foo"`
	Obj struct {
		A int   `json:"a"`
		B int   `json:"b"`
		C []int `json:"c"`
		D []struct {
			E int   `json:"e"`
			F []int `json:"f"`
		} `json:"d"`
	} `json:"obj"`
}

type aliasedMap map[string]any
