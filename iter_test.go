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
	"sync"
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

func TestGetAllRace(t *testing.T) {
	m := NewHint[int, int](100, func(a int, b int) bool { return a == b }, intHash)
	for i := 0; i < 100; i++ {
		m.Set(i, i)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			v, ok := m.Get(i)
			if !ok || v != i {
				t.Errorf("expected: %d got: %d, %t", i, v, ok)
			}
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			v, ok := m.Get(i)
			if !ok || v != i {
				t.Errorf("expected: %d got: %d, %t", i, v, ok)
			}
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
	outer:
		for i := 0; i < 100; i++ {
			for range m.All() {
				continue outer
			}
			t.Error("Should have iterated at least once, but didn't.")
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
	outer:
		for i := 0; i < 100; i++ {
			for range m.All() {
				continue outer
			}
			t.Error("Should have iterated at least once, but didn't.")
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestAll(t *testing.T) {
	m := New[uint64, uint64](
		func(a, b uint64) bool { return a == b },
		badIntHash,
	)
	expected := make(map[uint64]uint64, 9)
	for i := uint64(0); i < 9; i++ {
		expected[i] = i
		m.Set(i, i)
	}
	for k, v := range m.All() {
		e, ok := expected[k]
		if !ok {
			t.Errorf("unexpected value in m: [%d: %d]", k, v)
			continue
		}
		if e != v {
			t.Errorf("wrong value for key %d. Expected: %d Got: %d", k, e, v)
			continue
		}
		delete(expected, k)
	}
	if len(expected) > 0 {
		t.Errorf("Values not found in m: %v", expected)
	}
}

func TestAllDuringResize(t *testing.T) {
	m := New[uint64, uint64](
		func(a, b uint64) bool { return a == b },
		badIntHash,
	)

	// insert numbers that initially hash to the same bucket, but will
	// be split into different buckets on resize. Evens will end up in
	// bucket[0], odds end up in bucket[1] thanks to badIntHash.
	initial := map[uint64]uint64{0: 0, 1: 1, 2: 2, 3: 3}
	for k, e := range initial {
		m.Set(k, e)
	}
	additional := map[uint64]uint64{100: 100, 101: 101, 102: 102, 103: 103, 104: 104}

	first := true
	// start the iter
	for k, v := range m.All() {
		if first {
			first = false
			// Add some additional data to cause a resize
			for k, e := range additional {
				m.Set(k, e)
			}
			// Remove 1 value that in each of the initial and split
			// buckets that we haven't seen yet
			if k == 0 {
				m.Delete(2)
				delete(initial, 2)
			} else {
				m.Delete(0)
				delete(initial, 0)
			}
			if k == 1 {
				m.Delete(3)
				delete(initial, 3)
			} else {
				m.Delete(1)
				delete(initial, 1)
			}
		}
		if k != v {
			t.Errorf("expected key == elem, but got: %d != %d", k, v)
			t.Error(m.debugString())
		}
		if _, ok := initial[k]; ok {
			delete(initial, v)
			continue
		}
		if _, ok := additional[k]; ok {
			t.Logf("Saw key from additional: %d", k)
			continue
		}
		t.Errorf("Unexpected value from iter: %d", k)
	}
	for k := range initial {
		t.Errorf("iter missing key: %d", k)
	}
}

func TestAllDuringGrow(t *testing.T) {
	m := New[uint64, uint64](
		func(a, b uint64) bool { return a == b },
		badIntHash,
	)

	// Insert exactly 27 numbers so we end up in the middle of a grow.
	expected := make(map[uint64]uint64, 27)
	for i := uint64(0); i < 27; i++ {
		expected[i] = i
		m.Set(i, i)
	}

	for k, v := range m.All() {
		t.Logf("Key: %d", k)
		if k != v {
			t.Errorf("expected key == elem, but got: %d != %d", k, v)
			t.Error(m.debugString())
		}

		if _, ok := expected[k]; ok {
			delete(expected, v)
			continue
		}
		t.Errorf("Unexpected value from iter: %d", k)
	}
	for k := range expected {
		t.Errorf("iter missing key: %d", k)
	}
}

func BenchmarkAll(b *testing.B) {
	m := New[string, int](
		func(a, b string) bool { return a == b },
		maphash.String,
		KeyElem[string, int]{"one", 1},
		KeyElem[string, int]{"two", 2},
		KeyElem[string, int]{"three", 3},
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range m.All() {
		}
	}
}
