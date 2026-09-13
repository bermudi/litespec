package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCLIValidateSuccessStatesStructuralScope(t *testing.T) {
	bin, root := setupCLITest(t)

	t.Run("text", func(t *testing.T) {
		out, code := runCLI(t, bin, root, "validate")
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if !strings.HasPrefix(out, "structure ok; implementation semantics not verified") {
			t.Fatalf("output does not state validation scope:\n%s", out)
		}
	})

	t.Run("minimal text", func(t *testing.T) {
		out, code := runCLI(t, bin, root, "validate", "--minimal")
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if !strings.HasPrefix(out, "structure-ok\tsemantics-unverified") {
			t.Fatalf("minimal output does not state validation scope:\n%s", out)
		}
	})
}

func TestCLIValidateStructuredSuccessStatesStructuralScope(t *testing.T) {
	bin, root := setupCLITest(t)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "full JSON", args: []string{"validate", "--json"}},
		{name: "minimal JSON", args: []string{"validate", "--minimal", "--json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runCLI(t, bin, root, tc.args...)
			if code != 0 {
				t.Fatalf("exit %d: %s", code, out)
			}

			var result struct {
				Valid                   bool   `json:"valid"`
				ValidationScope         string `json:"validationScope"`
				ImplementationSemantics string `json:"implementationSemantics"`
			}
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatalf("json: %v\n%s", err, out)
			}
			if !result.Valid {
				t.Error("expected structural valid=true")
			}
			if result.ValidationScope != "structure" {
				t.Errorf("validationScope = %q, want structure", result.ValidationScope)
			}
			if result.ImplementationSemantics != "unverified" {
				t.Errorf("implementationSemantics = %q, want unverified", result.ImplementationSemantics)
			}
		})
	}
}
