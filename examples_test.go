// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"encoding/json"
	"fmt"
)

type exampleDocument struct {
	Foo []string `json:"foo"`
}

func ExamplePointer_Get() {
	var doc exampleDocument

	if err := json.Unmarshal(testDocumentJSONBytes, &doc); err != nil { // populates doc
		panic(err)
	}

	pointer, err := New("/foo/1")
	if err != nil {
		panic(err)
	}

	value, kind, err := pointer.Get(doc)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"value: %q\nkind: %v\n",
		value, kind,
	)

	// Output:
	// value: "baz"
	// kind: string
}

func ExamplePointer_Set() {
	var doc exampleDocument

	if err := json.Unmarshal(testDocumentJSONBytes, &doc); err != nil { // populates doc
		panic(err)
	}

	pointer, err := New("/foo/1")
	if err != nil {
		panic(err)
	}

	result, err := pointer.Set(&doc, "hey my")
	if err != nil {
		panic(err)
	}

	fmt.Printf("result: %#v\n", result)
	fmt.Printf("doc: %#v\n", doc)

	// Output:
	// result: &jsonpointer.exampleDocument{Foo:[]string{"bar", "hey my"}}
	// doc: jsonpointer.exampleDocument{Foo:[]string{"bar", "hey my"}}
}
