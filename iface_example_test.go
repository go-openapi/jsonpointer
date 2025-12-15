// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer_test

import (
	"fmt"

	"github.com/go-openapi/jsonpointer"
)

var (
	_ jsonpointer.JSONPointable = CustomDoc{}
	_ jsonpointer.JSONSetable   = &CustomDoc{}
)

// CustomDoc accepts 2 preset properties "propA" and "propB", plus any number of extra properties.
//
// All values are strings.
type CustomDoc struct {
	a string
	b string
	c map[string]string
}

// JSONLookup implements [jsonpointer.JSONPointable].
func (d CustomDoc) JSONLookup(key string) (any, error) {
	switch key {
	case "propA":
		return d.a, nil
	case "propB":
		return d.b, nil
	default:
		if len(d.c) == 0 {
			return nil, fmt.Errorf("key %q not found: %w", key, ErrExampleIface)
		}
		extra, ok := d.c[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found: %w", key, ErrExampleIface)
		}

		return extra, nil
	}
}

// JSONSet implements [jsonpointer.JSONSetable].
func (d *CustomDoc) JSONSet(key string, value any) error {
	asString, ok := value.(string)
	if !ok {
		return fmt.Errorf("a CustomDoc only access strings as values, but got %T: %w", value, ErrExampleIface)
	}

	switch key {
	case "propA":
		d.a = asString

		return nil
	case "propB":
		d.b = asString

		return nil
	default:
		if len(d.c) == 0 {
			d.c = make(map[string]string)
		}
		d.c[key] = asString

		return nil
	}
}

func Example_iface() {
	doc := CustomDoc{
		a: "initial value for a",
		b: "initial value for b",
		// no extra values
	}

	pointerA, err := jsonpointer.New("/propA")
	if err != nil {
		fmt.Println(err)

		return
	}

	// get the initial value for a
	propA, kind, err := pointerA.Get(doc)
	if err != nil {
		fmt.Println(err)

		return
	}
	fmt.Printf("propA (%v): %v\n", kind, propA)

	pointerB, err := jsonpointer.New("/propB")
	if err != nil {
		fmt.Println(err)

		return
	}

	// get the initial value for b
	propB, kind, err := pointerB.Get(doc)
	if err != nil {
		fmt.Println(err)

		return
	}
	fmt.Printf("propB (%v): %v\n", kind, propB)

	pointerC, err := jsonpointer.New("/extra")
	if err != nil {
		fmt.Println(err)

		return
	}

	// not found yet
	_, _, err = pointerC.Get(doc)
	fmt.Printf("propC: %v\n", err)

	_, err = pointerA.Set(&doc, "new value for a") // doc is updated in place
	if err != nil {
		fmt.Println(err)

		return
	}

	_, err = pointerB.Set(&doc, "new value for b")
	if err != nil {
		fmt.Println(err)

		return
	}

	_, err = pointerC.Set(&doc, "new extra value")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("updated doc: %v", doc)

	// output:
	// propA (string): initial value for a
	// propB (string): initial value for b
	// propC: key "extra" not found: example error
	// updated doc: {new value for a new value for b map[extra:new extra value]}
}
