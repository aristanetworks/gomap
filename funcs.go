// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomap

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

// String converts m to a string representation using K's and E's
// String functions.
func String[K fmt.Stringer, E fmt.Stringer](m *Map[K, E]) string {
	return StringFunc(m,
		func(key K) string { return key.String() },
		func(elem E) string { return elem.String() },
	)
}

type strKE struct {
	k string
	e string
}

// StringFunc converts m to a string representation with the help of
// strK and strE functions to stringify m's keys and elems.
func StringFunc[K any, E any](m *Map[K, E],
	strK func(key K) string,
	strE func(elem E) string) string {
	if m == nil || m.Len() == 0 {
		return "gomap.Map[]"
	}
	strs := make([]strKE, m.Len())
	s := 0
	i := 0
	for it := m.Iter(); it.Next(); {
		ke := &strs[i]
		ke.k = strK(it.Key())
		ke.e = strE(it.Elem())
		s += len(ke.k) + len(ke.e)
		i++
	}
	slices.SortFunc(strs, func(a, b strKE) bool { return a.k < b.k })

	var b strings.Builder
	b.Grow(len("gomap.Map[]") + // space for header and footer
		len(strs)*2 - 1 + // space for delimiters
		s) // space for keys and elems
	b.WriteString("gomap.Map[")
	for i, ke := range strs {
		if i != 0 {
			b.WriteByte(' ')
		}
		b.WriteString(ke.k)
		b.WriteByte(':')
		b.WriteString(ke.e)
	}
	b.WriteByte(']')
	return b.String()
}

// Equal returns true if the same set of keys and elems are in m1 and
// m2. Elements are compared using ==.
func Equal[K any, E comparable](m1, m2 *Map[K, E]) bool {
	if m1.Len() != m2.Len() {
		return false
	}
	for it := m1.Iter(); it.Next(); {
		e2, ok := m2.Get(it.Key())
		if !ok || it.Elem() != e2 {
			return false
		}
	}
	return true
}

// Equal returns true if the same set of keys and elems are in m1 and
// m2. Elements are compared using eq.
func EqualFunc[K, E any](m1, m2 *Map[K, E], eq func(E, E) bool) bool {
	if m1.Len() != m2.Len() {
		return false
	}
	for it := m1.Iter(); it.Next(); {
		e2, ok := m2.Get(it.Key())
		if !ok || !eq(it.Elem(), e2) {
			return false
		}
	}
	return true
}
