// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomap

import (
	"bytes"
	"hash/maphash"
	"testing"
)

func TestString(t *testing.T) {
	m := New(bytes.Equal, maphash.Bytes,
		KeyElem[[]byte, struct{}]{[]byte("abc"), struct{}{}},
		KeyElem[[]byte, struct{}]{[]byte("def"), struct{}{}},
		KeyElem[[]byte, struct{}]{[]byte("ghi"), struct{}{}},
	)
	s := m.String()
	expected := "gomap.Map[[100 101 102]:{} [103 104 105]:{} [97 98 99]:{}]"
	if expected != s {
		t.Errorf("Got: %q Expected: %q", s, expected)
	}

	s = StringFunc(m,
		func(b []byte) string { return string(b) },
		func(struct{}) string { return "✅" })
	expected = "gomap.Map[abc:✅ def:✅ ghi:✅]"
	if s != expected {
		t.Errorf("Got: %q Expected: %q", s, expected)
	}
}

func TestEqual(t *testing.T) {
	fromMap := func(m map[string]int) *Map[string, int] {
		mm := NewHint[string, int](len(m), func(a, b string) bool { return a == b }, maphash.String)
		for k, v := range m {
			mm.Set(k, v)
		}
		return mm
	}

	for _, tc := range []struct {
		a     *Map[string, int]
		b     *Map[string, int]
		equal bool
	}{{
		a: nil, b: nil, equal: true,
	}, {
		a: fromMap(map[string]int{"a": 1}), b: nil, equal: false,
	}, {
		a: nil, b: fromMap(map[string]int{"a": 1}), equal: false,
	}, {
		a:     fromMap(map[string]int{"a": 1}),
		b:     fromMap(map[string]int{"a": 1}),
		equal: true,
	}, {
		a:     fromMap(map[string]int{"a": 1, "b": 2}),
		b:     fromMap(map[string]int{"a": 1}),
		equal: false,
	}, {
		a:     fromMap(map[string]int{"a": 1, "b": 2}),
		b:     fromMap(map[string]int{"a": 1, "b": 3}),
		equal: false,
	}, {
		a:     fromMap(map[string]int{"a": 1, "b": 2}),
		b:     fromMap(map[string]int{"a": 1, "c": 3}),
		equal: false,
	}} {
		t.Run(tc.a.String()+tc.b.String(), func(t *testing.T) {
			if got := Equal(tc.a, tc.b); got != tc.equal {
				t.Errorf("Equal: %t Exp: %t", got, tc.equal)
			}
			if got := EqualFunc(tc.a, tc.b, func(x, y int) bool { return x == y }); got != tc.equal {
				t.Errorf("EqualFunc: %t Exp: %t", got, tc.equal)
			}
		})
	}
}

func BenchmarkStringFunc(b *testing.B) {
	m := New(bytes.Equal, maphash.Bytes,
		KeyElem[[]byte, struct{}]{[]byte("abc"), struct{}{}},
		KeyElem[[]byte, struct{}]{[]byte("def"), struct{}{}},
		KeyElem[[]byte, struct{}]{[]byte("ghi"), struct{}{}},
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StringFunc(m,
			func(b []byte) string { return string(b) },
			func(struct{}) string { return "x" })
	}
}
