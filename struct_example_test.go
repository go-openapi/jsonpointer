package jsonpointer_test

import (
	"errors"
	"fmt"

	"github.com/go-openapi/jsonpointer"
)

var ErrExampleIface = errors.New("example error")

type ExampleDoc struct {
	PromotedDoc

	Promoted     EmbeddedDoc `json:"promoted"`
	AnonPromoted EmbeddedDoc `json:"-"`
	A            string      `json:"propA"`
	Ignored      string      `json:"-"`
	Untagged     string

	unexported string
}

type EmbeddedDoc struct {
	B string `json:"propB"`
}

type PromotedDoc struct {
	C string `json:"propC"`
}

func Example_struct() {
	doc := ExampleDoc{
		PromotedDoc: PromotedDoc{
			C: "c",
		},
		Promoted: EmbeddedDoc{
			B: "promoted",
		},
		A:          "a",
		Ignored:    "ignored",
		unexported: "unexported",
	}

	{
		// tagged simple field
		pointerA, _ := jsonpointer.New("/propA")
		a, _, err := pointerA.Get(doc)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("a: %v\n", a)
	}

	{
		// tagged struct field is resolved
		pointerB, _ := jsonpointer.New("/promoted/propB")
		b, _, err := pointerB.Get(doc)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("b: %v\n", b)
	}

	{
		// tagged embedded field is resolved
		pointerC, _ := jsonpointer.New("/propC")
		c, _, err := pointerC.Get(doc)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("c: %v\n", c)
	}

	{
		// exlicitly ignored by JSON tag.
		pointerI, _ := jsonpointer.New("/ignored")
		_, _, err := pointerI.Get(doc)
		fmt.Printf("ignored: %v\n", err)
	}

	{
		// unexported field is ignored: use [JSONPointable] to alter this behavior.
		pointerX, _ := jsonpointer.New("/unexported")
		_, _, err := pointerX.Get(doc)
		fmt.Printf("unexported: %v\n", err)
	}

	{
		// Limitation: anonymous field is not resolved.
		pointerC, _ := jsonpointer.New("/propB")
		_, _, err := pointerC.Get(doc)
		fmt.Printf("anonymous: %v\n", err)
	}

	{
		// Limitation: untagged exported field is ignored, unlike with json standard MarshalJSON.
		pointerU, _ := jsonpointer.New("/untagged")
		_, _, err := pointerU.Get(doc)
		fmt.Printf("untagged: %v\n", err)
	}

	// output:
	// a: a
	// b: promoted
	// c: c
	// ignored: object has no field "ignored": JSON pointer error
	// unexported: object has no field "unexported": JSON pointer error
	// anonymous: object has no field "propB": JSON pointer error
	// untagged: object has no field "untagged": JSON pointer error
}
