package main

import (
	"reflect"
	"testing"
)

func TestParsePicks(t *testing.T) {
	cases := []struct {
		raw   string
		count int
		want  []int
		ok    bool
	}{
		{"1", 3, []int{1}, true},
		{"1, 3", 3, []int{1, 3}, true},
		{"1 2 3", 3, []int{1, 2, 3}, true},
		{"all", 3, []int{1, 2, 3}, true},
		{"*", 2, []int{1, 2}, true},
		{"first 3", 5, []int{1, 2, 3}, true},
		{"count:2", 5, []int{1, 2}, true},
		{"first 0", 3, nil, false},
		{"first 4", 3, nil, false},
		{"0", 3, nil, false},
		{"4", 3, nil, false},
		{"", 3, nil, false},
	}
	for _, item := range cases {
		got, err := parsePicks(item.raw, item.count)
		if item.ok && err != nil {
			t.Fatalf("parsePicks(%q) returned error: %v", item.raw, err)
		}
		if !item.ok {
			if err == nil {
				t.Fatalf("parsePicks(%q) expected an error", item.raw)
			}
			continue
		}
		if !reflect.DeepEqual(got, item.want) {
			t.Fatalf("parsePicks(%q) = %v, want %v", item.raw, got, item.want)
		}
	}
}
