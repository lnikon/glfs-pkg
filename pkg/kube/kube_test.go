package kube

import (
	"testing"
)

func TestListPods(t *testing.T) {
	want := 13
	if got := GetPodsCount(); got != want {
		t.Errorf("Got %d pods, expected %d", got, want)
	} else {
		t.Logf("Got %d pods, expected %d", got, want)
	}
}
