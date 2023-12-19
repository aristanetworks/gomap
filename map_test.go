// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomap

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"strings"
	"sync"
	"testing"

	"golang.org/x/exp/slices"
)

func (m *Map[K, E]) debugString() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "count: %d, buckets: %d, overflows: %d growing: %t\n",
		m.count, len(m.buckets), m.noverflow, m.growing())

	for i, b := range m.buckets {
		fmt.Fprintf(&buf, "bucket: %d\n", i)
		b := &b
		for {
			for i := uintptr(0); i < bucketCnt; i++ {
				seen := map[uint8]struct{}{}
				switch b.tophash[i] {
				case emptyRest:
					buf.WriteString("  emptyRest\n")
				case emptyOne:
					buf.WriteString("  emptyOne\n")
				case evacuatedX:
					buf.WriteString("  evacuatedX?\n")
				case evacuatedY:
					buf.WriteString("  evacuatedY?\n")
				case evacuatedEmpty:
					buf.WriteString("  evacuatedEmpty?\n")
				default:
					var s string
					if _, ok := seen[b.tophash[i]]; ok {
						s = " duplicate"
					} else {
						seen[b.tophash[i]] = struct{}{}
					}
					fmt.Fprintf(&buf, "  0x%02x"+s+"\n", b.tophash[i])
				}
			}
			if b.overflow == nil {
				break
			}
			buf.WriteString("overflow->\n")
			b = b.overflow
		}
	}

	return buf.String()
}

func intHash(seed maphash.Seed, a int) uint64 {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(a))
	return maphash.Bytes(seed, buf[:])
}

func TestSetGetDelete(t *testing.T) {
	const count = 1000
	t.Run("nohint", func(t *testing.T) {
		m := New[int, int](func(a int, b int) bool { return a == b }, intHash)
		t.Logf("Buckets: %d Unused-overflow: %d", len(m.buckets), cap(m.buckets)-len(m.buckets))
		for i := 0; i < count; i++ {
			m.Set(i, i)
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}
			if m.Len() != i+1 {
				t.Errorf("expected len: %d got: %d", i+1, m.Len())
			}
		}
		t.Logf("Buckets: %d Unused-overflow: %d", len(m.buckets), cap(m.buckets)-len(m.buckets))
		t.Log("Overflow:", m.noverflow)
		for i := 0; i < count; i++ {
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}
			if m.Len() != count {
				t.Errorf("expected len: %d got: %d", count, m.Len())
			}

		}
		for i := 0; i < count; i++ {
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}

			m.Delete(i)

			if v, ok := m.Get(i); ok {
				t.Errorf("found %d: %d, but it should have been deleted", i, v)
			}
			if m.Len() != count-i-1 {
				t.Errorf("expected len: %d got: %d", count, m.Len())
			}
		}
	})
	t.Run("hint", func(t *testing.T) {
		m := NewHint[int, int](count, func(a int, b int) bool { return a == b }, intHash)
		t.Logf("Buckets: %d Unused-overflow: %d", len(m.buckets), cap(m.buckets)-len(m.buckets))
		for i := 0; i < count; i++ {
			m.Set(i, i)
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}
			if m.Len() != i+1 {
				t.Errorf("expected len: %d got: %d", i+1, m.Len())
			}
		}
		t.Logf("Buckets: %d Unused-overflow: %d", len(m.buckets), cap(m.buckets)-len(m.buckets))
		t.Log("Overflow:", m.noverflow)
		for i := 0; i < count; i++ {
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}
			if m.Len() != count {
				t.Errorf("expected len: %d got: %d", count, m.Len())
			}

		}
		for i := 0; i < count; i++ {
			if v, ok := m.Get(i); !ok {
				t.Errorf("got not ok for %d", i)
			} else if v != i {
				t.Errorf("unexpected value for %d: %d", i, v)
			}

			m.Delete(i)

			if v, ok := m.Get(i); ok {
				t.Errorf("found %d: %d, but it should have been deleted", i, v)
			}
			if m.Len() != count-i-1 {
				t.Errorf("expected len: %d got: %d", count, m.Len())
			}
		}
	})
}

func BenchmarkGrow(b *testing.B) {
	b.Run("hint", func(b *testing.B) {
		b.ReportAllocs()
		m := NewHint[int, int](b.N, func(a int, b int) bool { return a == b }, intHash)
		for i := 0; i < b.N; i++ {
			m.Set(i, i)
		}
	})
	b.Run("nohint", func(b *testing.B) {
		b.ReportAllocs()
		m := New[int, int](func(a int, b int) bool { return a == b }, intHash)
		for i := 0; i < b.N; i++ {
			m.Set(i, i)
		}
	})

	b.Run("std:hint", func(b *testing.B) {
		b.ReportAllocs()
		m := make(map[int]int, b.N)
		for i := 0; i < b.N; i++ {
			m[i] = i
		}
	})
	b.Run("std:nohint", func(b *testing.B) {
		b.ReportAllocs()
		m := map[int]int{}
		for i := 0; i < b.N; i++ {
			m[i] = i
		}
	})
}

