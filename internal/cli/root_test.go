package cli

import (
	"reflect"
	"testing"
)

func TestNormalizeRootArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "direct prompt", in: []string{"hello"}, want: []string{"run", "hello"}},
		{name: "root flags", in: []string{"--cwd", ".", "hello"}, want: []string{"run", "--cwd", ".", "hello"}},
		{name: "provider shorthand", in: []string{"claude", "--cwd", ".", "review"}, want: []string{"run", "-p", "claude", "--cwd", ".", "review"}},
		{name: "known command", in: []string{"agents", "--json"}, want: []string{"agents", "--json"}},
		{name: "run command", in: []string{"run", "hello"}, want: []string{"run", "hello"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeRootArgs(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("normalizeRootArgs(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
