package adapter

import "testing"

func TestFletcher16(t *testing.T) {
	// Standard Fletcher-16 test vector.
	if got := Fletcher16([]byte("abcde")); got != 0xC8F0 {
		t.Fatalf("Fletcher16(abcde) = %04X, want C8F0", got)
	}
}
