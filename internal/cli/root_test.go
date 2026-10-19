package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRoot("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestEqListsPresets(t *testing.T) {
	out, err := run(t, "eq")
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Fields(out); len(lines) != 22 || lines[0] != "signature" {
		t.Errorf("got %q", out)
	}
}

func TestAncRejectsUnknownMode(t *testing.T) {
	if _, err := run(t, "anc", "loud"); err == nil || !strings.Contains(err.Error(), "invalid argument") {
		t.Errorf("err = %v", err)
	}
}

func TestCompletion(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		{[]string{"__complete", ""}, []string{"status", "anc", "eq", "completion"}},
		{[]string{"__complete", "anc", ""}, []string{"transport", "indoor", "outdoor"}},
		{[]string{"__complete", "eq", "bass"}, []string{"custom", "bass-booster", "bass-reducer"}},
	}
	for _, tt := range tests {
		out, err := run(t, tt.args...)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range tt.want {
			if !strings.Contains(out, w) {
				t.Errorf("%v: missing %q in %q", tt.args, w, out)
			}
		}
	}
}

func TestVersion(t *testing.T) {
	out, err := run(t, "--version")
	if err != nil || strings.TrimSpace(out) != "q30 version test" {
		t.Errorf("out = %q, err = %v", out, err)
	}
}
