package ratelimit

import "testing"

func TestSizes(t *testing.T) {
	if B != 1 {
		t.Error("Wrong size for byte")
	}

	if KB != 1024 {
		t.Error("Wrong size for kilobyte")
	}
}
