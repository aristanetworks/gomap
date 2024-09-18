// Modifications copyright (c) Arista Networks, Inc. 2024
// Underlying
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.23

package gomap

import "iter"

// All returns an iterator over key-value pairs from m.
func (m *Map[K, E]) All() iter.Seq2[K, E] {
	return m.iterFunc
}

func (m *Map[K, E]) iterFunc(yield func(k K, e E) bool) {
	if m == nil || m.count == 0 {
		return
	}
	// Remember we have an iterator.
	// Can run concurrently with another m.Iter().
	atomicOr(&m.flags, iterator|oldIterator)

	var (
		r           = rand64()
		buckets     = m.buckets
		startBucket = int(r & m.bucketMask())
		nextBucket  = startBucket
		b           *bucket[K, E]
		checkBucket int
		offset      = uint8(r >> (64 - bucketCntBits))
		wrapped     bool
	)

	for {
		if b == nil {
			if nextBucket == startBucket && wrapped {
				return
			}

			if m.growing() && len(buckets) == len(m.buckets) {
				// Iterator was started in the middle of a grow, and the grow isn't done yet.
				// If the bucket we're looking at hasn't been filled in yet (i.e. the old
				// bucket hasn't been evacuated) then we need to iterate through the old
				// bucket and only return the ones that will be migrated to this bucket.
				oldbucket := uint64(nextBucket) & m.oldbucketmask()
				b = &(*m.oldbuckets)[oldbucket]
				if !evacuated(b) {
					checkBucket = nextBucket
				} else {
					b = &buckets[nextBucket]
					checkBucket = noCheck
				}
			} else {
				b = &buckets[nextBucket]
				checkBucket = noCheck
			}
			nextBucket++
			if nextBucket == len(buckets) {
				nextBucket = 0
				wrapped = true
			}
		}

		for i := uint8(0); i < bucketCnt; i++ {
			offi := (i + offset) & (bucketCnt - 1)
			if isEmpty(b.tophash[offi]) || b.tophash[offi] == evacuatedEmpty {
				// TODO: emptyRest is hard to use here, as we start iterating
				// in the middle of a bucket. It's feasible, just tricky.
				continue
			}
			k := b.keys[offi]
			if checkBucket != noCheck && !m.sameSizeGrow() {
				// Special case: iterator was started during a grow to a larger size
				// and the grow is not done yet. We're working on a bucket whose
				// oldbucket has not been evacuated yet. Or at least, it wasn't
				// evacuated when we started the bucket. So we're iterating
				// through the oldbucket, skipping any keys that will go
				// to the other new bucket (each oldbucket expands to two
				// buckets during a grow).
				// If the item in the oldbucket is not destined for
				// the current new bucket in the iteration, skip it.
				hash := m.hash(m.seed, k)
				if int(hash&m.bucketMask()) != checkBucket {
					continue
				}
			}
			if b.tophash[offi] != evacuatedX && b.tophash[offi] != evacuatedY {
				// This is the golden data, we can return it.
				if !yield(k, b.elems[offi]) {
					return
				}
			} else {
				// The hash table has grown since the iterator was started.
				// The golden data for this key is now somewhere else.
				// Check the current hash table for the data.
				// This code handles the case where the key
				// has been deleted, updated, or deleted and reinserted.
				// NOTE: we need to regrab the key as it has potentially been
				// updated to an equal() but not identical key (e.g. +0.0 vs -0.0).
				rk, re := m.mapaccessK(k)
				if rk == nil {
					continue // key has been deleted
				}
				if !yield(*rk, *re) {
					return
				}
			}
		}
		b = b.overflow
	}
}

// Keys returns an iterator over keys in m.
func (m *Map[K, E]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over values in m.
func (m *Map[K, E]) Values() iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}
