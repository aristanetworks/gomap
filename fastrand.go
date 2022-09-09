// Modifications copyright (c) Arista Networks, Inc. 2022
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.19

package gomap

import (
	_ "unsafe"
)

//go:linkname fastrand64 runtime.fastrand64
func fastrand64() uint64
