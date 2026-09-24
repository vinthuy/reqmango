// Package randutil provides a small random source backed by crypto/rand.
//
// It exists so that non-security code (demo seeding, local QA tools) never has to
// reach for math/rand — which gosec flags as G404 — while also avoiding blanket
// nosec suppressions that would hide a real misuse elsewhere.
package randutil

import (
	"crypto/rand"
	"math/big"
)

// RNG mirrors the subset of math/rand that the seeders use.
type RNG struct{}

// New returns a random source backed by crypto/rand.
func New() *RNG { return &RNG{} }

// Intn returns a random int in [0, n).
//
// It returns 0 for n <= 0 and on the (practically impossible) failure of the
// system entropy source, so callers can keep using the result as an index.
func (r *RNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// Int63n returns a random int64 in [0, n).
func (r *RNG) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return 0
	}
	return v.Int64()
}

// Perm returns a random permutation of the integers in [0, n).
func (r *RNG) Perm(n int) []int {
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	r.Shuffle(n, func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
	return perm
}

// Shuffle randomises the order of n elements, calling swap for each exchanged
// pair. It follows the same contract as math/rand.Rand.Shuffle.
func (r *RNG) Shuffle(n int, swap func(i, j int)) {
	if n < 2 {
		return
	}
	for i := n - 1; i > 0; i-- {
		swap(i, r.Intn(i+1))
	}
}
