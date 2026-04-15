package timex

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if DateTime != "2006-01-02 15:04:05" {
		t.Errorf("DateTime constant mismatch")
	}
	if Date != "2006-01-02" {
		t.Errorf("Date constant mismatch")
	}
	if Time != "15:04:05" {
		t.Errorf("Time constant mismatch")
	}
}
