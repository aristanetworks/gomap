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
