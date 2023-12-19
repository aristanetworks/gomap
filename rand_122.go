// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.22

package gomap

import (
	"math/rand/v2"
)

func rand64() uint64 {
	return rand.Uint64()
}
