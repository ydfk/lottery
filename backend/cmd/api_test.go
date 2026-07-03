package main

import "testing"

func TestIsNoStoreFrontendPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/", want: true},
		{path: "/index.html", want: true},
		{path: "/app-version.json", want: true},
		{path: "/dashboard", want: true},
		{path: "/api/lotteries", want: false},
		{path: "/uploads/tickets/example.jpg", want: false},
		{path: "/assets/index-abc.js", want: false},
	}

	for _, tt := range tests {
		if got := isNoStoreFrontendPath(tt.path); got != tt.want {
			t.Fatalf("isNoStoreFrontendPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
