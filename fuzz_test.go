// SPDX-FileCopyrightText: Copyright (c) 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package jsonpointer

import (
	"iter"
	"slices"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

func FuzzParse(f *testing.F) {
	cumulated := make([]string, 0, 100)
	for generator := range generators() {
		f.Add(generator)

		cumulated = append(cumulated, generator)
		f.Add(strings.Join(cumulated, ""))
	}

	f.Fuzz(func(t *testing.T, input string) {
		require.NotPanics(t, func() {
			_, _ = New(input)
		})
	})
}

func generators() iter.Seq[string] {
	return slices.Values([]string{
		`a`,
		``, `/`, `/`, `/a~1b`, `/a~1b`, `/c%d`, `/e^f`, `/g|h`, `/i\j`, `/k"l`, `/ `, `/m~0n`,
		`/foo`, `/0`,
	})
}