func TestGetIterateRace(t *testing.T) {
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
		for i := 0; i < 100; i++ {
			iter := m.Iter()
			if !iter.Next() {
				t.Error("unexpected end of iter")
			}
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			iter := m.Iter()
			if !iter.Next() {
				t.Error("unexpected end of iter")
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

// badIntHash is a bad hash function that gives simple deterministic
// hash to give control over which bucket a key lands in.
func badIntHash(seed maphash.Seed, a uint64) uint64 {
	return uint64(a)
}

func TestIter(t *testing.T) {
	m := New[uint64, uint64](
		func(a, b uint64) bool { return a == b },
		badIntHash,
	)
	expected := make(map[uint64]uint64, 9)
	for i := uint64(0); i < 9; i++ {
		expected[i] = i
		m.Set(i, i)
	}
	for i := m.Iter(); i.Next(); {
		e, ok := expected[i.Key()]
		if !ok {
			t.Errorf("unexpected value in m: [%d: %d]", i.Key(), i.Elem())
			continue
		}
		if e != i.Elem() {
			t.Errorf("wrong value for key %d. Expected: %d Got: %d", i.Key(), e, i.Elem())
			continue
		}
		delete(expected, i.Key())
	}
	if len(expected) > 0 {
		t.Errorf("Values not found in m: %v", expected)
	}
}

func TestClear(t *testing.T) {
	m := New(
		func(a, b string) bool { return a == b },
		maphash.String,
		KeyElem[string, string]{"a", "a"},
		KeyElem[string, string]{"b", "b"},
		KeyElem[string, string]{"c", "c"},
		KeyElem[string, string]{"d", "d"},
	)
	if m.Len() != 4 {
		t.Fatalf("Unexpected size after New (%d): %s", m.Len(), m.debugString())
	}
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("expected empty map: %s", m.debugString())
	}
	for i := m.Iter(); i.Next(); {
		t.Errorf("unexpected entry in map: [%s: %s]", i.Key(), i.Elem())
	}
}

func TestIterResize(t *testing.T) {
	m := New[uint64, uint64](
		func(a, b uint64) bool { return a == b },
		badIntHash,
	)

	// insert numbers that initially hash to the same bucket, but will
	// be split into different buckets on resize.
	initial := map[uint64]uint64{0: 0, 1: 1, 2: 2, 3: 3}
	for k, e := range initial {
		m.Set(k, e)
	}
	// create the iter
	i := m.Iter()
	// Add some additional data to cause a resize
	additional := map[uint64]uint64{100: 100, 101: 101, 102: 102, 103: 103, 104: 104}
	for k, e := range additional {
		m.Set(k, e)
	}

	// Remove 1 value that in each of the initial and split buckets
	m.Delete(0)
	delete(initial, 0)
	m.Delete(1)
	delete(initial, 1)
	for i.Next() {
		if i.Key() != i.Elem() {
			t.Errorf("expected key == elem, but got: %d != %d", i.Key(), i.Elem())
			t.Error(m.debugString())
		}
		if _, ok := initial[i.Key()]; ok {
			delete(initial, i.Key())
			continue
		}
		if _, ok := additional[i.Key()]; ok {
			t.Logf("Saw key from additional: %d", i.Key())
			continue
		}
		t.Errorf("Unexpected value from iter: %d", i.Key())
	}
	for k := range initial {
		t.Errorf("iter missing key: %d", k)
	}
}

func TestIterDuringGrow(t *testing.T) {
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
	// create the iter while Map is growing
	i := m.Iter()

	for i.Next() {
		t.Logf("Key: %d", i.Key())
		if i.Key() != i.Elem() {
			t.Errorf("expected key == elem, but got: %d != %d", i.Key(), i.Elem())
			t.Error(m.debugString())
		}

		if _, ok := expected[i.Key()]; ok {
			delete(expected, i.Key())
			continue
		}
		t.Errorf("Unexpected value from iter: %d", i.Key())
	}
	for k := range expected {
		t.Errorf("iter missing key: %d", k)
	}
}

func TestUpdate(t *testing.T) {
	m := New[int, []int](
		func(a, b int) bool { return a == b },
		intHash)
	for key := 0; key < 10; key++ {
		var expected []int
		for i := 0; i < 3; i++ {
			m.Update(key, func(cur []int) []int { return append(cur, 1) })
			expected = append(expected, 1)
			got, ok := m.Get(key)
			if !ok {
				t.Errorf("m missing key: %v", key)
			} else if !slices.Equal(got, expected) {
				t.Errorf("Got: %v Expected: %v", got, expected)
			}
		}
	}
}

func BenchmarkIter(b *testing.B) {
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
		for it := m.Iter(); it.Next(); {
		}
	}
}

func BenchmarkRand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fastrand64()
	}
}
