// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestEscaping(t *testing.T) {
	t.Parallel()

	t.Run("escaped pointer strings against test document", func(t *testing.T) {
		ins := []string{`/`, `/`, `/a~1b`, `/a~1b`, `/c%d`, `/e^f`, `/g|h`, `/i\j`, `/k"l`, `/ `, `/m~0n`}
		outs := []float64{0, 0, 1, 1, 2, 3, 4, 5, 6, 7, 8}

		for i := range ins {
			t.Run("should create a JSONPointer", func(t *testing.T) {
				p, err := New(ins[i])
				require.NoError(t, err, "input: %v", ins[i])

				t.Run("should get JSONPointer from document", func(t *testing.T) {
					result, _, err := p.Get(testDocumentJSON(t))
					require.NoError(t, err, "input: %v", ins[i])
					assert.InDeltaf(t, outs[i], result, 1e-6, "input: %v", ins[i])
				})
			})
		}
	})

	t.Run("special escapes", func(t *testing.T) {
		t.Parallel()

		t.Run("with escape then unescape", func(t *testing.T) {
			const original = "a/"

			t.Run("unescaping an escaped string should yield the original", func(t *testing.T) {
				esc := Escape(original)
				assert.Equal(t, "a~1", esc)

				unesc := Unescape(esc)
				assert.Equal(t, original, unesc)
			})
		})

		t.Run("with multiple escapes", func(t *testing.T) {
			unesc := Unescape("~01")
			assert.Equal(t, "~1", unesc)
			assert.Equal(t, "~01", Escape(unesc))

			const (
				original = "~/"
				escaped  = "~0~1"
			)

			assert.Equal(t, escaped, Escape(original))
			assert.Equal(t, original, Unescape(escaped))
		})

		t.Run("with escaped characters in pointer", func(t *testing.T) {
			t.Run("escaped ~", func(t *testing.T) {
				s := Escape("m~n")
				assert.Equal(t, "m~0n", s)
			})
			t.Run("escaped /", func(t *testing.T) {
				s := Escape("m/n")
				assert.Equal(t, "m~1n", s)
			})
		})
	})
}

func TestFullDocument(t *testing.T) {
	t.Parallel()

	t.Run("with empty pointer", func(t *testing.T) {
		const in = ``

		p, err := New(in)
		require.NoErrorf(t, err, "New(%v) error %v", in, err)

		t.Run("should resolve full doc", func(t *testing.T) {
			result, _, err := p.Get(testDocumentJSON(t))
			require.NoErrorf(t, err, "Get(%v) error %v", in, err)

			asMap, ok := result.(map[string]any)
			require.True(t, ok)

			require.Lenf(t, asMap, testDocumentNBItems(), "Get(%v) = %v, expect full document", in, result)
		})

		t.Run("should resolve full doc, with nil name provider", func(t *testing.T) {
			result, _, err := p.get(testDocumentJSON(t), nil)
			require.NoErrorf(t, err, "Get(%v) error %v", in, err)

			asMap, ok := result.(map[string]any)
			require.True(t, ok)
			require.Lenf(t, asMap, testDocumentNBItems(), "Get(%v) = %v, expect full document", in, result)

			t.Run("should set value in doc, with nil name provider", func(t *testing.T) {
				setter, err := New("/foo/0")
				require.NoErrorf(t, err, "New(%v) error %v", in, err)

				const value = "hey"
				require.NoError(t, setter.set(asMap, value, nil))

				foos, ok := asMap["foo"]
				require.True(t, ok)

				asArray, ok := foos.([]any)
				require.True(t, ok)
				require.Len(t, asArray, 2)

				foo := asArray[0]
				bar, ok := foo.(string)
				require.True(t, ok)

				require.Equal(t, value, bar)
			})
		})
	})
}

func TestDecodedTokens(t *testing.T) {
	t.Parallel()

	p, err := New("/obj/a~1b")
	require.NoError(t, err)
	assert.Equal(t, []string{"obj", "a/b"}, p.DecodedTokens())
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("with empty pointer", func(t *testing.T) {
		p, err := New("")
		require.NoError(t, err)

		assert.True(t, p.IsEmpty())
	})

	t.Run("with non-empty pointer", func(t *testing.T) {
		p, err := New("/obj")
		require.NoError(t, err)

		assert.False(t, p.IsEmpty())
	})
}

