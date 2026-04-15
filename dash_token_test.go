// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// RFC 6901 §4: the "-" token refers to the (nonexistent) element after the
// last array element. It is always an error on Get/Offset, valid only as
// the terminal token of a Set against a slice (append, per RFC 6902).

func TestDashToken_GetAlwaysErrors(t *testing.T) {
	t.Parallel()

	t.Run("terminal dash on slice in map", func(t *testing.T) {
		doc := map[string]any{"arr": []any{1, 2, 3}}
		p, err := New("/arr/-")
		require.NoError(t, err)

		_, _, err = p.Get(doc)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDashToken)
		require.ErrorIs(t, err, ErrPointer)
	})

	t.Run("terminal dash on top-level slice", func(t *testing.T) {
		doc := []int{1, 2, 3}
		p, err := New("/-")
		require.NoError(t, err)

		_, _, err = p.Get(doc)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDashToken)
	})

	t.Run("intermediate dash during get", func(t *testing.T) {
		doc := map[string]any{"arr": []any{map[string]any{"x": 1}}}
		p, err := New("/arr/-/x")
		require.NoError(t, err)

		_, _, err = p.Get(doc)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDashToken)
	})

	t.Run("GetForToken on slice with dash", func(t *testing.T) {
		_, _, err := GetForToken([]int{1, 2}, "-")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDashToken)
	})

	t.Run("dash on map key is a regular lookup, not an error", func(t *testing.T) {
		// "-" is only special for arrays. A literal "-" key in a map is fine.
		doc := map[string]any{"-": 42}
		p, err := New("/-")
		require.NoError(t, err)

		v, _, err := p.Get(doc)
		require.NoError(t, err)
		assert.Equal(t, 42, v)
	})
}

func TestDashToken_OffsetErrors(t *testing.T) {
	t.Parallel()

	doc := `{"arr":[1,2,3]}`
	p, err := New("/arr/-")
	require.NoError(t, err)

	_, err = p.Offset(doc)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrDashToken)
}

func TestDashToken_SetAppend(t *testing.T) {
	t.Parallel()

	t.Run("append into slice nested in a map (in place)", func(t *testing.T) {
		doc := map[string]any{"arr": []any{1, 2}}
		p, err := New("/arr/-")
		require.NoError(t, err)

		out, err := p.Set(doc, 3)
		require.NoError(t, err)

		// returned doc is the same map reference
		assert.Equal(t, doc, out)

		// map's slice was rebound in place
		arr, ok := doc["arr"].([]any)
		require.True(t, ok)
		assert.Equal(t, []any{1, 2, 3}, arr)
	})

	t.Run("append into top-level slice passed by value (return value is source of truth)", func(t *testing.T) {
		doc := []int{1, 2}
		p, err := New("/-")
		require.NoError(t, err)

		out, err := p.Set(doc, 3)
		require.NoError(t, err)

		// returned doc has the appended element
		outSlice, ok := out.([]int)
		require.True(t, ok)
		assert.Equal(t, []int{1, 2, 3}, outSlice)
	})

	t.Run("append into top-level *[]T (in place)", func(t *testing.T) {
		doc := []int{1, 2}
		p, err := New("/-")
		require.NoError(t, err)

		_, err = p.Set(&doc, 3)
		require.NoError(t, err)

		// caller's slice variable now has the appended element
		assert.Equal(t, []int{1, 2, 3}, doc)
	})

	t.Run("append into struct slice field reached via pointer (in place)", func(t *testing.T) {
		type holder struct {
			Arr []int `json:"arr"`
		}
		doc := &holder{Arr: []int{1, 2}}
		p, err := New("/arr/-")
		require.NoError(t, err)

		_, err = p.Set(doc, 3)
		require.NoError(t, err)

		assert.Equal(t, []int{1, 2, 3}, doc.Arr)
	})

	t.Run("append into deeply nested slice", func(t *testing.T) {
		doc := map[string]any{
			"outer": []any{
				map[string]any{"inner": []any{"a"}},
			},
		}
		p, err := New("/outer/0/inner/-")
		require.NoError(t, err)

		_, err = p.Set(doc, "b")
		require.NoError(t, err)

		outer, ok := doc["outer"].([]any)
		require.True(t, ok)
		first, ok := outer[0].(map[string]any)
		require.True(t, ok)
		inner, ok := first["inner"].([]any)
		require.True(t, ok)
		assert.Equal(t, []any{"a", "b"}, inner)
	})

	t.Run("SetForToken with dash appends", func(t *testing.T) {
		out, err := SetForToken([]int{1, 2}, "-", 3)
		require.NoError(t, err)

		outSlice, ok := out.([]int)
		require.True(t, ok)
		assert.Equal(t, []int{1, 2, 3}, outSlice)
	})
}

func TestDashToken_SetErrors(t *testing.T) {
	t.Parallel()

	t.Run("intermediate dash is rejected", func(t *testing.T) {
		doc := map[string]any{"arr": []any{1, 2}}
		p, err := New("/arr/-/x")
		require.NoError(t, err)

		_, err = p.Set(doc, 3)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDashToken)
	})

	t.Run("append with wrong element type fails", func(t *testing.T) {
		doc := map[string]any{"arr": []int{1, 2}}
		p, err := New("/arr/-")
		require.NoError(t, err)

		_, err = p.Set(doc, "not-an-int")
		require.Error(t, err)
	})
}

// dashSetter captures whatever token JSONSet receives, including "-".
type dashSetter struct {
	key   string
	value any
}

func (d *dashSetter) JSONSet(key string, value any) error {
	d.key = key
	d.value = value
	return nil
}

func TestDashToken_JSONSetableReceivesRawDash(t *testing.T) {
	t.Parallel()

	// When the terminal parent implements JSONSetable, the dash token is
	// passed through verbatim. Semantics are the user type's responsibility.
	ds := &dashSetter{}
	p, err := New("/-")
	require.NoError(t, err)

	_, err = p.Set(ds, 42)
	require.NoError(t, err)
	assert.Equal(t, "-", ds.key)
	assert.Equal(t, 42, ds.value)
}

func TestDashToken_RoundTrip(t *testing.T) {
	t.Parallel()

	p, err := New("/a/-")
	require.NoError(t, err)
	assert.Equal(t, "/a/-", p.String())
	assert.Equal(t, []string{"a", "-"}, p.DecodedTokens())
}

func TestDashToken_WrappedErrors(t *testing.T) {
	t.Parallel()

	// Ensure errors.Is works through both wraps.
	p, _ := New("/arr/-")
	doc := map[string]any{"arr": []any{}}

	_, _, err := p.Get(doc)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDashToken))
	assert.True(t, errors.Is(err, ErrPointer))
}
