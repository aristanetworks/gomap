// Modifications copyright (c) Arista Networks, Inc. 2024
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.23

package gomap

import "sync/atomic"

func atomicOr(flags *uint32, or uint32) {
	atomic.OrUint32(flags, or)
}
