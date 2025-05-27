package require

import (
	"testing"
)

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Logf("Error is not nil, but should not have an error")
		t.Fail()
	}
}
