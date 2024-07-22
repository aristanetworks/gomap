// Modifications copyright (c) Arista Networks, Inc. 2024
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.23

package gomap

import (
	"hash/maphash"
	"maps"
	"testing"
)

func TestRangeFuncs(t *testing.T) {
	m := New(
		func(a, b string) bool { return a == b },
		maphash.String,
		KeyElem[string, string]{"Avenue", "AVE"},
		KeyElem[string, string]{"Street", "ST"},
		KeyElem[string, string]{"Court", "CT"},
	)

	t.Run("All", func(t *testing.T) {
		exp := map[string]string{
			"Avenue": "AVE",
			"Street": "ST",
			"Court":  "CT",
		}
		// ensure break works
		for k, v := range m.All() {
			if exp[k] != v {
				t.Errorf("k=%q exp=%q got=%q", k, exp[k], v)
			}
			break
		}

		got := make(map[string]string)
		for k, v := range m.All() {
			got[k] = v
		}
		if !maps.Equal(exp, got) {
			t.Errorf("expected: %v got: %v", exp, got)
		}
	})

	t.Run("Keys", func(t *testing.T) {
		exp := map[string]struct{}{
			"Avenue": struct{}{},
			"Street": struct{}{},
			"Court":  struct{}{},
		}
		// ensure break works
		for k := range m.Keys() {
			if _, ok := exp[k]; !ok {
				t.Errorf("k=%q not found", k)
			}
			break
		}

		got := make(map[string]struct{})
		for k := range m.Keys() {
			got[k] = struct{}{}
		}
		if !maps.Equal(exp, got) {
			t.Errorf("expected: %v got: %v", exp, got)
		}
	})

	t.Run("Values", func(t *testing.T) {
		exp := map[string]struct{}{
			"AVE": struct{}{},
			"ST":  struct{}{},
			"CT":  struct{}{},
		}
		// ensure break works
		for k := range m.Values() {
			if _, ok := exp[k]; !ok {
				t.Errorf("k=%q not found", k)
			}
			break
		}

		got := make(map[string]struct{})
		for k := range m.Values() {
			got[k] = struct{}{}
		}
		if !maps.Equal(exp, got) {
			t.Errorf("expected: %v got: %v", exp, got)
		}
	})
}
