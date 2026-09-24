package randutil

import (
	"sort"
	"testing"
)

func TestIntn_StaysInRange(t *testing.T) {
	rng := New()

	if got := rng.Intn(1); got != 0 {
		t.Fatalf("Intn(1) = %d, want 0", got)
	}
	if got := rng.Intn(0); got != 0 {
		t.Fatalf("Intn(0) = %d, want 0", got)
	}
	if got := rng.Intn(-5); got != 0 {
		t.Fatalf("Intn(-5) = %d, want 0", got)
	}

	const n = 7
	seen := map[int]bool{}
	for i := 0; i < 500; i++ {
		got := rng.Intn(n)
		if got < 0 || got >= n {
			t.Fatalf("Intn(%d) = %d, out of range", n, got)
		}
		seen[got] = true
	}
	if len(seen) < 2 {
		t.Fatalf("Intn(%d) produced only %d distinct values; source is not random", n, len(seen))
	}
}

func TestInt63n_StaysInRange(t *testing.T) {
	rng := New()

	if got := rng.Int63n(0); got != 0 {
		t.Fatalf("Int63n(0) = %d, want 0", got)
	}
	if got := rng.Int63n(-1); got != 0 {
		t.Fatalf("Int63n(-1) = %d, want 0", got)
	}

	const n = int64(1000)
	for i := 0; i < 200; i++ {
		got := rng.Int63n(n)
		if got < 0 || got >= n {
			t.Fatalf("Int63n(%d) = %d, out of range", n, got)
		}
	}
}

func TestPerm_IsAPermutation(t *testing.T) {
	rng := New()
	const n = 10

	perm := rng.Perm(n)
	if len(perm) != n {
		t.Fatalf("Perm(%d) returned %d elements", n, len(perm))
	}

	sorted := append([]int(nil), perm...)
	sort.Ints(sorted)
	for i, v := range sorted {
		if v != i {
			t.Fatalf("Perm(%d) = %v is not a permutation of [0,%d)", n, perm, n)
		}
	}
}

func TestShuffle_KeepsElementsAndHandlesSmallN(t *testing.T) {
	rng := New()

	// n < 2 must not panic and must not call swap.
	for _, n := range []int{-1, 0, 1} {
		called := false
		rng.Shuffle(n, func(i, j int) { called = true })
		if called {
			t.Fatalf("Shuffle(%d) invoked swap", n)
		}
	}

	values := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < 100; i++ {
		rng.Shuffle(len(values), func(a, b int) { values[a], values[b] = values[b], values[a] })
		sorted := append([]int(nil), values...)
		sort.Ints(sorted)
		for idx, v := range sorted {
			if v != idx {
				t.Fatalf("Shuffle lost elements: %v", values)
			}
		}
	}
}
