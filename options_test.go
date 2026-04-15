// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"reflect"
	"sync"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// stubNameProvider is a NameProvider that maps JSON names to Go field names
// via a fixed dictionary. It lets tests observe which provider was used by
// the resolver without relying on the default reflection-based behavior.
type stubNameProvider struct {
	mu       sync.Mutex
	mapping  map[string]string
	lookups  []string
	forTypes []string
}

func (s *stubNameProvider) GetGoName(_ any, name string) (string, bool) {
	s.record(name, false)
	goName, ok := s.mapping[name]
	return goName, ok
}

func (s *stubNameProvider) GetGoNameForType(_ reflect.Type, name string) (string, bool) {
	s.record(name, true)
	goName, ok := s.mapping[name]
	return goName, ok
}

func (s *stubNameProvider) record(name string, forType bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if forType {
		s.forTypes = append(s.forTypes, name)
		return
	}
	s.lookups = append(s.lookups, name)
}

type optionStruct struct {
	// intentional: the JSON name "renamed" is deliberately not a valid
	// struct tag so that only a custom provider can resolve it.
	Field string
}

func TestWithNameProvider_overridesDefault(t *testing.T) {
	t.Parallel()

	stub := &stubNameProvider{mapping: map[string]string{"renamed": "Field"}}

	doc := optionStruct{Field: "hello"}
	p, err := New("/renamed")
	require.NoError(t, err)

	v, _, err := p.Get(doc, WithNameProvider(stub))
	require.NoError(t, err)
	assert.Equal(t, "hello", v)

	stub.mu.Lock()
	defer stub.mu.Unlock()
	assert.Contains(t, stub.forTypes, "renamed", "custom provider must be consulted")
}

func TestWithNameProvider_setRoutesThroughProvider(t *testing.T) {
	t.Parallel()

	stub := &stubNameProvider{mapping: map[string]string{"renamed": "Field"}}

	doc := &optionStruct{Field: "before"}
	p, err := New("/renamed")
	require.NoError(t, err)

	_, err = p.Set(doc, "after", WithNameProvider(stub))
	require.NoError(t, err)
	assert.Equal(t, "after", doc.Field)
}

func TestSetDefaultNameProvider_roundTrip(t *testing.T) {
	// Not Parallel: mutates package state.
	original := DefaultNameProvider()
	t.Cleanup(func() { SetDefaultNameProvider(original) })

	stub := &stubNameProvider{mapping: map[string]string{"renamed": "Field"}}
	SetDefaultNameProvider(stub)

	assert.Same(t, stub, DefaultNameProvider())

	doc := optionStruct{Field: "hello"}
	p, err := New("/renamed")
	require.NoError(t, err)

	v, _, err := p.Get(doc)
	require.NoError(t, err)
	assert.Equal(t, "hello", v)
}

func TestSetDefaultNameProvider_nilIgnored(t *testing.T) {
	// Not Parallel: mutates package state.
	original := DefaultNameProvider()
	t.Cleanup(func() { SetDefaultNameProvider(original) })

	SetDefaultNameProvider(nil)
	assert.Same(t, original, DefaultNameProvider(), "nil must be a no-op")
}

func TestUseGoNameProvider_resolvesUntaggedFields(t *testing.T) {
	// Not Parallel: mutates package state.
	original := DefaultNameProvider()
	t.Cleanup(func() { SetDefaultNameProvider(original) })

	// optionStruct.Field has no json tag; the default provider can't resolve it,
	// but the Go-name provider follows encoding/json conventions and can.
	doc := optionStruct{Field: "hello"}

	p, err := New("/Field")
	require.NoError(t, err)

	_, _, err = p.Get(doc)
	require.Error(t, err, "default provider should not resolve untagged fields")

	UseGoNameProvider()

	v, _, err := p.Get(doc)
	require.NoError(t, err)
	assert.Equal(t, "hello", v)
}

func TestDefaultNameProvider_reachesGetForToken(t *testing.T) {
	// Not Parallel: mutates package state.
	original := DefaultNameProvider()
	t.Cleanup(func() { SetDefaultNameProvider(original) })

	stub := &stubNameProvider{mapping: map[string]string{"renamed": "Field"}}
	SetDefaultNameProvider(stub)

	doc := optionStruct{Field: "hello"}
	v, _, err := GetForToken(doc, "renamed")
	require.NoError(t, err)
	assert.Equal(t, "hello", v)
}