func TestGetSingle(t *testing.T) {
	t.Parallel()

	const key = "obj"

	t.Run("should create a new JSON pointer", func(t *testing.T) {
		const in = "/" + key

		_, err := New(in)
		require.NoError(t, err)
	})

	t.Run(fmt.Sprintf("should find token %q in JSON", key), func(t *testing.T) {
		result, _, err := GetForToken(testDocumentJSON(t), key)
		require.NoError(t, err)
		assert.Len(t, result, testNodeObjNBItems())
	})

	t.Run(fmt.Sprintf("should find token %q in type alias interface", key), func(t *testing.T) {
		type alias any
		var in alias = testDocumentJSON(t)

		result, _, err := GetForToken(in, key)
		require.NoError(t, err)
		assert.Len(t, result, testNodeObjNBItems())
	})

	t.Run(fmt.Sprintf("should find token %q in pointer to interface", key), func(t *testing.T) {
		in := testDocumentJSON(t)

		result, _, err := GetForToken(&in, key)
		require.NoError(t, err)
		assert.Len(t, result, testNodeObjNBItems())
	})

	t.Run(`should NOT find token "Obj" in struct`, func(t *testing.T) {
		result, _, err := GetForToken(testStructJSONDoc(t), "Obj")
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(`should not find token "Obj2" in struct`, func(t *testing.T) {
		result, _, err := GetForToken(testStructJSONDoc(t), "Obj2")
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("should not find token in nil", func(t *testing.T) {
		result, _, err := GetForToken(nil, key)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("should not find token in nil interface", func(t *testing.T) {
		var in any

		result, _, err := GetForToken(in, key)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

type pointableImpl struct {
	a string
}

func (p pointableImpl) JSONLookup(token string) (any, error) {
	if token == "some" {
		return p.a, nil
	}
	return nil, fmt.Errorf("object has no field %q: %w", token, ErrPointer)
}

type pointableMap map[string]string

func (p pointableMap) JSONLookup(token string) (any, error) {
	if token == "swap" {
		return p["swapped"], nil
	}

	v, ok := p[token]
	if ok {
		return v, nil
	}

	return nil, fmt.Errorf("object has no key %q: %w", token, ErrPointer)
}

func TestPointableInterface(t *testing.T) {
	t.Parallel()

	t.Run("with pointable type", func(t *testing.T) {
		p := &pointableImpl{"hello"}
		result, _, err := GetForToken(p, "some")
		require.NoError(t, err)
		assert.Equal(t, p.a, result)

		result, _, err = GetForToken(p, "something")
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with pointable map", func(t *testing.T) {
		p := pointableMap{"swapped": "hello", "a": "world"}
		result, _, err := GetForToken(p, "swap")
		require.NoError(t, err)
		assert.Equal(t, p["swapped"], result)

		result, _, err = GetForToken(p, "a")
		require.NoError(t, err)
		assert.Equal(t, p["a"], result)
	})
}

func TestGetNode(t *testing.T) {
	t.Parallel()

	const in = `/obj`

	t.Run("should build pointer", func(t *testing.T) {
		p, err := New(in)
		require.NoError(t, err)

		t.Run("should resolve pointer against document", func(t *testing.T) {
			result, _, err := p.Get(testDocumentJSON(t))
			require.NoError(t, err)
			assert.Len(t, result, testNodeObjNBItems())
		})

		t.Run("with aliased map", func(t *testing.T) {
			asMap, ok := testDocumentJSON(t).(map[string]any)
			require.True(t, ok)
			alias := aliasedMap(asMap)

			result, _, err := p.Get(alias)
			require.NoError(t, err)
			assert.Len(t, result, testNodeObjNBItems())
		})

		t.Run("with struct", func(t *testing.T) {
			doc := testStructJSONDoc(t)
			expected := testStructJSONDoc(t).Obj

			result, _, err := p.Get(doc)
			require.NoError(t, err)
			assert.Equal(t, expected, result)
		})

		t.Run("with pointer to struct", func(t *testing.T) {
			doc := testStructJSONPtr(t)
			expected := testStructJSONDoc(t).Obj

			result, _, err := p.Get(doc)
			require.NoError(t, err)
			assert.Equal(t, expected, result)
		})
	})
}

func TestArray(t *testing.T) {
	t.Parallel()

	ins := []string{`/foo/0`, `/foo/0`, `/foo/1`}
	outs := []string{"bar", "bar", "baz"}

	for i, pointer := range ins {
		expected := outs[i]

		t.Run(fmt.Sprintf("with pointer %q", pointer), func(t *testing.T) {
			p, err := New(pointer)
			require.NoError(t, err)

			t.Run("should resolve against struct", func(t *testing.T) {
				result, _, err := p.Get(testStructJSONDoc(t))
				require.NoError(t, err)
				assert.Equal(t, expected, result)
			})

			t.Run("should resolve against pointer to struct", func(t *testing.T) {
				result, _, err := p.Get(testStructJSONPtr(t))
				require.NoError(t, err)
				assert.Equal(t, expected, result)
			})

			t.Run("should resolve against dynamic JSON map", func(t *testing.T) {
				result, _, err := p.Get(testDocumentJSON(t))
				require.NoError(t, err)
				assert.Equal(t, expected, result)
			})
		})
	}
}

func TestOtherThings(t *testing.T) {
	t.Parallel()

	t.Run("single string pointer should be valid", func(t *testing.T) {
		_, err := New("abc")
		require.Error(t, err)
	})

	t.Run("empty string pointer should be valid", func(t *testing.T) {
		p, err := New("")
		require.NoError(t, err)
		assert.Empty(t, p.String())
	})

	t.Run("string representation of a pointer", func(t *testing.T) {
		p, err := New("/obj/a")
		require.NoError(t, err)
		assert.Equal(t, "/obj/a", p.String())
	})

	t.Run("out of bound array index should error", func(t *testing.T) {
		p, err := New("/foo/3")
		require.NoError(t, err)

		_, _, err = p.Get(testDocumentJSON(t))
		require.Error(t, err)
	})

	t.Run("referring to a key in an array should error", func(t *testing.T) {
		p, err := New("/foo/a")
		require.NoError(t, err)
		_, _, err = p.Get(testDocumentJSON(t))
		require.Error(t, err)
	})

	t.Run("referring to a non-existing key in an array should error", func(t *testing.T) {
		p, err := New("/notthere")
		require.NoError(t, err)
		_, _, err = p.Get(testDocumentJSON(t))
		require.Error(t, err)
	})

	t.Run("resolving pointer against an unsupported type (int) should error", func(t *testing.T) {
		p, err := New("/invalid")
		require.NoError(t, err)
		_, _, err = p.Get(1234)
		require.Error(t, err)
	})

	t.Run("with pointer to an array index", func(t *testing.T) {
		for index := range 2 {
			p, err := New(fmt.Sprintf("/foo/%d", index))
			require.NoError(t, err)

			v, _, err := p.Get(testDocumentJSON(t))
			require.NoError(t, err)

			expected := extractFooKeyIndex(t, index)
			assert.Equal(t, expected, v)
		}
	})
}

func extractFooKeyIndex(t *testing.T, index int) any {
	t.Helper()

	asMap, ok := testDocumentJSON(t).(map[string]any)
	require.True(t, ok)

	// {"foo": [ ... ] }
	bbb, ok := asMap["foo"]
	require.True(t, ok)

	asArray, ok := bbb.([]any)
	require.True(t, ok)

	return asArray[index]
}

func TestObject(t *testing.T) {
	t.Parallel()

	ins := []string{`/obj/a`, `/obj/b`, `/obj/c/0`, `/obj/c/1`, `/obj/c/1`, `/obj/d/1/f/0`}
	outs := []float64{1, 2, 3, 4, 4, 50}

	for i := range ins {
		p, err := New(ins[i])
		require.NoError(t, err)

		result, _, err := p.Get(testDocumentJSON(t))
		require.NoError(t, err)
		assert.InDelta(t, outs[i], result, 1e-6)

		result, _, err = p.Get(testStructJSONDoc(t))
		require.NoError(t, err)
		assert.InDelta(t, outs[i], result, 1e-6)

		result, _, err = p.Get(testStructJSONPtr(t))
		require.NoError(t, err)
		assert.InDelta(t, outs[i], result, 1e-6)
	}
}

type setJSONDoc struct {
	A []struct {
		B int `json:"b"`
		C int `json:"c"`
	} `json:"a"`
	D int `json:"d"`
}

type settableDoc struct {
	Coll settableColl
	Int  settableInt
}

func (s settableDoc) MarshalJSON() ([]byte, error) {
	var res struct {
		A settableColl `json:"a"`
		D settableInt  `json:"d"`
	}
	res.A = s.Coll
	res.D = s.Int
	return json.Marshal(res)
}
func (s *settableDoc) UnmarshalJSON(data []byte) error {
	var res struct {
		A settableColl `json:"a"`
		D settableInt  `json:"d"`
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	s.Coll = res.A
	s.Int = res.D
	return nil
}

// JSONLookup implements an interface to customize json pointer lookup
func (s settableDoc) JSONLookup(token string) (any, error) {
	switch token {
	case "a":
		return &s.Coll, nil
	case "d":
		return &s.Int, nil
	default:
		return nil, fmt.Errorf("%s is not a known field: %w", token, ErrPointer)
	}
}

// JSONLookup implements an interface to customize json pointer lookup
func (s *settableDoc) JSONSet(token string, data any) error {
	switch token {
	case "a":
		switch dt := data.(type) {
		case settableColl:
			s.Coll = dt
			return nil
		case *settableColl:
			if dt != nil {
				s.Coll = *dt
			} else {
				s.Coll = settableColl{}
			}
			return nil
		case []settableCollItem:
			s.Coll.Items = dt
			return nil
		}
	case "d":
		switch dt := data.(type) {
		case settableInt:
			s.Int = dt
			return nil
		case int:
			s.Int.Value = dt
			return nil
		case int8:
			s.Int.Value = int(dt)
			return nil
		case int16:
			s.Int.Value = int(dt)
			return nil
		case int32:
			s.Int.Value = int(dt)
			return nil
		case int64:
			s.Int.Value = int(dt)
			return nil
		default:
			return fmt.Errorf("invalid type %T for %s: %w", data, token, ErrPointer)
		}
	}
	return fmt.Errorf("%s is not a known field: %w", token, ErrPointer)
}

type settableColl struct {
	Items []settableCollItem
}

func (s settableColl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Items)
}
func (s *settableColl) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Items)
}

// JSONLookup implements an interface to customize json pointer lookup
func (s settableColl) JSONLookup(token string) (any, error) {
	if tok, err := strconv.Atoi(token); err == nil {
		return &s.Items[tok], nil
	}
	return nil, fmt.Errorf("%s is not a valid index: %w", token, ErrPointer)
}

// JSONLookup implements an interface to customize json pointer lookup
func (s *settableColl) JSONSet(token string, data any) error {
	if _, err := strconv.Atoi(token); err == nil {
		_, err := SetForToken(s.Items, token, data)
		return err
	}
	return fmt.Errorf("%s is not a valid index: %w", token, ErrPointer)
}

type settableCollItem struct {
	B int `json:"b"`
	C int `json:"c"`
}

type settableInt struct {
	Value int
}

func (s settableInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}
func (s *settableInt) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Value)
}

func TestSetNode(t *testing.T) {
	t.Parallel()

	const jsonText = `{"a":[{"b": 1, "c": 2}], "d": 3}`

	var jsonDocument any
	require.NoError(t, json.Unmarshal([]byte(jsonText), &jsonDocument))

	t.Run("with set node c", func(t *testing.T) {
		const in = "/a/0/c"
		p, err := New(in)
		require.NoError(t, err)

		_, err = p.Set(jsonDocument, 999)
		require.NoError(t, err)

		firstNode, ok := jsonDocument.(map[string]any)
		require.True(t, ok)
		assert.Len(t, firstNode, 2)

		sliceNode, ok := firstNode["a"].([]any)
		require.True(t, ok)
		assert.Len(t, sliceNode, 1)

		changedNode, ok := sliceNode[0].(map[string]any)
		require.True(t, ok)
		chNodeVI := changedNode["c"]

		require.IsType(t, 0, chNodeVI)
		changedNodeValue, ok := chNodeVI.(int)
		require.True(t, ok)

		require.Equal(t, 999, changedNodeValue)
		assert.Len(t, sliceNode, 1)
	})

	t.Run("with set node 0 with map", func(t *testing.T) {
		v, err := New("/a/0")
		require.NoError(t, err)

		_, err = v.Set(jsonDocument, map[string]any{"b": 3, "c": 8})
		require.NoError(t, err)

		firstNode, ok := jsonDocument.(map[string]any)
		require.True(t, ok)
		assert.Len(t, firstNode, 2)

		sliceNode, ok := firstNode["a"].([]any)
		require.True(t, ok)
		assert.Len(t, sliceNode, 1)

		changedNode, ok := sliceNode[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 3, changedNode["b"])
		assert.Equal(t, 8, changedNode["c"])
	})

	t.Run("with struct", func(t *testing.T) {
		var structDoc setJSONDoc
		require.NoError(t, json.Unmarshal([]byte(jsonText), &structDoc))

		t.Run("with set array node", func(t *testing.T) {
			g, err := New("/a")
			require.NoError(t, err)

			_, err = g.Set(&structDoc, []struct {
				B int `json:"b"`
				C int `json:"c"`
			}{{B: 4, C: 7}})
			require.NoError(t, err)
			assert.Len(t, structDoc.A, 1)
			changedNode := structDoc.A[0]
			assert.Equal(t, 4, changedNode.B)
			assert.Equal(t, 7, changedNode.C)
		})

		t.Run("with set node 0 with struct", func(t *testing.T) {
			v, err := New("/a/0")
			require.NoError(t, err)

			_, err = v.Set(structDoc, struct {
				B int `json:"b"`
				C int `json:"c"`
			}{B: 3, C: 8})
			require.NoError(t, err)
			assert.Len(t, structDoc.A, 1)
			changedNode := structDoc.A[0]
			assert.Equal(t, 3, changedNode.B)
			assert.Equal(t, 8, changedNode.C)
		})

		t.Run("with set node c with struct", func(t *testing.T) {
			p, err := New("/a/0/c")
			require.NoError(t, err)

			_, err = p.Set(&structDoc, 999)
			require.NoError(t, err)

			require.Len(t, structDoc.A, 1)
			assert.Equal(t, 999, structDoc.A[0].C)
		})
	})

	t.Run("with Settable", func(t *testing.T) {
		var setDoc settableDoc
		require.NoError(t, json.Unmarshal([]byte(jsonText), &setDoc))

		t.Run("with array node a", func(t *testing.T) {
			g, err := New("/a")
			require.NoError(t, err)

			_, err = g.Set(&setDoc, []settableCollItem{{B: 4, C: 7}})
			require.NoError(t, err)
			assert.Len(t, setDoc.Coll.Items, 1)
			changedNode := setDoc.Coll.Items[0]
			assert.Equal(t, 4, changedNode.B)
			assert.Equal(t, 7, changedNode.C)
		})

		t.Run("with node 0", func(t *testing.T) {
			v, err := New("/a/0")
			require.NoError(t, err)

			_, err = v.Set(setDoc, settableCollItem{B: 3, C: 8})
			require.NoError(t, err)
			assert.Len(t, setDoc.Coll.Items, 1)
			changedNode := setDoc.Coll.Items[0]
			assert.Equal(t, 3, changedNode.B)
			assert.Equal(t, 8, changedNode.C)
		})

		t.Run("with node c", func(t *testing.T) {
			p, err := New("/a/0/c")
			require.NoError(t, err)
			_, err = p.Set(setDoc, 999)
			require.NoError(t, err)
			require.Len(t, setDoc.Coll.Items, 1)
			assert.Equal(t, 999, setDoc.Coll.Items[0].C)
		})
	})

	t.Run("with nil traversal panic", func(t *testing.T) {
		// This test exposes the panic that occurs when trying to set a value
		// through a path that contains nil intermediate values
		data := map[string]any{
			"level1": map[string]any{
				"level2": map[string]any{
					"level3": nil, // This nil causes the panic
				},
			},
		}

		ptr, err := New("/level1/level2/level3/value")
		require.NoError(t, err)

		// This should return an error, not panic
		_, err = ptr.Set(data, "test-value")

		// The library should handle this gracefully and return an error
		// instead of panicking
		require.Error(t, err, "Setting value through nil intermediate path should return an error, not panic")
	})

	t.Run("with direct nil map value", func(t *testing.T) {
		// Simpler test case that directly tests nil traversal
		data := map[string]any{
			"container": nil,
		}

		ptr, err := New("/container/nested/value")
		require.NoError(t, err)

		// Attempting to traverse through nil should return an error, not panic
		_, err = ptr.Set(data, "test")
		require.Error(t, err, "Cannot traverse through nil intermediate values")
	})

	t.Run("with nil in nested structure", func(t *testing.T) {
		// Test case with multiple nil values in nested structure
		data := map[string]any{
			"config": map[string]any{
				"settings": nil,
			},
			"data": map[string]any{
				"nested": map[string]any{
					"properties": map[string]any{
						"attributes": nil, // Nil intermediate value
					},
				},
			},
		}

		ptr, err := New("/data/nested/properties/attributes/name")
		require.NoError(t, err)

		// Should return error, not panic
		_, err = ptr.Set(data, "test-name")
		require.Error(t, err, "Setting through nil intermediate path should return error")
	})

	t.Run("with path creation through nil intermediate", func(t *testing.T) {
		// Test case that simulates path creation functions encountering nil
		// This happens when tools try to create missing paths but encounter nil intermediate values
		data := map[string]any{
			"spec": map[string]any{
				"template": nil, // This blocks path creation attempts
			},
		}

		// Attempting to create a path like /spec/template/metadata/labels should fail gracefully
		ptr, err := New("/spec/template/metadata")
		require.NoError(t, err)

		// Should return error when trying to set on nil intermediate during path creation
		_, err = ptr.Set(data, map[string]any{"labels": map[string]any{}})
		require.Error(t, err, "Setting on nil intermediate during path creation should return error")
	})

	t.Run("with SetForToken on nil", func(t *testing.T) {
		// Test the single-level SetForToken function with nil
		data := map[string]any{
			"container": nil,
		}

		// Should handle nil gracefully at single token level
		_, err := SetForToken(data["container"], "nested", "value")
		require.Error(t, err, "SetForToken on nil should return error, not panic")
	})
}

func TestOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		ptr      string
		input    string
		offset   int64
		hasError bool
	}{
		{
			name:   "object key",
			ptr:    "/foo/bar",
			input:  `{"foo": {"bar": 21}}`,
			offset: 9,
		},
		{
			name:   "complex object key",
			ptr:    "/paths/~1p~1{}/get",
			input:  `{"paths": {"foo": {"bar": 123, "baz": {}}, "/p/{}": {"get": {}}}}`,
			offset: 53,
		},
		{
			name:   "array index",
			ptr:    "/0/1",
			input:  `[[1,2], [3,4]]`,
			offset: 3,
		},
		{
			name:   "mix array index and object key",
			ptr:    "/0/1/foo/0",
			input:  `[[1, {"foo": ["a", "b"]}], [3, 4]]`,
			offset: 14,
		},
		{
			name:     "nonexist object key",
			ptr:      "/foo/baz",
			input:    `{"foo": {"bar": 21}}`,
			hasError: true,
		},
		{
			name:     "nonexist array index",
			ptr:      "/0/2",
			input:    `[[1,2], [3,4]]`,
			hasError: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ptr, err := New(tt.ptr)
			require.NoError(t, err)

			offset, err := ptr.Offset(tt.input)
			if tt.hasError {
				require.Error(t, err)
				return
			}

			t.Log(offset, err)
			require.NoError(t, err)
			assert.Equal(t, tt.offset, offset)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("set at pointer against an unsupported type (int) should error", func(t *testing.T) {
		p, err := New("/invalid")
		require.NoError(t, err)
		_, err = p.Set(1, 1234)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnsupportedValueType)
	})

	t.Run("set with empty pointer", func(t *testing.T) {
		p, err := New("")
		require.NoError(t, err)

		doc := testDocumentJSON(t)
		newDoc, err := p.Set(doc, 1)
		require.NoError(t, err)

		require.Equal(t, doc, newDoc)
	})
}
