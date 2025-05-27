package assert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func Error(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Logf("Error is nil, but should have an error")
		t.Fail()
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Logf("Error is not nil, but should not have an error")
		t.Fail()
	}
}

func Equal(t *testing.T, expect, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expect, actual) {
		t.Logf("Expect %v, but got %v", expect, actual)
		t.Fail()
	}
}

func Equalf(t *testing.T, expect, actual interface{}, format string, args ...interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expect, actual) {
		t.Logf("Expect %v, but got %v: %s", expect, actual, fmt.Sprintf(format, args...))
		t.Fail()
	}
}

func True(t *testing.T, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(true, actual) {
		t.Logf("Expect %v, but got %v", true, actual)
		t.Fail()
	}
}

func False(t *testing.T, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(false, actual) {
		t.Logf("Expect %v, but got %v", false, actual)
		t.Fail()
	}
}

func CompareJson(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedBytes, _ := json.Marshal(expected)
	actualBytes, _ := json.Marshal(actual)

	if !bytes.Equal(expectedBytes, actualBytes) {
		t.Logf("Expect %v, but got %v", expected, actual)
		t.Fail()
	}
}
