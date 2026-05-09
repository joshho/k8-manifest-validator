//go:generate go run ../../tools/codegen

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodegenDeterministic(t *testing.T) {
	toolsDir := filepath.Join("..", "..")

	runCodegen := func() ([]byte, error) {
		cmd := exec.Command("/tmp/go/bin/go", "run", ".")
		cmd.Dir = filepath.Join(toolsDir, "tools", "codegen")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("codegen failed: %v\n%s", err, string(out))
		}
		genPath := filepath.Join(toolsDir, "pkg", "validator", "generated_structural.go")
		return os.ReadFile(genPath)
	}

	out1, err := runCodegen()
	if err != nil {
		t.Fatalf("first codegen run failed: %v", err)
	}

	out2, err := runCodegen()
	if err != nil {
		t.Fatalf("second codegen run failed: %v", err)
	}

	if string(out1) != string(out2) {
		t.Errorf("codegen output is non-deterministic")
	}
}

func TestCodegenOutputCompiles(t *testing.T) {
	cmd := exec.Command("/tmp/go/bin/go", "build", "-o", "/dev/null", "./pkg/validator")
	cmd.Dir = filepath.Join("..", "..")
	if err := cmd.Run(); err != nil {
		t.Fatalf("generated file does not compile: %v", err)
	}
}

func TestCodegenContainerFieldCoverage(t *testing.T) {
	toolsDir := filepath.Join("..", "..")
	genPath := filepath.Join(toolsDir, "pkg", "validator", "generated_structural.go")

	content, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	generated := string(content)

	testCases := []string{
		`"name"`,   // name field present
		`"image"`,  // image field present
		`"ports"`,  // ports field present
		`"resources"`, // resources field present
	}

	for _, tc := range testCases {
		if !strings.Contains(generated, tc) {
			t.Errorf("Container missing expected field path %q", tc)
		}
	}
}