package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHashSlice(t *testing.T) {
	svc := NewService()

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "Slice of int",
			data: []int{1, 2, 3, 4, 5},
		},
		{
			name: "Struct",
			data: []Person{{Name: "bar", Age: 19}, {Name: "foo", Age: 12}},
		},
		{
			name: "Nil value",
			data: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := svc.HashSlice(tc.data)
			if err != nil {
				t.Fatalf("HashSlice(%v) returned error: %v", tc.data, err)
			}

			jsonData, err := json.Marshal(tc.data)
			if err != nil {
				t.Fatalf("json.Marshal(%v) returned error: %v", tc.data, err)
			}

			expected := fmt.Sprintf("%x", sha256.Sum256(jsonData))
			if got != expected {
				t.Errorf("HashSlice(%v) = %s, expected %s", tc.data, got, expected)
			}
		})
	}
}
