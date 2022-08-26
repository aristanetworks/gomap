// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomap_test

import (
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
