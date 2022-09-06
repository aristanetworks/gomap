// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomap_test

import (
	"bytes"
	"fmt"
	"hash/maphash"

	"github.com/aristanetworks/gomap"
)

func ExampleMap_Iter() {
	m := gomap.New(
		func(a, b string) bool { return a == b },
		maphash.String,
		gomap.KeyElem[string, string]{"Avenue", "AVE"},
		gomap.KeyElem[string, string]{"Street", "ST"},
		gomap.KeyElem[string, string]{"Court", "CT"},
	)

	for i := m.Iter(); i.Next(); {
		fmt.Printf("The abbreviation for %q is %q", i.Key(), i.Elem())
	}
}

func ExampleNew() {
	// Some examples of different maps:

	// Map that uses []byte as the key type
	byteSliceMap := gomap.New[[]byte, int](bytes.Equal, maphash.Bytes)
	byteSliceMap.Set([]byte("hello"), 42)

	// Map that uses map[string]struct{} as the key type
	stringSetEqual := func(a, b map[string]struct{}) bool {
		if len(a) != len(b) {
			return false
		}
		for k := range a {
			_, ok := b[k]
			if !ok {
				return false
			}
		}
		return true
	}
	stringSetHash := func(seed maphash.Seed, ss map[string]struct{}) uint64 {
		var sum uint64
		for s := range ss {
			// combine hashes with addition so that iteration order does not matter
			sum += maphash.String(seed, s)
		}
		return sum
	}
	stringSetMap := gomap.New[map[string]struct{}, int](stringSetEqual, stringSetHash)
	stringSetMap.Set(map[string]struct{}{"foo": {}, "bar": {}}, 42)
}
