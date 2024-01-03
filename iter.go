// Modifications copyright (c) Arista Networks, Inc. 2024
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.rangefunc

package gomap

import "iter"

// All returns an iterator over key-value pairs from m.
func (m *Map[K, E]) All() iter.Seq2[K, E] {
	return func(yield func(K, E) bool) {
		for it := m.Iter(); it.Next(); {
			if !yield(it.Key(), it.Elem()) {
				return
			}
		}
	}
}

// Keys returns an iterator over keys in m.
func (m *Map[K, E]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for it := m.Iter(); it.Next(); {
			if !yield(it.Key()) {
				return
			}
		}
	}
}

// Values returns an iterator over values in m.
func (m *Map[K, E]) Values() iter.Seq[E] {
	return func(yield func(E) bool) {
		for it := m.Iter(); it.Next(); {
			if !yield(it.Elem()) {
				return
			}
		}
	}
}
