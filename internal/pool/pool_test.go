package pool

import "testing"

// testStruct is a simple test type used to verify Pool behavior.
type testStruct struct {
	Value int
}

// Reset resets the internal state of testStruct.
func (t *testStruct) Reset() {
	t.Value = 0
}

// TestPool_GetPut verifies that:
// 1. objects can be retrieved from the pool
// 2. Put() calls Reset()
// 3. the object is returned in a clean (zero) state
func TestPool_GetPut(t *testing.T) {
	p := New(func() *testStruct {
		return &testStruct{}
	})

	// Get object from pool and modify it
	obj := p.Get()
	obj.Value = 42

	// Return object to pool (should trigger Reset)
	p.Put(obj)

	// Get object again
	obj2 := p.Get()

	// Ensure state was reset
	if obj2.Value != 0 {
		t.Fatalf("expected 0, got %d", obj2.Value)
	}
}

// TestPool_ReusesObject attempts to verify that Pool reuses objects.
// Note: sync.Pool does not guarantee reuse, so this test is best-effort.
func TestPool_ReusesObject(t *testing.T) {
	p := New(func() *testStruct {
		return &testStruct{}
	})

	obj1 := p.Get()
	p.Put(obj1)

	obj2 := p.Get()

	if obj1 != obj2 {
		t.Fatal("expected same object instance (pool reuse)")
	}
}
