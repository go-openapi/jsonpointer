// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-openapi/swag/jsonname"
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

// ExamplePointer_Set_append demonstrates the RFC 6901 "-" token as an
// append operation on a slice. On nested slices reached through an
// addressable parent (map entry, pointer to struct, ...), the append is
// performed in place and the returned document is the same reference.
func ExamplePointer_Set_append() {
	doc := map[string]any{"foo": []any{"bar"}}

	pointer, err := New("/foo/-")
	if err != nil {
		fmt.Println(err)

		return
	}

	if _, err := pointer.Set(doc, "baz"); err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("doc: %v\n", doc["foo"])

	// Output:
	// doc: [bar baz]
}

// ExamplePointer_Set_appendTopLevelSlice shows the one case where the
// returned document is load-bearing: appending to a top-level slice
// passed by value. The library cannot rebind the slice header in the
// caller's variable, so callers must use the returned document (or pass
// *[]T to get in-place rebind).
func ExamplePointer_Set_appendTopLevelSlice() {
	doc := []int{1, 2}

	pointer, err := New("/-")
	if err != nil {
		fmt.Println(err)

		return
	}

	out, err := pointer.Set(doc, 3)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("original: %v\n", doc)
	fmt.Printf("returned: %v\n", out)

	// Output:
	// original: [1 2]
	// returned: [1 2 3]
}

// ExampleUseGoNameProvider contrasts the two [NameProvider] implementations
// shipped by [github.com/go-openapi/swag/jsonname]:
//
//   - the default provider requires a `json` struct tag to expose a field;
//   - the Go-name provider follows encoding/json conventions and accepts
//     exported untagged fields and promoted embedded fields as well.
func ExampleUseGoNameProvider() {
	type Embedded struct {
		Nested string // untagged: promoted only by the Go-name provider
	}
	type Doc struct {
		Embedded // untagged embedded: promoted only by the Go-name provider

		Tagged   string `json:"tagged"`
		Untagged string // no tag: visible only to the Go-name provider
	}

	doc := Doc{
		Embedded: Embedded{Nested: "promoted"},
		Tagged:   "hit",
		Untagged: "hidden-by-default",
	}

	for _, path := range []string{"/tagged", "/Untagged", "/Nested"} {
		p, err := New(path)
		if err != nil {
			fmt.Println(err)

			return
		}

		// Default provider: only the tagged field resolves.
		defV, _, defErr := p.Get(doc)
		// Go-name provider: untagged and promoted fields resolve too.
		goV, _, goErr := p.Get(doc, WithNameProvider(jsonname.NewGoNameProvider()))

		fmt.Printf("%s -> default=%v (err=%v) | goname=%v (err=%v)\n",
			path, defV, defErr != nil, goV, goErr != nil)
	}

	// Output:
	// /tagged -> default=hit (err=false) | goname=hit (err=false)
	// /Untagged -> default=<nil> (err=true) | goname=hidden-by-default (err=false)
	// /Nested -> default=<nil> (err=true) | goname=promoted (err=false)
}
