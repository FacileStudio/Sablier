package env

import (
	"slices"
	"testing"
)

func TestWithNativeOriginAddsTheAppOnce(t *testing.T) {
	got := withNativeOrigin([]string{"https://sablier.example"}, DefaultMobileRedirect)
	want := []string{"https://sablier.example", NativeOrigin}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := withNativeOrigin(want, DefaultMobileRedirect); !slices.Equal(got, want) {
		t.Fatalf("an origin already listed must not repeat, got %v", got)
	}
	if got := withNativeOrigin([]string{"https://sablier.example"}, ""); len(got) != 1 {
		t.Fatalf("mobile login off must leave the list alone, got %v", got)
	}
}
