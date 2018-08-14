package util

import "testing"

func TestGetCurrentDirectory(t *testing.T) {
	s := GetCurrentDirectory("test")
	t.Log(s)
}