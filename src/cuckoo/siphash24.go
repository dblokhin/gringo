// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

// siphash adopted RFC siphash24 for Cuckoo
func siphash24(v []uint64, nonce uint64) uint64 {
	v0 := v[0]
	v1 := v[1]
	v2 := v[2]
	v3 := v[3] ^ nonce

	round := func() {
		v0 += v1
		v1 = v1<<13 | v1>>(64-13)
		v1 ^= v0
		v0 = v0<<32 | v0>>(64-32)

		v2 += v3
		v3 = v3<<16 | v3>>(64-16)
		v3 ^= v2

		v0 += v3
		v3 = v3<<21 | v3>>(64-21)
		v3 ^= v0

		v2 += v1
		v1 = v1<<17 | v1>>(64-17)
		v1 ^= v2
		v2 = v2<<32 | v2>>(64-32)
	}

	round()
	round()

	v0 ^= nonce;
	v2 ^= 0xff;

	round()
	round()
	round()
	round()

	return v0 ^ v1 ^ v2 ^ v3;
}