// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonname

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

type testNameStruct struct {
	Name       string `json:"name"`
	NotTheSame int64  `json:"plain"`
	Ignored    string `json:"-"`
}

func TestNameProvider(t *testing.T) {
	provider := NewNameProvider()

	obj := testNameStruct{}

	nm, ok := provider.GetGoName(obj, "name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "Name", nm)

	nm, ok = provider.GetGoName(obj, "plain")
	assert.TrueT(t, ok)
	assert.EqualT(t, "NotTheSame", nm)

	nm, ok = provider.GetGoName(obj, "doesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetGoName(obj, "ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	tpe := reflect.TypeFor[testNameStruct]()
	nm, ok = provider.GetGoNameForType(tpe, "name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "Name", nm)

	nm, ok = provider.GetGoNameForType(tpe, "plain")
	assert.TrueT(t, ok)
	assert.EqualT(t, "NotTheSame", nm)

	nm, ok = provider.GetGoNameForType(tpe, "doesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetGoNameForType(tpe, "ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	ptr := &obj
	nm, ok = provider.GetGoName(ptr, "name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "Name", nm)

	nm, ok = provider.GetGoName(ptr, "plain")
	assert.TrueT(t, ok)
	assert.EqualT(t, "NotTheSame", nm)

	nm, ok = provider.GetGoName(ptr, "doesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetGoName(ptr, "ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONName(obj, "Name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "name", nm)

	nm, ok = provider.GetJSONName(obj, "NotTheSame")
	assert.TrueT(t, ok)
	assert.EqualT(t, "plain", nm)

	nm, ok = provider.GetJSONName(obj, "DoesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONName(obj, "Ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONNameForType(tpe, "Name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "name", nm)

	nm, ok = provider.GetJSONNameForType(tpe, "NotTheSame")
	assert.TrueT(t, ok)
	assert.EqualT(t, "plain", nm)

	nm, ok = provider.GetJSONNameForType(tpe, "doesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONNameForType(tpe, "Ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONName(ptr, "Name")
	assert.TrueT(t, ok)
	assert.EqualT(t, "name", nm)

	nm, ok = provider.GetJSONName(ptr, "NotTheSame")
	assert.TrueT(t, ok)
	assert.EqualT(t, "plain", nm)

	nm, ok = provider.GetJSONName(ptr, "doesNotExist")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nm, ok = provider.GetJSONName(ptr, "Ignored")
	assert.FalseT(t, ok)
	assert.Empty(t, nm)

	nms := provider.GetJSONNames(ptr)
	assert.Len(t, nms, 2)

	assert.Len(t, provider.index, 1)
}

type EmbeddedBase struct {
	ID string `json:"id"`
}

// WithEmbeddedPointer embeds a pointer to a struct, the shape that used to panic in
// buildnameIndex with "reflect: NumField of non-struct type *jsonname.EmbeddedBase".
type WithEmbeddedPointer struct {
	*EmbeddedBase

	Name string `json:"name"`
}

type NamedSlice []string

// WithEmbeddedNonStruct embeds a named slice type, which carries no field to promote.
type WithEmbeddedNonStruct struct {
	NamedSlice

	Name string `json:"name"`
}

type WithEmbeddedTaggedNonStruct struct {
	NamedSlice `json:"list"`

	Name string `json:"name"`
}

func TestNameProviderAnonymousFields(t *testing.T) {
	t.Run("should promote fields of an embedded pointer to struct", func(t *testing.T) {
		provider := NewNameProvider()

		nm, ok := provider.GetGoName(WithEmbeddedPointer{EmbeddedBase: nil, Name: ""}, "id")
		assert.TrueT(t, ok)
		assert.EqualT(t, "ID", nm)

		nm, ok = provider.GetGoName(WithEmbeddedPointer{EmbeddedBase: nil, Name: ""}, "name")
		assert.TrueT(t, ok)
		assert.EqualT(t, "Name", nm)
	})

	t.Run("should index a struct embedding a non-struct type", func(t *testing.T) {
		provider := NewNameProvider()

		nm, ok := provider.GetGoName(WithEmbeddedNonStruct{NamedSlice: nil, Name: ""}, "name")
		assert.TrueT(t, ok)
		assert.EqualT(t, "Name", nm)
	})

	t.Run("should index a struct embedding a tagged non-struct type", func(t *testing.T) {
		provider := NewNameProvider()

		nm, ok := provider.GetGoName(WithEmbeddedTaggedNonStruct{NamedSlice: nil, Name: ""}, "name")
		assert.TrueT(t, ok)
		assert.EqualT(t, "Name", nm)
	})
}

func TestNameProviderNonStructSubjects(t *testing.T) {
	// a subject that is not a struct carries no json name: it indexes to nothing rather than
	// panicking in reflect.Type.NumField or reflect.Value.Type.
	for _, subject := range []any{
		map[string]any{"a": 1},
		42,
		"a string",
		[]int{1, 2},
		nil,
	} {
		t.Run(fmt.Sprintf("subject %T", subject), func(t *testing.T) {
			t.Run("NameProvider", func(t *testing.T) {
				provider := NewNameProvider()
				assert.Empty(t, provider.GetJSONNames(subject))

				nm, ok := provider.GetGoName(subject, "a")
				assert.FalseT(t, ok)
				assert.Empty(t, nm)

				nm, ok = provider.GetJSONName(subject, "A")
				assert.FalseT(t, ok)
				assert.Empty(t, nm)
			})

			t.Run("GoNameProvider", func(t *testing.T) {
				provider := NewGoNameProvider()
				assert.Empty(t, provider.GetJSONNames(subject))

				nm, ok := provider.GetGoName(subject, "a")
				assert.FalseT(t, ok)
				assert.Empty(t, nm)

				nm, ok = provider.GetJSONName(subject, "A")
				assert.FalseT(t, ok)
				assert.Empty(t, nm)
			})
		})
	}
}
