// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrExampleStruct = errors.New("example error")

type exampleDocument struct {
	Foo []string `json:"foo"`
}

func ExampleNew() {
	empty, err := New("")
	if err != nil {
		fmt.Println(err)

		return
	}
	fmt.Printf("empty pointer: %q\n", empty.String())

	key, err := New("/foo")
	if err != nil {
		fmt.Println(err)

		return
	}
	fmt.Printf("pointer to object key: %q\n", key.String())

	elem, err := New("/foo/1")
	if err != nil {
		fmt.Println(err)

		return
	}
	fmt.Printf("pointer to array element: %q\n", elem.String())

	escaped0, err := New("/foo~0")
	if err != nil {
		fmt.Println(err)

		return
	}
	// key contains "~"
	fmt.Printf("pointer to key %q: %q\n", Unescape("foo~0"), escaped0.String())

	escaped1, err := New("/foo~1")
	if err != nil {
		fmt.Println(err)

		return
	}
	// key contains "/"
	fmt.Printf("pointer to key %q: %q\n", Unescape("foo~1"), escaped1.String())

	// output:
	// empty pointer: ""
	// pointer to object key: "/foo"
	// pointer to array element: "/foo/1"
	// pointer to key "foo~": "/foo~0"
	// pointer to key "foo/": "/foo~1"
}

func ExamplePointer_Get() {
	var doc exampleDocument

	if err := json.Unmarshal(testDocumentJSONBytes, &doc); err != nil { // populates doc
		fmt.Println(err)

		return
	}

	pointer, err := New("/foo/1")
	if err != nil {
		fmt.Println(err)

		return
	}

	value, kind, err := pointer.Get(doc)
	if err != nil {
		fmt.Println(err)

		return
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
		fmt.Println(err)

		return
	}

	pointer, err := New("/foo/1")
	if err != nil {
		fmt.Println(err)

		return
	}

	result, err := pointer.Set(&doc, "hey my")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("result: %#v\n", result)
	fmt.Printf("doc: %#v\n", doc)

	// Output:
	// result: &jsonpointer.exampleDocument{Foo:[]string{"bar", "hey my"}}
	// doc: jsonpointer.exampleDocument{Foo:[]string{"bar", "hey my"}}
}
